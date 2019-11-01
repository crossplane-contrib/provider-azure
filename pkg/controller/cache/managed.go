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
	"strings"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/redis/mgmt/redis/redisapi"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-azure/apis/cache/v1alpha3"
	"github.com/crossplaneio/stack-azure/pkg/clients/redis"
)

// Controller is responsible for adding the MySQLServer controller and its
// corresponding reconciler to the manager with any runtime configuration.
type Controller struct{}

// SetupWithManager creates a new MySQLServer Controller and adds it to the
// Manager with default RBAC. The Manager will set fields on the Controller and
// start it when the Manager is Started.
func (c *Controller) SetupWithManager(mgr ctrl.Manager) error {
	r := resource.NewManagedReconciler(mgr,
		resource.ManagedKind(v1alpha3.RedisClassGroupVersionKind),
		resource.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: redis.NewClient}))

	name := strings.ToLower(fmt.Sprintf("%s.%s", v1alpha3.RedisKind, v1alpha3.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.Redis{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(ctx context.Context, credentials []byte) (redisapi.ClientAPI, error)
}

func (c connector) Connect(ctx context.Context, mg resource.Managed) (resource.ExternalClient, error) {
	return &external{}, nil
}

type external struct {
	kube   client.Client
	client redisapi.ClientAPI
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (resource.ExternalObservation, error) {
	return resource.ExternalObservation{}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (resource.ExternalCreation, error) {
	return resource.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (resource.ExternalUpdate, error) {
	return resource.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error { return nil }
