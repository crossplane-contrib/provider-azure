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

package postgresqlservervirtualnetworkrule

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql/postgresqlapi"
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

	azure "github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients"
	"github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients/database"
	"github.com/crossplane-contrib/provider-jet-azure/internal/pkg/features"

	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/database/v1alpha3"
	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/v1alpha1"
)

// Error strings.
const (
	errNotPostgreSQLServerVirtualNetworkRule    = "managed resource is not an PostgreSQLServerVirtualNetworkRule"
	errCreatePostgreSQLServerVirtualNetworkRule = "cannot create PostgreSQLServerVirtualNetworkRule"
	errUpdatePostgreSQLServerVirtualNetworkRule = "cannot update PostgreSQLServerVirtualNetworkRule"
	errGetPostgreSQLServerVirtualNetworkRule    = "cannot get PostgreSQLServerVirtualNetworkRule"
	errDeletePostgreSQLServerVirtualNetworkRule = "cannot delete PostgreSQLServerVirtualNetworkRule"
)

// Setup adds a controller that reconciles PostgreSQLServerVirtualNetworkRules.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha3.PostgreSQLServerVirtualNetworkRuleGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha3.PostgreSQLServerVirtualNetworkRule{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.PostgreSQLServerVirtualNetworkRuleGroupVersionKind),
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

	cl := postgresql.NewVirtualNetworkRulesClient(creds[azure.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client postgresqlapi.VirtualNetworkRulesClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	v, ok := mg.(*v1alpha3.PostgreSQLServerVirtualNetworkRule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPostgreSQLServerVirtualNetworkRule)
	}

	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, v.Spec.ServerName, meta.GetExternalName(v))
	if azure.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPostgreSQLServerVirtualNetworkRule)
	}

	database.UpdatePostgreSQLVirtualNetworkRuleStatusFromAzure(v, az)

	v.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
		ResourceUpToDate:  !database.PostgreSQLServerVirtualNetworkRuleNeedsUpdate(v, az),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	v, ok := mg.(*v1alpha3.PostgreSQLServerVirtualNetworkRule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPostgreSQLServerVirtualNetworkRule)
	}

	v.SetConditions(xpv1.Creating())

	vnet := database.NewPostgreSQLVirtualNetworkRuleParameters(v)
	_, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, v.Spec.ServerName, meta.GetExternalName(v), vnet)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreatePostgreSQLServerVirtualNetworkRule)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	v, ok := mg.(*v1alpha3.PostgreSQLServerVirtualNetworkRule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPostgreSQLServerVirtualNetworkRule)
	}

	vnet := database.NewPostgreSQLVirtualNetworkRuleParameters(v)
	_, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, v.Spec.ServerName, meta.GetExternalName(v), vnet)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePostgreSQLServerVirtualNetworkRule)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	v, ok := mg.(*v1alpha3.PostgreSQLServerVirtualNetworkRule)
	if !ok {
		return errors.New(errNotPostgreSQLServerVirtualNetworkRule)
	}

	v.SetConditions(xpv1.Deleting())

	_, err := e.client.Delete(ctx, v.Spec.ResourceGroupName, v.Spec.ServerName, meta.GetExternalName(v))

	return errors.Wrap(resource.Ignore(azure.IsNotFound, err), errDeletePostgreSQLServerVirtualNetworkRule)
}
