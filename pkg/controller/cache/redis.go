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
	"fmt"
	"strconv"
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/redis/mgmt/redis/redisapi"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-azure/apis/cache/v1beta1"
	azurev1alpha3 "github.com/crossplaneio/stack-azure/apis/v1alpha3"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
	"github.com/crossplaneio/stack-azure/pkg/clients/redis"
)

const (
	errNotRedis                = "the custom resource is not a Redis instance"
	errUpdateRedisCRFailed     = "cannot update Redis custom resource instance"
	errGetProviderFailed       = "cannot get provider"
	errGetProviderSecretFailed = "cannot get provider secret"

	errConnectFailed        = "cannot connect to Azure API"
	errGetFailed            = "cannot get Redis instance from Azure API"
	errListAccessKeysFailed = "cannot get access key list"
	errCreateFailed         = "cannot create the Redis instance"
	errUpdateFailed         = "cannot update the Redis instance"
	errDeleteFailed         = "cannot delete the Redis instance"
)

// RedisController is responsible for adding the MySQLServer controller and its
// corresponding reconciler to the manager with any runtime configuration.
type RedisController struct{}

// SetupWithManager creates a new MySQLServer RedisController and adds it to the
// Manager with default RBAC. The Manager will set fields on the RedisController and
// start it when the Manager is Started.
func (c *RedisController) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1beta1.RedisGroupVersionKind),
		resource.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: redis.NewClient}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1beta1.RedisKind, v1beta1.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.Redis{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte) (redisapi.ClientAPI, error)
}

func (c connector) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return nil, errors.New(errNotRedis)
	}
	p := &azurev1alpha3.Provider{}
	if err := c.kube.Get(ctx, meta.NamespacedNameOf(cr.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProviderFailed)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.Secret.Namespace, Name: p.Spec.Secret.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecretFailed)
	}
	rclient, err := c.newClientFn(ctx, s.Data[p.Spec.Secret.Key])
	return &external{kube: c.kube, client: rclient}, errors.Wrap(err, errConnectFailed)
}

type external struct {
	kube   client.Client
	client redisapi.ClientAPI
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return resource.ExternalObservation{}, errors.New(errNotRedis)
	}
	cache, err := c.client.Get(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
	if err != nil {
		return resource.ExternalObservation{ResourceExists: false}, errors.Wrap(resource.Ignore(azure.IsNotFound, err), errGetFailed)
	}

	redis.LateInitialize(&cr.Spec.ForProvider, cache)
	if err := c.kube.Update(ctx, cr); err != nil {
		return resource.ExternalObservation{}, errors.Wrap(err, errUpdateRedisCRFailed)
	}
	cr.Status.AtProvider = redis.GenerateObservation(cache)

	var conn resource.ConnectionDetails
	switch cr.Status.AtProvider.ProvisioningState {
	case redis.ProvisioningStateSucceeded:
		k, err := c.client.ListKeys(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
		if err != nil {
			return resource.ExternalObservation{}, errors.Wrap(err, errListAccessKeysFailed)
		}
		conn = resource.ConnectionDetails{
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
	return resource.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  !redis.NeedsUpdate(cr.Spec.ForProvider, cache),
		ConnectionDetails: conn,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return resource.ExternalCreation{}, errors.New(errNotRedis)
	}
	cr.Status.SetConditions(runtimev1alpha1.Creating())
	_, err := c.client.Create(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr), redis.NewCreateParameters(cr))
	return resource.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Redis)
	if !ok {
		return resource.ExternalUpdate{}, errors.New(errNotRedis)
	}
	// NOTE(muvaf): redis service rejects updates while another operation
	// is ongoing.
	if cr.Status.AtProvider.ProvisioningState != redis.ProvisioningStateSucceeded {
		return resource.ExternalUpdate{}, nil
	}
	cache, err := c.client.Get(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
	if err != nil {
		return resource.ExternalUpdate{}, errors.Wrap(err, errGetFailed)
	}
	_, err = c.client.Update(
		ctx,
		cr.Spec.ForProvider.ResourceGroupName,
		meta.GetExternalName(cr),
		redis.NewUpdateParameters(cr.Spec.ForProvider, cache))
	return resource.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
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
