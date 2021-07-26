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

package agentpool

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-03-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-03-01/containerservice/containerserviceapi"
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

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	azureclients "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/compute/agentpool"
)

// Error strings.
const (
	errNotAgentPool    = "managed resource is not an AgentPool"
	errCreateAgentPool = "cannot create AgentPool"
	errGetAgentPool    = "cannot get AgentPool"
	errDeleteAgentPool = "cannot delete AgentPool"
)

// SetupAgentPool adds a controller that reconciles AgentPools.
func SetupAgentPool(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha3.AgentPoolGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha3.AgentPool{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.AgentPoolGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(poll),
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
	cl := containerservice.NewAgentPoolsClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{c: cl}, nil
}

type external struct {
	c containerserviceapi.AgentPoolsClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha3.AgentPool)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAgentPool)
	}

	c, err := e.c.Get(ctx, cr.Spec.AgentPoolParameters.ResourceGroupName, cr.Spec.AgentPoolParameters.AKSClusterName, meta.GetExternalName(cr))
	if azure.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetAgentPool)
	}

	agentpool.UpdateStatus(cr, &c)
	needUpdate := agentpool.NeedUpdate(cr, &c)

	if cr.Status.State != "Succeeded" {
		return managed.ExternalObservation{
			ResourceExists: true,
		}, nil
	}

	cr.SetConditions(xpv1.Available())
	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: !needUpdate,
	}
	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha3.AgentPool)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAgentPool)
	}
	_, err := e.c.CreateOrUpdate(ctx, cr.Spec.ResourceGroupName, cr.Spec.AKSClusterName, meta.GetExternalName(cr), agentpool.New(cr))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateAgentPool)
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha3.AgentPool)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotAgentPool)
	}
	_, err := e.c.CreateOrUpdate(ctx, cr.Spec.ResourceGroupName, cr.Spec.AKSClusterName, meta.GetExternalName(cr), agentpool.New(cr))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errCreateAgentPool)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha3.AgentPool)
	if !ok {
		return errors.New(errNotAgentPool)
	}
	_, err := e.c.Delete(ctx, cr.Spec.AgentPoolParameters.ResourceGroupName, cr.Spec.AgentPoolParameters.AKSClusterName, meta.GetExternalName(cr))
	if azure.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, errDeleteAgentPool)
	}
	return nil
}
