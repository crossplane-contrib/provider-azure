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
package AzureFirewall

import (
	"context"
	azurenetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	azureclients "github.com/crossplane/provider-azure/pkg/clients"
	azurefirewall "github.com/crossplane/provider-azure/pkg/clients/network"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Error strings.
const (
	errNotAzureFirewall    = "managed resource is not an AzureFirewall"
	errCreateAzureFirewall = "cannot create AzureFirewall"
	errUpdateAzureFirewall = "cannot update AzureFirewall"
	errGetAzureFirewall    = "cannot get AzureFirewall"
	errDeleteAzureFirewall = "cannot delete AzureFirewall"
)

// Setup adds a controller that reconciles Security Group.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.AzureFirewallKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.AzureFirewall{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.AzureFirewallGroupVersionKind),
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
	creds, auth, err := azureclients.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := azurenetwork.NewAzureFirewallsClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client azurenetwork.AzureFirewallsClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	v, ok := mg.(*v1alpha3.AzureFirewall)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAzureFirewall)
	}
	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, v.Name)

	if azureclients.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetAzureFirewall)
	}

	if az.Name != nil {
		azurefirewall.UpdateAzureFirewallStatusFromAzure(v, az)
	}

	v.SetConditions(runtimev1alpha1.Available())

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}
	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	v, ok := mg.(*v1alpha3.AzureFirewall)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAzureFirewall)
	}
	v.Status.SetConditions(runtimev1alpha1.Creating())

	af := azurefirewall.NewAzureFirewallParameters(v)

	if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), af); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateAzureFirewall)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	v, ok := mg.(*v1alpha3.AzureFirewall)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotAzureFirewall)
	}
	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, v.Name)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errNotAzureFirewall)
	}
	if azurefirewall.AzureFirewallNeedsUpdate(v, az) {
		vnet := azurefirewall.NewAzureFirewallParameters(v)
		if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), vnet); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateAzureFirewall)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	v, ok := mg.(*v1alpha3.AzureFirewall)
	if !ok {
		return errors.New(errNotAzureFirewall)
	}

	mg.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.Delete(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v))
	return errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errDeleteAzureFirewall)
}
