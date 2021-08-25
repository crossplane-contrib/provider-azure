/*
Copyright 2021 The Crossplane Authors.

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

package postgresqlserverconfiguration

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/database/v1beta1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/database/configuration"
)

const (
	// error messages
	errUpdateCR                     = "cannot update PostgreSQLServerConfiguration custom resource"
	errNotPostgreSQLServerConfig    = "managed resource is not a PostgreSQLServerConfiguration"
	errCreatePostgreSQLServerConfig = "cannot create PostgreSQLServerConfiguration"
	errUpdatePostgreSQLServerConfig = "cannot update PostgreSQLServerConfiguration"
	errGetPostgreSQLServerConfig    = "cannot get PostgreSQLServerConfiguration"
	errDeletePostgreSQLServerConfig = "cannot delete PostgreSQLServerConfiguration"
	errFetchLastOperation           = "cannot fetch last operation"

	fmtExternalName = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DBforPostgreSQL/servers/%s/configurations/%s"
)

// Setup adds a controller that reconciles PostgreSQLInstances.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1beta1.PostgreSQLServerConfigurationGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1beta1.PostgreSQLServerConfiguration{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.PostgreSQLServerConfigurationGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azure.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := postgresql.NewConfigurationsClient(creds[azure.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{
		kube:           c.client,
		client:         configuration.NewPostgreSQLConfigurationClient(cl),
		subscriptionID: creds[azure.CredentialsKeySubscriptionID],
	}, nil
}

type external struct {
	kube           client.Client
	client         configuration.PostgreSQLConfigurationAPI
	subscriptionID string
}

func (e external) getExternalName(resourceGroupName, serverName, configName string) string {
	return fmt.Sprintf(fmtExternalName, e.subscriptionID, resourceGroupName, serverName, configName)
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	// cyclomatic complexity of this method (13) is slightly higher than our goal of 10.
	cr, ok := mg.(*v1beta1.PostgreSQLServerConfiguration)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPostgreSQLServerConfig)
	}
	config, err := e.client.Get(ctx, cr)
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
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPostgreSQLServerConfig)
	}
	// ARM does not return a 404 for the configuration resource even if we set its value to the server default
	// and source to "system-default". Hence, we check those conditions here:
	if meta.WasDeleted(cr) && cr.Status.AtProvider.Source == configuration.SourceSystemManaged && cr.Status.AtProvider.Value == cr.Status.AtProvider.DefaultValue {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	// it's possible that external.Create has never been called, thus set ext. name if not set
	if meta.GetExternalName(cr) == "" {
		meta.SetExternalName(cr, e.getExternalName(cr.Spec.ForProvider.ResourceGroupName,
			cr.Spec.ForProvider.ServerName, cr.Spec.ForProvider.Name))
	}

	configuration.LateInitializePostgreSQLConfiguration(&cr.Spec.ForProvider, config)
	if err := e.kube.Update(ctx, cr); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpdateCR)
	}
	configuration.UpdatePostgreSQLConfigurationObservation(&cr.Status.AtProvider, config)
	// We make this call after kube.Update since it doesn't update the
	// status subresource but fetches the whole object after it's done. So,
	// changes to status has to be done after kube.Update in order not to get them
	// lost.
	if err := azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errFetchLastOperation)
	}
	// if the configuration has been applied successfully, then mark MR as available
	if cr.Status.AtProvider.Value == azure.ToString(cr.Spec.ForProvider.Value) {
		cr.SetConditions(xpv1.Available())
	} else {
		cr.SetConditions(xpv1.Unavailable())
	}

	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: configuration.IsPostgreSQLConfigurationUpToDate(cr.Spec.ForProvider, config),
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.PostgreSQLServerConfiguration)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPostgreSQLServerConfig)
	}

	if err := e.client.CreateOrUpdate(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePostgreSQLServerConfig)
	}
	// no error if ext name does not match
	meta.SetExternalName(cr, e.getExternalName(cr.Spec.ForProvider.ResourceGroupName,
		cr.Spec.ForProvider.ServerName, cr.Spec.ForProvider.Name))

	return managed.ExternalCreation{
			ExternalNameAssigned: true,
		}, errors.Wrap(
			azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
			errFetchLastOperation)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.PostgreSQLServerConfiguration)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPostgreSQLServerConfig)
	}
	if cr.Status.AtProvider.LastOperation.Status == azure.AsyncOperationStatusInProgress {
		return managed.ExternalUpdate{}, nil
	}
	if err := e.client.CreateOrUpdate(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePostgreSQLServerConfig)
	}

	return managed.ExternalUpdate{}, errors.Wrap(
		azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
		errFetchLastOperation)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.PostgreSQLServerConfiguration)
	if !ok {
		return errors.New(errNotPostgreSQLServerConfig)
	}

	if err := e.client.Delete(ctx, cr); resource.Ignore(azure.IsNotFound, err) != nil {
		return errors.Wrap(err, errDeletePostgreSQLServerConfig)
	}
	return errors.Wrap(
		azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
		errFetchLastOperation)
}
