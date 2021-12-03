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
	"time"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/database/v1beta1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/database"
)

// Error strings.
const (
	errGenPassword            = "cannot generate admin password"
	errNotPostgreSQLServer    = "managed resource is not a PostgreSQLServer"
	errCreatePostgreSQLServer = "cannot create PostgreSQLServer"
	errUpdatePostgreSQLServer = "cannot update PostgreSQLServer"
	errGetPostgreSQLServer    = "cannot get PostgreSQLServer"
	errDeletePostgreSQLServer = "cannot delete PostgreSQLServer"
	errGetConnSecret          = "cannot get connection secret"
)

// Setup adds a controller that reconciles PostgreSQLInstances.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.PostgreSQLServerGroupKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.PostgreSQLServerGroupVersionKind),
		managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithPollInterval(o.PollInterval),
		managed.WithCreationGracePeriod(5*time.Minute), // TODO(negz): Tune me.
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.PostgreSQLServer{}).
		Complete(ratelimiter.NewReconciler(name, r, o.GlobalRateLimiter))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azure.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := postgresql.NewServersClient(creds[azure.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{kube: c.client, client: database.NewPostgreSQLServerClient(cl), newPasswordFn: password.Generate}, nil
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
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPostgreSQLServer)
	}

	// NOTE(negz): We make a best effort attempt to fetch the async op for
	// backward compatibility. We used to use this operation to determine
	// whether we had requested the database be created and whether the
	// creation was successful. Now we instead use a creation grace period.
	_ = azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation)

	database.UpdatePostgreSQLObservation(&cr.Status.AtProvider, server)

	// Any state beside 'ready' is considered unavailable.
	switch server.UserVisibleState { //nolint:exhaustive
	case v1beta1.StateReady:
		cr.SetConditions(xpv1.Available())
	default:
		cr.SetConditions(xpv1.Unavailable())
	}

	o := managed.ExternalObservation{
		ResourceExists:          true,
		ResourceLateInitialized: database.LateInitializePostgreSQL(&cr.Spec.ForProvider, server),
		ResourceUpToDate:        database.IsPostgreSQLUpToDate(cr.Spec.ForProvider, server),
		ConnectionDetails: managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretEndpointKey: []byte(cr.Status.AtProvider.FullyQualifiedDomainName),
			xpv1.ResourceCredentialsSecretUserKey:     []byte(fmt.Sprintf("%s@%s", cr.Spec.ForProvider.AdministratorLogin, meta.GetExternalName(cr))),
			xpv1.ResourceCredentialsSecretPortKey:     []byte(v1beta1.PostgreSQLServerPort),
		},
	}

	return o, nil
}

func (e *external) getPassword(ctx context.Context, cr *v1beta1.PostgreSQLServer) (string, error) {
	if cr.Spec.WriteConnectionSecretToReference == nil ||
		cr.Spec.WriteConnectionSecretToReference.Name == "" || cr.Spec.WriteConnectionSecretToReference.Namespace == "" {
		return "", nil
	}

	s := &v1.Secret{}
	if err := e.kube.Get(ctx, types.NamespacedName{
		Namespace: cr.Spec.WriteConnectionSecretToReference.Namespace,
		Name:      cr.Spec.WriteConnectionSecretToReference.Name,
	}, s); err != nil {
		return "", errors.Wrap(err, errGetConnSecret)
	}

	return string(s.Data[xpv1.ResourceCredentialsSecretPasswordKey]), nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPostgreSQLServer)
	}

	cr.SetConditions(xpv1.Creating())

	pw, err := e.getPassword(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, err
	}
	if pw == "" {
		pw, err = e.newPasswordFn()
		if err != nil {
			return managed.ExternalCreation{}, errors.Wrap(err, errGenPassword)
		}
	}
	if err := e.client.CreateServer(ctx, cr, pw); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePostgreSQLServer)
	}

	return managed.ExternalCreation{ConnectionDetails: managed.ConnectionDetails{xpv1.ResourceCredentialsSecretPasswordKey: []byte(pw)}}, nil
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

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.PostgreSQLServer)
	if !ok {
		return errors.New(errNotPostgreSQLServer)
	}
	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.UserVisibleState == v1beta1.StateDropping {
		return nil
	}
	if err := e.client.DeleteServer(ctx, cr); resource.Ignore(azure.IsNotFound, err) != nil {
		return errors.Wrap(err, errDeletePostgreSQLServer)
	}
	return nil
}
