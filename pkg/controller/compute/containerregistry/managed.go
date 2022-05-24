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

package containerregistry

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2019-05-01/containerregistry"
	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2019-05-01/containerregistry/containerregistryapi"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	azureclients "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/registry"
)

// Error strings.
const (
	errNotRegistry    = "managed resource is not a Registry"
	errCreateRegistry = "cannot create Registry"
	errGetRegistry    = "cannot get Registry"
	errDeleteRegistry = "cannot delete Registry"
	errUpdateRegistry = "cannot update Registry"
)

// SetupRegistry adds a controller that reconciles Registry.
func SetupRegistry(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha3.RegistryGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha3.Registry{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.RegistryGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azure.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := containerregistry.NewRegistriesClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client containerregistryapi.RegistriesClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha3.Registry)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRegistry)
	}
	az, err := e.client.Get(ctx, cr.Spec.ResourceGroupName, meta.GetExternalName(cr))
	if azure.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetRegistry)
	}
	registry.Update(cr, &az)
	upToDate := registry.UpToDate(&cr.Spec, &az)
	if registry.Initialized(cr) {
		if cr.Status.State == "Succeeded" {
			cr.SetConditions(xpv1.Available())
		}
	} else {
		// Do not trigger create/delete/update during initialization
		upToDate = true
	}
	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}
	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha3.Registry)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRegistry)
	}
	_, err := e.client.Create(ctx, cr.Spec.ResourceGroupName, meta.GetExternalName(cr), registry.New(cr.Spec))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateRegistry)
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha3.Registry)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRegistry)
	}
	_, err := e.client.Update(ctx, cr.Spec.ResourceGroupName, meta.GetExternalName(cr), registry.NewUpdateParams(cr.Spec))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateRegistry)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha3.Registry)
	if !ok {
		return errors.New(errNotRegistry)
	}
	_, err := e.client.Delete(ctx, cr.Spec.ResourceGroupName, meta.GetExternalName(cr))
	if err != nil {
		return errors.Wrap(err, errDeleteRegistry)
	}
	return nil
}
