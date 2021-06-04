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

package ddosprotectionplan

import (
	"context"

	azurenetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network/networkapi"
	"github.com/google/go-cmp/cmp"

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

	azureclients "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/network"

	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
)

// Error strings.
const (
	errNotDdosProtectionPlan    = "managed resource is not an Ddos Protection Plan"
	errCreateDdosProtectionPlan = "cannot create Ddos Protection Plan"
	errUpdateDdosProtectionPlan = "cannot update Ddos Protection Plan"
	errGetDdosProtectionPlan    = "cannot get Ddos Protection Plan"
	errDeleteDdosProtectionPlan = "cannot delete Ddos Protection Plan"
)

// Setup adds a controller that reconciles Ddos Protection Plans.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha3.DdosProtectionPlanGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha3.DdosProtectionPlan{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.DdosProtectionPlanGroupVersionKind),
			managed.WithConnectionPublishers(),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
}

type external struct {
	client networkapi.DdosProtectionPlansClientAPI
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azureclients.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := azurenetwork.NewDdosProtectionPlansClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	d, ok := mg.(*v1alpha3.DdosProtectionPlan)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDdosProtectionPlan)
	}

	az, err := e.client.Get(ctx, d.Spec.ResourceGroupName, meta.GetExternalName(d))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errGetDdosProtectionPlan)
	}

	currentDdpp := d.DeepCopy()
	network.LateInitializeDdos(currentDdpp, az)
	if !cmp.Equal(currentDdpp.Spec, d.Spec) {
		if _, err := e.Update(ctx, mg); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdateDdosProtectionPlan)
		}
	}

	network.UpdateDdosProtectionPlanStatusFromAzure(d, az)
	d.SetConditions(xpv1.Available())
	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: network.IsDdosProtectionPlanUpToDate(d, az),
		ConnectionDetails: managed.ConnectionDetails{
			"ddpp": []byte(meta.GetExternalName(d)),
		},
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	d, ok := mg.(*v1alpha3.DdosProtectionPlan)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDdosProtectionPlan)
	}

	d.Status.SetConditions(xpv1.Creating())

	ddos := network.NewDdosProtectionPlanParameters(d)

	if _, err := e.client.CreateOrUpdate(ctx, d.Spec.ResourceGroupName, meta.GetExternalName(d), ddos); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateDdosProtectionPlan)
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	d, ok := mg.(*v1alpha3.DdosProtectionPlan)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDdosProtectionPlan)
	}

	az, err := e.client.Get(ctx, d.Spec.ResourceGroupName, meta.GetExternalName(d))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetDdosProtectionPlan)
	}

	if !network.IsDdosProtectionPlanUpToDate(d, az) {
		ddos := network.NewDdosProtectionPlanParameters(d)
		if _, err := e.client.CreateOrUpdate(ctx, d.Spec.ResourceGroupName, meta.GetExternalName(d), ddos); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateDdosProtectionPlan)
		}
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	d, ok := mg.(*v1alpha3.DdosProtectionPlan)
	if !ok {
		return errors.New(errNotDdosProtectionPlan)
	}

	mg.SetConditions(xpv1.Deleting())

	_, err := e.client.Delete(ctx, d.Spec.ResourceGroupName, meta.GetExternalName(d))
	return errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errDeleteDdosProtectionPlan)
}
