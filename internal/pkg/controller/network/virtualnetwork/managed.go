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

package virtualnetwork

import (
	"context"

	azurenetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network/networkapi"
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

	azureclients "github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients"
	"github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients/network"
	"github.com/crossplane-contrib/provider-jet-azure/internal/pkg/features"

	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/network/v1alpha3"
	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/v1alpha1"
)

// Error strings.
const (
	errNotVirtualNetwork    = "managed resource is not an VirtualNetwork"
	errCreateVirtualNetwork = "cannot create VirtualNetwork"
	errUpdateVirtualNetwork = "cannot update VirtualNetwork"
	errGetVirtualNetwork    = "cannot get VirtualNetwork"
	errDeleteVirtualNetwork = "cannot delete VirtualNetwork"
)

// Setup adds a controller that reconciles VirtualNetworks.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha3.VirtualNetworkGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha3.VirtualNetwork{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.VirtualNetworkGroupVersionKind),
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
	creds, auth, err := azureclients.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := azurenetwork.NewVirtualNetworksClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client networkapi.VirtualNetworksClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	v, ok := mg.(*v1alpha3.VirtualNetwork)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotVirtualNetwork)
	}

	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), "")
	if azureclients.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetVirtualNetwork)
	}

	network.UpdateVirtualNetworkStatusFromAzure(v, az)

	v.SetConditions(xpv1.Available())

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	v, ok := mg.(*v1alpha3.VirtualNetwork)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotVirtualNetwork)
	}

	v.Status.SetConditions(xpv1.Creating())

	vnet := network.NewVirtualNetworkParameters(v)
	if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), vnet); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateVirtualNetwork)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	v, ok := mg.(*v1alpha3.VirtualNetwork)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotVirtualNetwork)
	}

	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), "")
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetVirtualNetwork)
	}

	if network.VirtualNetworkNeedsUpdate(v, az) {
		vnet := network.NewVirtualNetworkParameters(v)
		if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), vnet); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateVirtualNetwork)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	v, ok := mg.(*v1alpha3.VirtualNetwork)
	if !ok {
		return errors.New(errNotVirtualNetwork)
	}

	mg.SetConditions(xpv1.Deleting())

	_, err := e.client.Delete(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v))
	return errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errDeleteVirtualNetwork)
}
