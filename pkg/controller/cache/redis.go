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

package cache

import (
	"context"
	"strconv"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/redis/mgmt/2018-03-01/redis"
	"github.com/Azure/azure-sdk-for-go/services/redis/mgmt/2018-03-01/redis/redisapi"
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

	"github.com/crossplane/provider-azure/apis/cache/v1beta1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	redisclients "github.com/crossplane/provider-azure/pkg/clients/redis"
)

const (
	errNotRedis            = "the custom resource is not a Redis instance"
	errUpdateRedisCRFailed = "cannot update Redis custom resource instance"

	errConnectFailed        = "cannot connect to Azure API"
	errGetFailed            = "cannot get Redis instance from Azure API"
	errListAccessKeysFailed = "cannot get access key list"
	errCreateFailed         = "cannot create the Redis instance"
	errUpdateFailed         = "cannot update the Redis instance"
	errDeleteFailed         = "cannot delete the Redis instance"
)

// SetupRedis adds a controller that reconciles Redis resources.
func SetupRedis(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1beta1.RedisGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1beta1.Redis{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.RedisGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube client.Client
}

func (c connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azure.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errConnectFailed)
	}
	cl := redis.NewClient(creds[azure.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{kube: c.kube, client: cl}, nil
}

type external struct {
	kube   client.Client
	client redisapi.ClientAPI
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRedis)
	}
	cache, err := c.client.Get(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(resource.Ignore(azure.IsNotFound, err), errGetFailed)
	}

	redisclients.LateInitialize(&cr.Spec.ForProvider, cache)
	if err := c.kube.Update(ctx, cr); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpdateRedisCRFailed)
	}
	cr.Status.AtProvider = redisclients.GenerateObservation(cache)

	var conn managed.ConnectionDetails
	switch cr.Status.AtProvider.ProvisioningState {
	case redisclients.ProvisioningStateSucceeded:
		k, err := c.client.ListKeys(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errListAccessKeysFailed)
		}
		conn = managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretEndpointKey: []byte(cr.Status.AtProvider.HostName),
			xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(cr.Status.AtProvider.Port)),
			xpv1.ResourceCredentialsSecretPasswordKey: []byte(azure.ToString(k.PrimaryKey)),
		}
		cr.Status.SetConditions(xpv1.Available())
	case redisclients.ProvisioningStateCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case redisclients.ProvisioningStateDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  !redisclients.NeedsUpdate(cr.Spec.ForProvider, cache),
		ConnectionDetails: conn,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRedis)
	}
	cr.Status.SetConditions(xpv1.Creating())
	_, err := c.client.Create(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr), redisclients.NewCreateParameters(cr))
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRedis)
	}
	// NOTE(muvaf): redis service rejects updates while another operation
	// is ongoing.
	if cr.Status.AtProvider.ProvisioningState != redisclients.ProvisioningStateSucceeded {
		return managed.ExternalUpdate{}, nil
	}
	cache, err := c.client.Get(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetFailed)
	}
	_, err = c.client.Update(
		ctx,
		cr.Spec.ForProvider.ResourceGroupName,
		meta.GetExternalName(cr),
		redisclients.NewUpdateParameters(cr.Spec.ForProvider, cache))
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return errors.New(errNotRedis)
	}
	cr.Status.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.ProvisioningState == redisclients.ProvisioningStateDeleting {
		return nil
	}
	_, err := c.client.Delete(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
	return errors.Wrap(resource.Ignore(azure.IsNotFound, err), errDeleteFailed)
}
