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

package expressroutecircuits

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
	errNotExpressRouteCircuits    = "managed resource is not an ExpressRouteCircuits"
	errCreateExpressRouteCircuits = "cannot create ExpressRouteCircuits"
	errUpdateExpressRouteCircuits = "cannot update ExpressRouteCircuits"
	errGetExpressRouteCircuits    = "cannot get ExpressRouteCircuits"
	errDeleteExpressRouteCircuits = "cannot delete ExpressRouteCircuits"
)

// Setup adds a controller that reconciles ExpressRouteCircuits.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha3.ExpressRouteCircuitsGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha3.ExpressRouteCircuits{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.ExpressRouteCircuitsGroupVersionKind),
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
	cl := azurenetwork.NewExpressRouteCircuitsClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client networkapi.ExpressRouteCircuitsClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	exp, ok := mg.(*v1alpha3.ExpressRouteCircuits)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotExpressRouteCircuits)
	}

	az, err := e.client.Get(ctx, exp.Spec.ResourceGroupName, meta.GetExternalName(exp))

	if azureclients.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetExpressRouteCircuits)
	}

	network.UpdateExpressRouteCircuitStatusFromAzure(exp, az)
	exp.SetConditions(xpv1.Available())

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	exp, ok := mg.(*v1alpha3.ExpressRouteCircuits)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotExpressRouteCircuits)
	}

	exp.Status.SetConditions(xpv1.Creating())

	ercParams := network.NewExpressRouteCircuitsParameters(exp)
	if _, err := e.client.CreateOrUpdate(ctx, exp.Spec.ResourceGroupName, meta.GetExternalName(exp), ercParams); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateExpressRouteCircuits)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	exp, ok := mg.(*v1alpha3.ExpressRouteCircuits)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotExpressRouteCircuits)
	}

	az, err := e.client.Get(ctx, exp.Spec.ResourceGroupName, meta.GetExternalName(exp))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetExpressRouteCircuits)
	}

	if network.ExpressRouteCircuitNeedsUpdate(exp, az) {
		ercParams := network.NewExpressRouteCircuitsParameters(exp)
		if _, err := e.client.CreateOrUpdate(ctx, exp.Spec.ResourceGroupName, meta.GetExternalName(exp), ercParams); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateExpressRouteCircuits)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	exp, ok := mg.(*v1alpha3.ExpressRouteCircuits)
	if !ok {
		return errors.New(errNotExpressRouteCircuits)
	}

	mg.SetConditions(xpv1.Deleting())

	_, err := e.client.Delete(ctx, exp.Spec.ResourceGroupName, meta.GetExternalName(exp))
	return errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errDeleteExpressRouteCircuits)
}
