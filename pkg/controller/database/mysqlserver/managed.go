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

package mysqlserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/crossplaneio/stack-azure/pkg/clients/database"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/util"

	"github.com/crossplaneio/stack-azure/apis/database/v1beta1"
	azurev1alpha3 "github.com/crossplaneio/stack-azure/apis/v1alpha3"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
)

const passwordDataLen = 20

// Error strings.
const (
	errNewClient            = "cannot create new MySQLServer client"
	errGetProvider          = "cannot get Azure provider"
	errGetProviderSecret    = "cannot get Azure provider Secret"
	errUpdateCR             = "cannot update MySQLServer custom resource"
	errGenPassword          = "cannot generate admin password"
	errNotMySQLServer       = "managed resource is not a MySQLServer"
	errCreateMySQLServer    = "cannot create MySQLServer"
	errUpdateMySQLServer    = "cannot update MySQLServer"
	errGetMySQLServer       = "cannot get MySQLServer"
	errDeleteMySQLServer    = "cannot delete MySQLServer"
	errCheckMySQLServerName = "cannot check MySQLServer name availability"
)

// Controller is responsible for adding the MySQLServer controller and its
// corresponding reconciler to the manager with any runtime configuration.
type Controller struct{}

// SetupWithManager creates a new MySQLServer Controller and adds it to the
// Manager with default RBAC. The Manager will set fields on the Controller and
// start it when the Manager is Started.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1beta1.MySQLServerGroupVersionKind),
		resource.WithExternalConnecter(&connecter{client: mgr.GetClient(), newClientFn: newClient}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1beta1.MySQLServerKind, v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.MySQLServer{}).
		Complete(r)
}

func newClient(credentials []byte) (database.MySQLServerAPI, error) {
	ac, err := azure.NewClient(credentials)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Azure client")
	}
	mc, err := database.NewMySQLServerClient(ac)
	return mc, errors.Wrap(err, "cannot create Azure MySQL client")
}

type connecter struct {
	client      client.Client
	newClientFn func(credentials []byte) (database.MySQLServerAPI, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	v, ok := mg.(*v1beta1.MySQLServer)
	if !ok {
		return nil, errors.New(errNotMySQLServer)
	}

	p := &azurev1alpha3.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(v.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.Secret.Namespace, Name: p.Spec.Secret.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}
	sqlClient, err := c.newClientFn(s.Data[p.Spec.Secret.Key])
	return &external{kube: c.client, client: sqlClient, newPasswordFn: util.GeneratePassword}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube          client.Client
	client        database.MySQLServerAPI
	newPasswordFn func(len int) (password string, err error)
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.MySQLServer)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotMySQLServer)
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
			return resource.ExternalObservation{}, errors.Wrap(err, errCheckMySQLServerName)
		}
		return resource.ExternalObservation{ResourceExists: creating}, nil
	}
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, errGetMySQLServer)
	}
	database.LateInitializeMySQL(&cr.Spec.ForProvider, server)
	if err := e.kube.Update(ctx, cr); err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, errUpdateCR)
	}
	cr.Status.AtProvider = database.GenerateMySQLObservation(server)

	switch cr.Status.AtProvider.UserVisibleState {
	case database.StateReady:
		cr.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	default:
		cr.SetConditions(runtimev1alpha1.Unavailable())
	}

	return resource.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: database.IsMySQLUpToDate(cr.Spec.ForProvider, server),
		ConnectionDetails: resource.ConnectionDetails{
			runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(cr.Status.AtProvider.FullyQualifiedDomainName),
			runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(fmt.Sprintf("%s@%s", cr.Spec.ForProvider.AdministratorLogin, meta.GetExternalName(cr))),
		},
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.MySQLServer)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotMySQLServer)
	}

	cr.SetConditions(runtimev1alpha1.Creating())
	pw, err := e.newPasswordFn(passwordDataLen)
	if err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errGenPassword)
	}
	if err := e.client.CreateServer(ctx, cr, pw); err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errCreateMySQLServer)
	}

	return resource.ExternalCreation{
		ConnectionDetails: resource.ConnectionDetails{
			runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(pw),
		},
	}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.MySQLServer)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotMySQLServer)
	}
	// NOTE(muvaf): If an async update operation is ongoing, state is still ready
	// according to Azure but your update calls will be rejected since the resource
	// is `busy`. However, GET call returns the updated object even though it is
	// still being applied, so, we call Update only once.
	if cr.Status.AtProvider.UserVisibleState != database.StateReady {
		return resource.ExternalUpdate{}, nil
	}
	return resource.ExternalUpdate{}, errors.Wrap(e.client.UpdateServer(ctx, cr), errUpdateMySQLServer)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.MySQLServer)
	if !ok {
		return errors.New(errNotMySQLServer)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.UserVisibleState == database.StateDropping {
		return nil
	}
	return errors.Wrap(resource.Ignore(azure.IsNotFound, e.client.DeleteServer(ctx, cr)), errDeleteMySQLServer)
}
