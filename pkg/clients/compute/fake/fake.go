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

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute/computeapi"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
)

// AKSClient is a fake AKS client.
type AKSClient struct {
	MockGetManagedCluster    func(ctx context.Context, ac *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error)
	MockEnsureManagedCluster func(ctx context.Context, ac *v1alpha3.AKSCluster, secret string) error
	MockDeleteManagedCluster func(ctx context.Context, ac *v1alpha3.AKSCluster) error
	MockGetKubeConfig        func(ctx context.Context, ac *v1alpha3.AKSCluster) ([]byte, error)
}

// GetManagedCluster calls MockGetManagedCluster.
func (c AKSClient) GetManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error) {
	return c.MockGetManagedCluster(ctx, ac)
}

// EnsureManagedCluster calls MockEnsureManagedCluster.
func (c AKSClient) EnsureManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster, secret string) error {
	return c.MockEnsureManagedCluster(ctx, ac, secret)
}

// DeleteManagedCluster calls DeleteManagedCluster.
func (c AKSClient) DeleteManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) error {
	return c.MockDeleteManagedCluster(ctx, ac)
}

// GetKubeConfig calls GetKubeConfig.
func (c AKSClient) GetKubeConfig(ctx context.Context, ac *v1alpha3.AKSCluster) ([]byte, error) {
	return c.MockGetKubeConfig(ctx, ac)
}

var _ computeapi.VirtualMachinesClientAPI = &VirtualMachineClient{}

// VirtualMachineClient is a fake VirtualMachine client.
type VirtualMachineClient struct {
	computeapi.VirtualMachinesClientAPI
	MockCreateOrUpdate func(ctx context.Context, resourceGroupName string, vmName string, parameters compute.VirtualMachine) (result compute.VirtualMachinesCreateOrUpdateFuture, err error)
	MockDelete         func(ctx context.Context, resourceGroupName string, vmName string) (result compute.VirtualMachinesDeleteFuture, err error)
	MockGet            func(ctx context.Context, resourceGroupName string, vmName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error)
}

// CreateOrUpdate calls MockCreateOrUpdate.
func (c *VirtualMachineClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, vmName string, parameters compute.VirtualMachine) (result compute.VirtualMachinesCreateOrUpdateFuture, err error) {
	return c.MockCreateOrUpdate(ctx, resourceGroupName, vmName, parameters)
}

// Delete calls MockDelete.
func (c *VirtualMachineClient) Delete(ctx context.Context, resourceGroupName string, vmName string) (result compute.VirtualMachinesDeleteFuture, err error) {
	return c.MockDelete(ctx, resourceGroupName, vmName)
}

// Get calls MockGet.
func (c *VirtualMachineClient) Get(ctx context.Context, resourceGroupName string, vmName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
	return c.MockGet(ctx, resourceGroupName, vmName, expand)
}

var _ computeapi.DisksClientAPI = &DiskClient{}

// DiskClient is a fake Disk client.
type DiskClient struct {
	computeapi.DisksClientAPI
	MockGet    func(ctx context.Context, resourceGroupName string, diskName string) (result compute.Disk, err error)
	MockDelete func(ctx context.Context, resourceGroupName string, diskName string) (result compute.DisksDeleteFuture, err error)
}

// Get calls MockGet.
func (c *DiskClient) Get(ctx context.Context, resourceGroupName string, diskName string) (result compute.Disk, err error) {
	return c.MockGet(ctx, resourceGroupName, diskName)
}

// Delete calls MockDelete.
func (c *DiskClient) Delete(ctx context.Context, resourceGroupName string, diskName string) (result compute.DisksDeleteFuture, err error) {
	return c.MockDelete(ctx, resourceGroupName, diskName)
}
