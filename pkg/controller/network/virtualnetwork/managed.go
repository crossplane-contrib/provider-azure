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
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-azure/apis/network/v1alpha2"
	azurev1alpha2 "github.com/crossplaneio/stack-azure/apis/v1alpha2"
	azureclients "github.com/crossplaneio/stack-azure/pkg/clients"
	"github.com/crossplaneio/stack-azure/pkg/clients/network"
)

// Error strings.
const (
	errNewClient            = "cannot create new VirtualNetworks client"
	errNotVirtualNetwork    = "managed resource is not an VirtualNetwork"
	errCreateVirtualNetwork = "cannot create VirtualNetwork"
	errUpdateVirtualNetwork = "cannot update VirtualNetwork"
	errGetVirtualNetwork    = "cannot get VirtualNetwork"
	errDeleteVirtualNetwork = "cannot delete VirtualNetwork"
)

// Controller is responsible for adding the VirtualNetwork
// controller and its corresponding reconciler to the manager with any runtime configuration.
type Controller struct{}

// SetupWithManager creates a new VirtualNetwork Controller and adds it to the
// Manager with default RBAC. The Manager will set fields on the Controller and
// start it when the Manager is Started.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1alpha2.VirtualNetworkGroupVersionKind),
		resource.WithManagedConnectionPublishers(),
		resource.WithExternalConnecter(&connecter{client: mgr.GetClient()}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1alpha2.VirtualNetworkKind, v1alpha2.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha2.VirtualNetwork{}).
		Complete(r)
}

type connecter struct {
	client      client.Client
	newClientFn func(ctx context.Context, credentials []byte) (network.VirtualNetworksClient, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	g, ok := mg.(*v1alpha2.VirtualNetwork)
	if !ok {
		return nil, errors.New(errNotVirtualNetwork)
	}

	p := &azurev1alpha2.Provider{}
	n := meta.NamespacedNameOf(g.Spec.ProviderReference)
	if err := c.client.Get(ctx, n, p); err != nil {
		return nil, errors.Wrapf(err, "cannot get provider %s", n)
	}

	s := &corev1.Secret{}
	n = types.NamespacedName{Namespace: p.Spec.Secret.Namespace, Name: p.Spec.Secret.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrapf(err, "cannot get provider secret %s", n)
	}
	newClientFn := network.NewVirtualNetworksClient
	if c.newClientFn != nil {
		newClientFn = c.newClientFn
	}
	client, err := newClientFn(ctx, s.Data[p.Spec.Secret.Key])
	return &external{client: client}, errors.Wrap(err, errNewClient)
}

type external struct{ client network.VirtualNetworksClient }

func (e *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	v, ok := mg.(*v1alpha2.VirtualNetwork)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotVirtualNetwork)
	}

	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, v.Spec.Name, "")
	if azureclients.IsNotFound(err) {
		return resource.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, errGetVirtualNetwork)
	}

	network.UpdateVirtualNetworkStatusFromAzure(v, az)

	v.SetConditions(runtimev1alpha1.Available())

	o := resource.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: resource.ConnectionDetails{},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	v, ok := mg.(*v1alpha2.VirtualNetwork)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotVirtualNetwork)
	}

	v.Status.SetConditions(runtimev1alpha1.Creating())

	vnet := network.NewVirtualNetworkParameters(v)
	if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, v.Spec.Name, vnet); err != nil {
		return resource.ExternalCreation{}, errors.Wrap(err, errCreateVirtualNetwork)
	}

	return resource.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	v, ok := mg.(*v1alpha2.VirtualNetwork)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotVirtualNetwork)
	}

	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, v.Spec.Name, "")
	if err != nil {
		return resource.ExternalUpdate{}, errors.Wrap(err, errGetVirtualNetwork)
	}

	if network.VirtualNetworkNeedsUpdate(v, az) {
		vnet := network.NewVirtualNetworkParameters(v)
		if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, v.Spec.Name, vnet); err != nil {
			return resource.ExternalUpdate{}, errors.Wrap(err, errUpdateVirtualNetwork)
		}
	}
	return resource.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	v, ok := mg.(*v1alpha2.VirtualNetwork)
	if !ok {
		return errors.New(errNotVirtualNetwork)
	}

	mg.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.Delete(ctx, v.Spec.ResourceGroupName, v.Spec.Name)
	return errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errDeleteVirtualNetwork)
}
