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

package mysqlserverconfiguration

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
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	azure "github.com/crossplane/provider-azure/internal/pkg/clients"
	"github.com/crossplane/provider-azure/internal/pkg/clients/database/configuration"
	"github.com/crossplane/provider-azure/internal/pkg/features"

	"github.com/crossplane/provider-azure/apis/classic/database/v1beta1"
	"github.com/crossplane/provider-azure/apis/classic/v1alpha1"
)

const (
	// error messages
	errNotMySQLServerConfig      = "managed resource is not a MySQLServerConfiguration"
	errCreateMySQLServerConfig   = "cannot create MySQLServerConfiguration"
	errUpdateMySQLServerConfig   = "cannot update MySQLServerConfiguration"
	errGetMySQLServerConfig      = "cannot get MySQLServerConfiguration"
	errDeleteMySQLServerConfig   = "cannot delete MySQLServerConfiguration"
	errFetchLastOperation        = "cannot fetch last operation"
	errNotFoundMySQLServerConfig = "the specified MySQLServerConfiguration does not exist"

	fmtExternalName = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.DBforMySQL/servers/%s/configurations/%s"
)

// Setup adds a controller that reconciles MySQLInstances.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.MySQLServerConfigurationGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1beta1.MySQLServerConfiguration{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.MySQLServerConfigurationGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithInitializers(managed.NewDefaultProviderConfig(mgr.GetClient())),
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
	cl := mysql.NewConfigurationsClient(creds[azure.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{
		kube:           c.client,
		client:         configuration.NewMySQLConfigurationClient(cl),
		subscriptionID: creds[azure.CredentialsKeySubscriptionID],
	}, nil
}

type external struct {
	kube           client.Client
	client         configuration.MySQLConfigurationAPI
	subscriptionID string
}

func (e external) generateExtName(resourceGroupName, serverName, configName string) string {
	return fmt.Sprintf(fmtExternalName, e.subscriptionID, resourceGroupName, serverName, configName)
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) { // nolint:gocyclo
	// cyclomatic complexity of this method (13) is slightly higher than our goal of 10.
	cr, ok := mg.(*v1beta1.MySQLServerConfiguration)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMySQLServerConfig)
	}
	config, err := e.client.Get(ctx, cr)
	if azure.IsNotFound(err) {
		// Valid configurations are pre-determined in server side and new ones cannot be created.
		// Only existing valid configurations can be updated.
		// Therefore, if the config cannot be found in the result of the get call, instead of returning nil, an error
		// is returned and the error is reported in status conditions of the MySQLServerConfiguration managed resource.
		return managed.ExternalObservation{}, errors.Wrap(err, errNotFoundMySQLServerConfig)
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetMySQLServerConfig)
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
		meta.SetExternalName(cr, e.generateExtName(cr.Spec.ForProvider.ResourceGroupName,
			cr.Spec.ForProvider.ServerName, cr.Spec.ForProvider.Name))
	}

	l := resource.NewLateInitializer()
	cr.Spec.ForProvider.Value = l.LateInitializeStringPtr(cr.Spec.ForProvider.Value, config.Value)

	configuration.UpdateMySQLConfigurationObservation(&cr.Status.AtProvider, config)
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

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        configuration.IsMySQLConfigurationUpToDate(cr.Spec.ForProvider, config),
		ResourceLateInitialized: l.IsChanged(),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.MySQLServerConfiguration)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMySQLServerConfig)
	}

	if err := e.client.CreateOrUpdate(ctx, cr); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateMySQLServerConfig)
	}
	// no error if ext name does not match
	meta.SetExternalName(cr, e.generateExtName(cr.Spec.ForProvider.ResourceGroupName,
		cr.Spec.ForProvider.ServerName, cr.Spec.ForProvider.Name))

	return managed.ExternalCreation{
			ExternalNameAssigned: true,
		}, errors.Wrap(
			azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
			errFetchLastOperation)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.MySQLServerConfiguration)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotMySQLServerConfig)
	}
	if cr.Status.AtProvider.LastOperation.Status == azure.AsyncOperationStatusInProgress {
		return managed.ExternalUpdate{}, nil
	}
	if err := e.client.CreateOrUpdate(ctx, cr); err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateMySQLServerConfig)
	}

	return managed.ExternalUpdate{}, errors.Wrap(
		azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
		errFetchLastOperation)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.MySQLServerConfiguration)
	if !ok {
		return errors.New(errNotMySQLServerConfig)
	}

	if err := e.client.Delete(ctx, cr); resource.Ignore(azure.IsNotFound, err) != nil {
		return errors.Wrap(err, errDeleteMySQLServerConfig)
	}
	return errors.Wrap(
		azure.FetchAsyncOperation(ctx, e.client.GetRESTClient(), &cr.Status.AtProvider.LastOperation),
		errFetchLastOperation)
}
