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

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-azure/apis/database/v1beta1"
	"github.com/crossplane-contrib/provider-azure/apis/v1alpha1"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
	"github.com/crossplane-contrib/provider-azure/pkg/clients/database"
	"github.com/crossplane-contrib/provider-azure/pkg/features"
)

// Error strings.
const (
	errUpdateCR           = "cannot update MySQLServer custom resource"
	errGenPassword        = "cannot generate admin password"
	errNotMySQLServer     = "managed resource is not a MySQLServer"
	errCreateMySQLServer  = "cannot create MySQLServer"
	errUpdateMySQLServer  = "cannot update MySQLServer"
	errGetMySQLServer     = "cannot get MySQLServer"
	errDeleteMySQLServer  = "cannot delete MySQLServer"
	errFetchLastOperation = "cannot fetch last operation"
)

// Setup adds a controller that reconciles MySQLServers.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.MySQLServerGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.MySQLServer{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.MySQLServerGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azure.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := mysql.NewServersClient(creds[azure.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{kube: c.client, client: database.NewMySQLServerClient(cl), newPasswordFn: password.Generate}, nil
}

type external struct {
	kube          client.Client
	client        database.MySQLServerAPI
	newPasswordFn func() (password string, err error)
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.MySQLServer)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMySQLServer)
	}

	server, err := e.client.GetServer(ctx, cr)
	if azure.IsNotFound(err) {
		if err := azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errFetchLastOperation)
		}
		// Azure returns NotFound for GET calls until creation is completed
		// successfully and we cannot return `ResourceExists: false` during creation
		// since this will cause `Create` to be called again and it's not idempotent.
		// So, we check whether a creation operation in fact is in motion.
		creating := cr.Status.AtProvider.LastOperation.Method == "PUT" &&
			cr.Status.AtProvider.LastOperation.Status == azure.AsyncOperationStatusInProgress
		return managed.ExternalObservation{ResourceExists: creating}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetMySQLServer)
	}
	database.LateInitializeMySQL(&cr.Spec.ForProvider, server)
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpdateCR)
	}
	database.UpdateMySQLObservation(&cr.Status.AtProvider, server)
	// We make this call after kube.Update since it doesn't update the
	// status subresource but fetches the the whole object after it's done. So,
	// changes to status has to be done after kube.Update in order not to get them
	// lost.
	if err := azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errFetchLastOperation)
	}
	switch cr.Status.AtProvider.UserVisibleState {
	case v1beta1.StateReady:
		cr.SetConditions(xpv1.Available())
	default:
		cr.SetConditions(xpv1.Unavailable())
	}

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: database.IsMySQLUpToDate(cr.Spec.ForProvider, server),
		ConnectionDetails: managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretEndpointKey: []byte(cr.Status.AtProvider.FullyQualifiedDomainName),
			xpv1.ResourceCredentialsSecretUserKey:     []byte(fmt.Sprintf("%s@%s", cr.Spec.ForProvider.AdministratorLogin, meta.GetExternalName(cr))),
		},
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.MySQLServer)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMySQLServer)
	}

	cr.SetConditions(xpv1.Creating())
	pw, err := e.newPasswordFn()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGenPassword)
	}
	if err := e.client.CreateServer(ctx, cr, pw); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateMySQLServer)
	}

	return managed.ExternalCreation{
			ConnectionDetails: managed.ConnectionDetails{
				xpv1.ResourceCredentialsSecretPasswordKey: []byte(pw),
			},
		}, errors.Wrap(
			azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
			errFetchLastOperation)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.MySQLServer)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotMySQLServer)
	}
	if cr.Status.AtProvider.LastOperation.Status == azure.AsyncOperationStatusInProgress {
		return managed.ExternalUpdate{}, nil
	}
	if err := e.client.UpdateServer(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateMySQLServer)
	}

	return managed.ExternalUpdate{}, errors.Wrap(
		azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
		errFetchLastOperation)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.MySQLServer)
	if !ok {
		return errors.New(errNotMySQLServer)
	}
	cr.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.UserVisibleState == v1beta1.StateDropping {
		return nil
	}
	if err := e.client.DeleteServer(ctx, cr); resource.Ignore(azure.IsNotFound, err) != nil {
		return errors.Wrap(err, errDeleteMySQLServer)
	}

	return errors.Wrap(
		azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
		errFetchLastOperation)
}
