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

package publicipaddress

import (
	"context"

	azurenetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network/networkapi"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	azureclients "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/network"
)

// Error strings.
const (
	errNotPublicIPAddress    = "managed resource is not an PublicIPAddress"
	errCreatePublicIPAddress = "cannot create PublicIPAddress"
	errGetPublicIPAddress    = "cannot get PublicIPAddress"
	errDeletePublicIPAddress = "cannot delete PublicIPAddress"
)

// Setup adds a controller that reconciles Public Ip Address.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha3.PublicIPAddressGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha3.PublicIPAddress{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.PublicIPAddressGroupVersionKind),
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
	cl := azurenetwork.NewPublicIPAddressesClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client networkapi.PublicIPAddressesClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	s, ok := mg.(*v1alpha3.PublicIPAddress)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotPublicIPAddress)
	}

	az, err := e.client.Get(ctx, s.Spec.ResourceGroupName, meta.GetExternalName(s), "")
	if azureclients.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetPublicIPAddress)
	}

	network.UpdatePublicIPAddressStatusFromAzure(s, az)
	s.SetConditions(xpv1.Available())

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	s, ok := mg.(*v1alpha3.PublicIPAddress)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotPublicIPAddress)
	}

	s.Status.SetConditions(xpv1.Creating())

	snet := network.NewPublicIPAddressParameters(s)
	if _, err := e.client.CreateOrUpdate(ctx, s.Spec.ResourceGroupName, meta.GetExternalName(s), snet); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreatePublicIPAddress)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	_, ok := mg.(*v1alpha3.PublicIPAddress)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotPublicIPAddress)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	s, ok := mg.(*v1alpha3.PublicIPAddress)
	if !ok {
		return errors.New(errNotPublicIPAddress)
	}

	mg.SetConditions(xpv1.Deleting())

	_, err := e.client.Delete(ctx, s.Spec.ResourceGroupName, meta.GetExternalName(s))
	return errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errDeletePublicIPAddress)
}
