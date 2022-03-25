/*
Copyright 2020 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the c.Specific language governing permissions and
limitations under the License.
*/

package fake

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-01-01/containerservice"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
)

// AKSClient is a fake AKS client.
type AKSClient struct {
	MockGetManagedCluster    func(ctx context.Context, ac *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error)
	MockEnsureManagedCluster func(ctx context.Context, ac *v1alpha3.AKSCluster) error
	MockDeleteManagedCluster func(ctx context.Context, ac *v1alpha3.AKSCluster) error
	MockGetKubeConfig        func(ctx context.Context, ac *v1alpha3.AKSCluster) ([]byte, error)
}

// GetManagedCluster calls MockGetManagedCluster.
func (c AKSClient) GetManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error) {
	return c.MockGetManagedCluster(ctx, ac)
}

// EnsureManagedCluster calls MockEnsureManagedCluster.
func (c AKSClient) EnsureManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) error {
	return c.MockEnsureManagedCluster(ctx, ac)
}

// DeleteManagedCluster calls DeleteManagedCluster.
func (c AKSClient) DeleteManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) error {
	return c.MockDeleteManagedCluster(ctx, ac)
}

// GetKubeConfig calls GetKubeConfig.
func (c AKSClient) GetKubeConfig(ctx context.Context, ac *v1alpha3.AKSCluster) ([]byte, error) {
	return c.MockGetKubeConfig(ctx, ac)
}
