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

package privateendpoint

import (
	"context"

	azurenetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-03-01/network"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-03-01/network/networkapi"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-azure/apis/network/v1alpha3"
	"github.com/crossplane-contrib/provider-azure/apis/v1alpha1"
	azureclients "github.com/crossplane-contrib/provider-azure/pkg/clients"
	"github.com/crossplane-contrib/provider-azure/pkg/clients/network"
	"github.com/crossplane-contrib/provider-azure/pkg/features"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// Error strings.
const (
	errNotPrivateEndpoint    = "managed resource is not an PrivateEndpoint"
	errCreatePrivateEndpoint = "cannot create PrivateEndpoint"
	errUpdatePrivateEndpoint = "cannot update PrivateEndpoint"
	errGetPrivateEndpoint    = "cannot get PrivateEndpoint"
	errDeletePrivateEndpoint = "cannot delete PrivateEndpoint"
)

// Setup adds a controller that reconciles PrivateEndpoint.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha3.PrivateEndpointKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha3.PrivateEndpoint{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.PrivateEndpointVersionKind),
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
	cl := azurenetwork.NewPrivateEndpointsClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client networkapi.PrivateEndpointsClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	v, ok := mg.(*v1alpha3.PrivateEndpoint)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPrivateEndpoint)
	}

	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), "")
	if azureclients.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPrivateEndpoint)
	}

	network.UpdatePrivateEndpointStatusFromAzure(v, az)

	v.SetConditions(xpv1.Available())

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	v, ok := mg.(*v1alpha3.PrivateEndpoint)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPrivateEndpoint)
	}

	v.Status.SetConditions(xpv1.Creating())

	endpoint := network.NewPrivateEndpointParameters(v)

	if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), endpoint); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePrivateEndpoint)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	v, ok := mg.(*v1alpha3.PrivateEndpoint)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPrivateEndpoint)
	}

	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), "")
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetPrivateEndpoint)
	}

	if network.PrivateEndpointNeedsUpdate(v, az) {
		pe := network.NewPrivateEndpointParameters(v)
		if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), pe); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdatePrivateEndpoint)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	v, ok := mg.(*v1alpha3.PrivateEndpoint)
	if !ok {
		return errors.New(errNotPrivateEndpoint)
	}

	mg.SetConditions(xpv1.Deleting())

	// Do not reissue deletion requests if PrivateEndpoint provisioning state reported is already showing an in-progress deletion.
	if v.Status.AtProvider.State == network.ProvisioningStateDeleting {
		return nil
	}

	_, err := e.client.Delete(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v))
	return errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errDeletePrivateEndpoint)
}
