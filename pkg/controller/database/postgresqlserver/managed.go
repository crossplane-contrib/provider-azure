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
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/postgresql/mgmt/postgresql"
	"github.com/negz/crossplane/pkg/util"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-azure/apis/database/v1alpha2"
	azurev1alpha2 "github.com/crossplaneio/stack-azure/apis/v1alpha2"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
)

const passwordDataLen = 20

// Error strings.
const (
	errNewClient                 = "cannot create new PostgresqlServer client"
	errGetProvider               = "cannot get Azure provider"
	errGetProviderSecret         = "cannot get Azure provider Secret"
	errGenPassword               = "cannot generate admin password"
	errNotPostgresqlServer       = "managed resource is not a PostgresqlServer"
	errCreatePostgresqlServer    = "cannot create PostgresqlServer"
	errGetPostgresqlServer       = "cannot get PostgresqlServer"
	errDeletePostgresqlServer    = "cannot delete PostgresqlServer"
	errCheckPostgresqlServerName = "cannot check PostgresqlServer name availability"
)

// Controller is responsible for adding the PostgresqlServer controller and its
// corresponding reconciler to the manager with any runtime configuration.
type Controller struct{}

// SetupWithManager creates a new PostgresqlServer Controller and adds it to the
// Manager with default RBAC. The Manager will set fields on the Controller and
// start it when the Manager is Started.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1alpha2.PostgresqlServerGroupVersionKind),
		resource.WithExternalConnecter(&connecter{client: mgr.GetClient(), newClientFn: newClient}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1alpha2.PostgresqlServerKind, v1alpha2.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha2.PostgresqlServer{}).
		Complete(r)
}

func newClient(credentials []byte) (azure.PostgreSQLServerAPI, error) {
	ac, err := azure.NewClient(credentials)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create Azure client")
	}
	pc, err := azure.NewPostgreSQLServerClient(ac)
	return pc, errors.Wrap(err, "cannot create Azure MySQL client")
}

type connecter struct {
	client      client.Client
	newClientFn func(credentials []byte) (azure.PostgreSQLServerAPI, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	v, ok := mg.(*v1alpha2.PostgresqlServer)
	if !ok {
		return nil, errors.New(errNotPostgresqlServer)
	}

	p := &azurev1alpha2.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(v.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.Secret.Namespace, Name: p.Spec.Secret.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}
	client, err := c.newClientFn(s.Data[p.Spec.Secret.Key])
	return &external{client: client, newPasswordFn: util.GeneratePassword}, errors.Wrap(err, errNewClient)
}

type external struct {
	client        azure.PostgreSQLServerAPI
	newPasswordFn func(len int) (password string, err error)
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	s, ok := mg.(*v1alpha2.PostgresqlServer)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotPostgresqlServer)
	}

	external, err := e.client.GetServer(ctx, s)
	if azure.IsNotFound(err) {
		// Azure SQL servers don't exist according to the Azure API until their
		// create operation has completed, and Azure will happily let you submit
		// several subsequent create operations for the same server. Our create
		// call is not idempotent because it creates a new random password each
		// time, so we want to ensure it's only called once. Fortunately Azure
		// exposes an API that reports server names to be taken as soon as their
		// create operation is accepted.
		creating, err := e.client.ServerNameTaken(ctx, s)
		if err != nil {
			return resource.ExternalObservation{}, errors.Wrap(err, errCheckPostgresqlServerName)
		}
		if creating {
			return resource.ExternalObservation{
				ResourceExists:   true,
				ResourceUpToDate: true, // NOTE(negz): We don't yet support updating Azure SQL servers.
			}, nil
		}

		return resource.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, errGetPostgresqlServer)
	}

	s.Status.State = string(external.UserVisibleState)
	s.Status.ProviderID = azure.ToString(external.ID)
	s.Status.Endpoint = azure.ToString(external.FullyQualifiedDomainName)

	switch external.UserVisibleState {
	case postgresql.ServerStateReady:
		s.SetConditions(runtimev1alpha1.Available())
		if s.Status.Endpoint != "" {
			resource.SetBindable(s)
		}
	default:
		s.SetConditions(runtimev1alpha1.Unavailable())
	}

	o := resource.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true, // NOTE(negz): We don't yet support updating Azure SQL servers.
		ConnectionDetails: resource.ConnectionDetails{
			runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(s.Status.Endpoint),
			runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(fmt.Sprintf("%s@%s", s.Spec.AdminLoginName, s.GetName())),
		},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	s, ok := mg.(*v1alpha2.PostgresqlServer)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotPostgresqlServer)
	}

	s.SetConditions(runtimev1alpha1.Creating())

	pw, err := e.newPasswordFn(passwordDataLen)
	if err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errGenPassword)
	}

	if err := e.client.CreateServer(ctx, s, pw); err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errCreatePostgresqlServer)
	}

	ec := resource.ExternalCreation{
		ConnectionDetails: resource.ConnectionDetails{
			runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(pw),
		},
	}

	return ec, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	// TODO(negz): Support updating Azure SQL servers. :)
	return resource.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	s, ok := mg.(*v1alpha2.PostgresqlServer)
	if !ok {
		return errors.New(errNotPostgresqlServer)
	}

	s.SetConditions(runtimev1alpha1.Deleting())
	return errors.Wrap(resource.Ignore(azure.IsNotFound, e.client.DeleteServer(ctx, s)), errDeletePostgresqlServer)
}
