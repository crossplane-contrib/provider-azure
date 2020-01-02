/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package postgresqlserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/negz/crossplane/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-azure/apis/database/v1beta1"
	azurev1alpha3 "github.com/crossplaneio/stack-azure/apis/v1alpha3"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
	"github.com/crossplaneio/stack-azure/pkg/clients/database"
)

const passwordDataLen = 20

// Error strings.
const (
	errNewClient                 = "cannot create new PostgreSQLServer client"
	errGetProvider               = "cannot get Azure provider"
	errGetProviderSecret         = "cannot get Azure provider Secret"
	errUpdateCR                  = "cannot update PostgreSQL custom resource"
	errGenPassword               = "cannot generate admin password"
	errNotPostgreSQLServer       = "managed resource is not a PostgreSQLServer"
	errCreatePostgreSQLServer    = "cannot create PostgreSQLServer"
	errUpdatePostgreSQLServer    = "cannot update PostgreSQLServer"
	errGetPostgreSQLServer       = "cannot get PostgreSQLServer"
	errDeletePostgreSQLServer    = "cannot delete PostgreSQLServer"
	errCheckPostgreSQLServerName = "cannot check PostgreSQLServer name availability"
	errFetchLastOperation        = "cannot fetch last operation"
)

// Controller is responsible for adding the PostgreSQLServer controller and its
// corresponding reconciler to the manager with any runtime configuration.
type Controller struct{}

// SetupWithManager creates a new PostgreSQLServer Controller and adds it to the
// Manager with default RBAC. The Manager will set fields on the Controller and
// start it when the Manager is Started.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1beta1.PostgreSQLServerGroupVersionKind),
		resource.WithExternalConnecter(&connecter{client: mgr.GetClient(), newClientFn: newClient}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1beta1.PostgreSQLServerKind, v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.PostgreSQLServer{}).
		Complete(r)
}

func newClient(credentials []byte) (database.PostgreSQLServerAPI, error) {
	ac, err := azure.NewClient(credentials)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Azure client")
	}
	pc, err := database.NewPostgreSQLServerClient(ac)
	return pc, errors.Wrap(err, "cannot create Azure MySQL client")
}

type connecter struct {
	client      client.Client
	newClientFn func(credentials []byte) (database.PostgreSQLServerAPI, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	v, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return nil, errors.New(errNotPostgreSQLServer)
	}

	p := &azurev1alpha3.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(v.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}
	sqlClient, err := c.newClientFn(s.Data[p.Spec.CredentialsSecretRef.Key])
	return &external{kube: c.client, client: sqlClient, newPasswordFn: util.GeneratePassword}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube          client.Client
	client        database.PostgreSQLServerAPI
	newPasswordFn func(len int) (password string, err error)
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotPostgreSQLServer)
	}
	server, err := e.client.GetServer(ctx, cr)
	if azure.IsNotFound(err) {
		// Azure SQL servers don't exist according to the Azure API until their
		// create operation has completed, and Azure will happily let you submit
		// several subsequent create operations for the same server. Our create
		// call is not idempotent because it creates a new random password each
		// time, so we want to ensure it's only called once. Fortunately Azure
		// exposes an API that reports server names to be taken as soon as their
		// create operation is accepted.
		creating, err := e.client.ServerNameTaken(ctx, cr)
		if err != nil {
			return resource.ExternalObservation{}, errors.Wrap(err, errCheckPostgreSQLServerName)
		}
		return resource.ExternalObservation{ResourceExists: creating}, nil
	}
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, errGetPostgreSQLServer)
	}
	database.LateInitializePostgreSQL(&cr.Spec.ForProvider, server)
	if err := e.kube.Update(ctx, cr); err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, errUpdateCR)
	}
	database.UpdatePostgreSQLObservation(&cr.Status.AtProvider, server)
	// We make this call after kube.Update since it doesn't update the
	// status subresource but fetches the the whole object after it's done. So,
	// changes to status has to be done after kube.Update in order not to get them
	// lost.
	if err := azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation); err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, errFetchLastOperation)
	}
	switch server.UserVisibleState {
	case v1beta1.StateReady:
		cr.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	default:
		cr.SetConditions(runtimev1alpha1.Unavailable())
	}

	o := resource.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: database.IsPostgreSQLUpToDate(cr.Spec.ForProvider, server), // NOTE(negz): We don't yet support updating Azure SQL servers.
		ConnectionDetails: resource.ConnectionDetails{
			runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(cr.Status.AtProvider.FullyQualifiedDomainName),
			runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(fmt.Sprintf("%s@%s", cr.Spec.ForProvider.AdministratorLogin, meta.GetExternalName(cr))),
		},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotPostgreSQLServer)
	}

	cr.SetConditions(runtimev1alpha1.Creating())

	pw, err := e.newPasswordFn(passwordDataLen)
	if err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errGenPassword)
	}
	if err := e.client.CreateServer(ctx, cr, pw); err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errCreatePostgreSQLServer)
	}

	return resource.ExternalCreation{
			ConnectionDetails: resource.ConnectionDetails{
				runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(pw),
			},
		}, errors.Wrap(
			azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
			errFetchLastOperation)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotPostgreSQLServer)
	}
	if cr.Status.AtProvider.LastOperation.Status == azure.AsyncOperationStatusInProgress {
		return resource.ExternalUpdate{}, nil
	}
	if err := e.client.UpdateServer(ctx, cr); err != nil {
		return resource.ExternalUpdate{}, errors.Wrap(err, errUpdatePostgreSQLServer)
	}

	return resource.ExternalUpdate{}, errors.Wrap(
		azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
		errFetchLastOperation)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return errors.New(errNotPostgreSQLServer)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.UserVisibleState == v1beta1.StateDropping {
		return nil
	}
	if err := e.client.DeleteServer(ctx, cr); resource.Ignore(azure.IsNotFound, err) != nil {
		return errors.Wrap(err, errDeletePostgreSQLServer)
	}
	return errors.Wrap(
		azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
		errFetchLastOperation)
}
