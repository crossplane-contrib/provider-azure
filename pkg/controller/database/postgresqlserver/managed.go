/*
Copyright 2019 The Crossplane Authors.

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

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/database/v1beta1"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/database"
)

// Error strings.
const (
	errNewClient                 = "cannot create new PostgreSQLServer client"
	errGetProvider               = "cannot get Azure provider"
	errGetProviderSecret         = "cannot get Azure provider Secret"
	errProviderSecretNil         = "Azure provider does not have a secret reference"
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

// Setup adds a controller that reconciles PostgreSQLInstances.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.PostgreSQLServerGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.PostgreSQLServer{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.PostgreSQLServerGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient(), newClientFn: newClient}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
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

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	v, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return nil, errors.New(errNotPostgreSQLServer)
	}

	p := &azurev1alpha3.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(v.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}
	sqlClient, err := c.newClientFn(s.Data[p.Spec.CredentialsSecretRef.Key])
	return &external{kube: c.client, client: sqlClient, newPasswordFn: password.Generate}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube          client.Client
	client        database.PostgreSQLServerAPI
	newPasswordFn func() (password string, err error)
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPostgreSQLServer)
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
			return managed.ExternalObservation{}, errors.Wrap(err, errCheckPostgreSQLServerName)
		}
		return managed.ExternalObservation{ResourceExists: creating}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPostgreSQLServer)
	}
	database.LateInitializePostgreSQL(&cr.Spec.ForProvider, server)
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpdateCR)
	}
	database.UpdatePostgreSQLObservation(&cr.Status.AtProvider, server)
	// We make this call after kube.Update since it doesn't update the
	// status subresource but fetches the the whole object after it's done. So,
	// changes to status has to be done after kube.Update in order not to get them
	// lost.
	if err := azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errFetchLastOperation)
	}
	switch server.UserVisibleState {
	case v1beta1.StateReady:
		cr.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	default:
		cr.SetConditions(runtimev1alpha1.Unavailable())
	}

	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: database.IsPostgreSQLUpToDate(cr.Spec.ForProvider, server), // NOTE(negz): We don't yet support updating Azure SQL servers.
		ConnectionDetails: managed.ConnectionDetails{
			runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(cr.Status.AtProvider.FullyQualifiedDomainName),
			runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(fmt.Sprintf("%s@%s", cr.Spec.ForProvider.AdministratorLogin, meta.GetExternalName(cr))),
		},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPostgreSQLServer)
	}

	cr.SetConditions(runtimev1alpha1.Creating())

	pw, err := e.newPasswordFn()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGenPassword)
	}
	if err := e.client.CreateServer(ctx, cr, pw); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePostgreSQLServer)
	}

	return managed.ExternalCreation{
			ConnectionDetails: managed.ConnectionDetails{
				runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(pw),
			},
		}, errors.Wrap(
			azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
			errFetchLastOperation)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPostgreSQLServer)
	}
	if cr.Status.AtProvider.LastOperation.Status == azure.AsyncOperationStatusInProgress {
		return managed.ExternalUpdate{}, nil
	}
	if err := e.client.UpdateServer(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePostgreSQLServer)
	}

	return managed.ExternalUpdate{}, errors.Wrap(
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
