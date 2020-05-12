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

package mysqlserverfirewallrule

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/database"
)

// Error strings.
const (
	errNewClient                     = "cannot create new MySQLServerFirewallRule"
	errProviderSecretNil             = "provider does not have a secret reference"
	errNotMySQLServerFirewallRule    = "managed resource is not an MySQLServerFirewallRule"
	errCreateMySQLServerFirewallRule = "cannot create MySQLServerFirewallRule"
	errUpdateMySQLServerFirewallRule = "cannot update MySQLServerFirewallRule"
	errGetMySQLServerFirewallRule    = "cannot get MySQLServerFirewallRule"
	errDeleteMySQLServerFirewallRule = "cannot delete MySQLServerFirewallRule"
)

// Setup adds a controller that reconciles MySQLServerFirewallRules.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.MySQLServerFirewallRuleGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.MySQLServerFirewallRule{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.MySQLServerFirewallRuleGroupVersionKind),
			managed.WithConnectionPublishers(),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient(), newClientFn: database.NewMySQLFirewallRulesClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client      client.Client
	newClientFn func(ctx context.Context, credentials []byte) (database.MySQLFirewallRulesClient, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	v, ok := mg.(*v1alpha3.MySQLServerFirewallRule)
	if !ok {
		return nil, errors.New(errNotMySQLServerFirewallRule)
	}

	p := &azurev1alpha3.Provider{}
	n := meta.NamespacedNameOf(v.Spec.ProviderReference)
	if err := c.client.Get(ctx, n, p); err != nil {
		return nil, errors.Wrapf(err, "cannot get provider %s", v.Spec.ProviderReference.Name)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	s := &corev1.Secret{}
	n = types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrapf(err, "cannot get provider secret %s", n)
	}
	client, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key])
	return &external{client: client}, errors.Wrap(err, errNewClient)
}

type external struct {
	client database.MySQLFirewallRulesClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	v, ok := mg.(*v1alpha3.MySQLServerFirewallRule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMySQLServerFirewallRule)
	}

	az, err := e.client.Get(ctx, v.Spec.ForProvider.ResourceGroupName, v.Spec.ForProvider.ServerName, meta.GetExternalName(v))
	if azure.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetMySQLServerFirewallRule)
	}

	v.Status.AtProvider.ID = azure.ToString(az.ID)
	v.Status.AtProvider.Type = azure.ToString(az.Type)
	v.SetConditions(runtimev1alpha1.Available())

	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: database.MySQLServerFirewallRuleIsUpToDate(v, az),
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	r, ok := mg.(*v1alpha3.MySQLServerFirewallRule)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMySQLServerFirewallRule)
	}

	r.SetConditions(runtimev1alpha1.Creating())
	p := database.NewMySQLFirewallRuleParameters(r)
	_, err := e.client.CreateOrUpdate(ctx, r.Spec.ForProvider.ResourceGroupName, r.Spec.ForProvider.ServerName, meta.GetExternalName(r), p)
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateMySQLServerFirewallRule)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	r, ok := mg.(*v1alpha3.MySQLServerFirewallRule)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotMySQLServerFirewallRule)
	}

	p := database.NewMySQLFirewallRuleParameters(r)
	_, err := e.client.CreateOrUpdate(ctx, r.Spec.ForProvider.ResourceGroupName, r.Spec.ForProvider.ServerName, meta.GetExternalName(r), p)
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateMySQLServerFirewallRule)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	r, ok := mg.(*v1alpha3.MySQLServerFirewallRule)
	if !ok {
		return errors.New(errNotMySQLServerFirewallRule)
	}

	r.SetConditions(runtimev1alpha1.Deleting())
	_, err := e.client.Delete(ctx, r.Spec.ForProvider.ResourceGroupName, r.Spec.ForProvider.ServerName, meta.GetExternalName(r))
	return errors.Wrap(resource.Ignore(azure.IsNotFound, err), errDeleteMySQLServerFirewallRule)
}
