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

	"github.com/Azure/azure-sdk-for-go/profiles/latest/redis/mgmt/redis/redisapi"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/cache/v1beta1"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/redis"
)

const (
	errNotRedis                = "the custom resource is not a Redis instance"
	errUpdateRedisCRFailed     = "cannot update Redis custom resource instance"
	errGetProviderFailed       = "cannot get provider"
	errGetProviderSecretFailed = "cannot get provider secret"
	errProviderSecretNil       = "provider does not have a secret reference"

	errConnectFailed        = "cannot connect to Azure API"
	errGetFailed            = "cannot get Redis instance from Azure API"
	errListAccessKeysFailed = "cannot get access key list"
	errCreateFailed         = "cannot create the Redis instance"
	errUpdateFailed         = "cannot update the Redis instance"
	errDeleteFailed         = "cannot delete the Redis instance"
)

// SetupRedis adds a controller that reconciles Redis resources.
func SetupRedis(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1beta1.RedisGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.Redis{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1beta1.RedisGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: redis.NewClient}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte) (redisapi.ClientAPI, error)
}

func (c connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return nil, errors.New(errNotRedis)
	}
	p := &azurev1alpha3.Provider{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.Spec.ProviderReference.Name}, p); err != nil {
		return nil, errors.Wrap(err, errGetProviderFailed)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecretFailed)
	}
	rclient, err := c.newClientFn(ctx, s.Data[p.Spec.CredentialsSecretRef.Key])
	return &external{kube: c.kube, client: rclient}, errors.Wrap(err, errConnectFailed)
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

	redis.LateInitialize(&cr.Spec.ForProvider, cache)
	if err := c.kube.Update(ctx, cr); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpdateRedisCRFailed)
	}
	cr.Status.AtProvider = redis.GenerateObservation(cache)

	var conn managed.ConnectionDetails
	switch cr.Status.AtProvider.ProvisioningState {
	case redis.ProvisioningStateSucceeded:
		k, err := c.client.ListKeys(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errListAccessKeysFailed)
		}
		conn = managed.ConnectionDetails{
			runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(cr.Status.AtProvider.HostName),
			runtimev1alpha1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(cr.Status.AtProvider.Port)),
			runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(azure.ToString(k.PrimaryKey)),
		}
		cr.Status.SetConditions(runtimev1alpha1.Available())
		resource.SetBindable(cr)
	case redis.ProvisioningStateCreating:
		cr.Status.SetConditions(runtimev1alpha1.Creating())
	case redis.ProvisioningStateDeleting:
		cr.Status.SetConditions(runtimev1alpha1.Deleting())
	default:
		cr.Status.SetConditions(runtimev1alpha1.Unavailable())
	}
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  !redis.NeedsUpdate(cr.Spec.ForProvider, cache),
		ConnectionDetails: conn,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRedis)
	}
	cr.Status.SetConditions(runtimev1alpha1.Creating())
	_, err := c.client.Create(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr), redis.NewCreateParameters(cr))
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRedis)
	}
	// NOTE(muvaf): redis service rejects updates while another operation
	// is ongoing.
	if cr.Status.AtProvider.ProvisioningState != redis.ProvisioningStateSucceeded {
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
		redis.NewUpdateParameters(cr.Spec.ForProvider, cache))
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return errors.New(errNotRedis)
	}
	cr.Status.SetConditions(runtimev1alpha1.Deleting())
	if cr.Status.AtProvider.ProvisioningState == redis.ProvisioningStateDeleting {
		return nil
	}
	_, err := c.client.Delete(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
	return errors.Wrap(resource.Ignore(azure.IsNotFound, err), errDeleteFailed)
}
