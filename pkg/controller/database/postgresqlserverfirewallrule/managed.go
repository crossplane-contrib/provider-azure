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

package postgresqlserverfirewallrule

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql/postgresqlapi"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/database"
)

// Error strings.
const (
	errNotPostgreSQLServerFirewallRule    = "managed resource is not an PostgreSQLServerFirewallRule"
	errCreatePostgreSQLServerFirewallRule = "cannot create PostgreSQLServerFirewallRule"
	errUpdatePostgreSQLServerFirewallRule = "cannot update PostgreSQLServerFirewallRule"
	errGetPostgreSQLServerFirewallRule    = "cannot get PostgreSQLServerFirewallRule"
	errDeletePostgreSQLServerFirewallRule = "cannot delete PostgreSQLServerFirewallRule"
)

// Setup adds a controller that reconciles PostgreSQLServerFirewallRules.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.PostgreSQLServerFirewallRuleGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.PostgreSQLServerFirewallRule{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.PostgreSQLServerFirewallRuleGroupVersionKind),
			managed.WithConnectionPublishers(),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	sid, auth, err := azure.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := postgresql.NewFirewallRulesClient(sid)
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client postgresqlapi.FirewallRulesClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	v, ok := mg.(*v1alpha3.PostgreSQLServerFirewallRule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPostgreSQLServerFirewallRule)
	}

	az, err := e.client.Get(ctx, v.Spec.ForProvider.ResourceGroupName, v.Spec.ForProvider.ServerName, meta.GetExternalName(v))
	if azure.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPostgreSQLServerFirewallRule)
	}

	v.Status.AtProvider.ID = azure.ToString(az.ID)
	v.Status.AtProvider.Type = azure.ToString(az.Type)
	v.SetConditions(runtimev1alpha1.Available())

	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: database.PostgreSQLServerFirewallRuleIsUpToDate(v, az),
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	r, ok := mg.(*v1alpha3.PostgreSQLServerFirewallRule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPostgreSQLServerFirewallRule)
	}

	r.SetConditions(runtimev1alpha1.Creating())
	p := database.NewPostgreSQLFirewallRuleParameters(r)
	_, err := e.client.CreateOrUpdate(ctx, r.Spec.ForProvider.ResourceGroupName, r.Spec.ForProvider.ServerName, meta.GetExternalName(r), p)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreatePostgreSQLServerFirewallRule)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	r, ok := mg.(*v1alpha3.PostgreSQLServerFirewallRule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPostgreSQLServerFirewallRule)
	}

	p := database.NewPostgreSQLFirewallRuleParameters(r)
	_, err := e.client.CreateOrUpdate(ctx, r.Spec.ForProvider.ResourceGroupName, r.Spec.ForProvider.ServerName, meta.GetExternalName(r), p)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePostgreSQLServerFirewallRule)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	r, ok := mg.(*v1alpha3.PostgreSQLServerFirewallRule)
	if !ok {
		return errors.New(errNotPostgreSQLServerFirewallRule)
	}

	r.SetConditions(runtimev1alpha1.Deleting())
	_, err := e.client.Delete(ctx, r.Spec.ForProvider.ResourceGroupName, r.Spec.ForProvider.ServerName, meta.GetExternalName(r))
	return errors.Wrap(resource.Ignore(azure.IsNotFound, err), errDeletePostgreSQLServerFirewallRule)
}
