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

package mysqlservervirtualnetworkrule

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql/mysqlapi"
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

	"github.com/crossplane-contrib/provider-azure/apis/database/v1alpha3"
	"github.com/crossplane-contrib/provider-azure/apis/v1alpha1"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
	"github.com/crossplane-contrib/provider-azure/pkg/clients/database"
	"github.com/crossplane-contrib/provider-azure/pkg/features"
)

// Error strings.
const (
	errNotMySQLServerVirtualNetworkRule    = "managed resource is not an MySQLServerVirtualNetworkRule"
	errCreateMySQLServerVirtualNetworkRule = "cannot create MySQLServerVirtualNetworkRule"
	errUpdateMySQLServerVirtualNetworkRule = "cannot update MySQLServerVirtualNetworkRule"
	errGetMySQLServerVirtualNetworkRule    = "cannot get MySQLServerVirtualNetworkRule"
	errDeleteMySQLServerVirtualNetworkRule = "cannot delete MySQLServerVirtualNetworkRule"
)

// Setup adds a controller that reconciles MySQLServerVirtualNetworkRules.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha3.MySQLServerVirtualNetworkRuleGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha3.MySQLServerVirtualNetworkRule{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.MySQLServerVirtualNetworkRuleGroupVersionKind),
			managed.WithConnectionPublishers(),
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

	cl := mysql.NewVirtualNetworkRulesClient(creds[azure.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client mysqlapi.VirtualNetworkRulesClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	v, ok := mg.(*v1alpha3.MySQLServerVirtualNetworkRule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMySQLServerVirtualNetworkRule)
	}

	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, v.Spec.ServerName, meta.GetExternalName(v))
	if azure.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetMySQLServerVirtualNetworkRule)
	}

	database.UpdateMySQLVirtualNetworkRuleStatusFromAzure(v, az)
	v.SetConditions(xpv1.Available())

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	v, ok := mg.(*v1alpha3.MySQLServerVirtualNetworkRule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMySQLServerVirtualNetworkRule)
	}

	v.SetConditions(xpv1.Creating())

	vnet := database.NewMySQLVirtualNetworkRuleParameters(v)
	if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, v.Spec.ServerName, meta.GetExternalName(v), vnet); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateMySQLServerVirtualNetworkRule)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	v, ok := mg.(*v1alpha3.MySQLServerVirtualNetworkRule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotMySQLServerVirtualNetworkRule)
	}

	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, v.Spec.ServerName, meta.GetExternalName(v))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetMySQLServerVirtualNetworkRule)
	}

	if database.MySQLServerVirtualNetworkRuleNeedsUpdate(v, az) {
		vnet := database.NewMySQLVirtualNetworkRuleParameters(v)
		if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, v.Spec.ServerName, meta.GetExternalName(v), vnet); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateMySQLServerVirtualNetworkRule)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	v, ok := mg.(*v1alpha3.MySQLServerVirtualNetworkRule)
	if !ok {
		return errors.New(errNotMySQLServerVirtualNetworkRule)
	}

	v.SetConditions(xpv1.Deleting())

	_, err := e.client.Delete(ctx, v.Spec.ResourceGroupName, v.Spec.ServerName, meta.GetExternalName(v))
	return errors.Wrap(resource.Ignore(azure.IsNotFound, err), errDeleteMySQLServerVirtualNetworkRule)
}
