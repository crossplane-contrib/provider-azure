/*
Copyright 2021 The Crossplane Authors.

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

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2019-05-01/containerregistry"

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2019-05-01/containerregistry/containerregistryapi"
)

var _ containerregistryapi.RegistriesClientAPI = &MockContainerRegistry{}

// MockContainerRegistry is a fake ContainerRegistry client.
type MockContainerRegistry struct {
	containerregistryapi.RegistriesClientAPI

	MockGet    func(ctx context.Context, resourceGroupName string, registryName string) (result containerregistry.Registry, err error)
	MockCreate func(ctx context.Context, resourceGroupName string, registryName string, registry containerregistry.Registry) (result containerregistry.RegistriesCreateFuture, err error)
	MockUpdate func(ctx context.Context, resourceGroupName string, registryName string, registryUpdateParameters containerregistry.RegistryUpdateParameters) (result containerregistry.RegistriesUpdateFuture, err error)
	MockDelete func(ctx context.Context, resourceGroupName string, registryName string) (result containerregistry.RegistriesDeleteFuture, err error)
}

// Get mock ContainerRegistry Get.
func (r *MockContainerRegistry) Get(ctx context.Context, resourceGroupName string, registryName string) (result containerregistry.Registry, err error) {
	return r.MockGet(ctx, resourceGroupName, registryName)
}

// Create mock ContainerRegistry Create.
func (r *MockContainerRegistry) Create(ctx context.Context, resourceGroupName string, registryName string, registry containerregistry.Registry) (result containerregistry.RegistriesCreateFuture, err error) {
	return r.MockCreate(ctx, resourceGroupName, registryName, registry)
}

// Update mock ContainerRegistry Update.
func (r *MockContainerRegistry) Update(ctx context.Context, resourceGroupName string, registryName string, registryUpdateParameters containerregistry.RegistryUpdateParameters) (result containerregistry.RegistriesUpdateFuture, err error) {
	return r.MockUpdate(ctx, resourceGroupName, registryName, registryUpdateParameters)
}

// Delete mock ContainerRegistry Delete.
func (r *MockContainerRegistry) Delete(ctx context.Context, resourceGroupName string, registryName string) (result containerregistry.RegistriesDeleteFuture, err error) {
	return r.MockDelete(ctx, resourceGroupName, registryName)
}
