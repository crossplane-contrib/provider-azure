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

package registry

import (
	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2019-05-01/containerregistry"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

// New returns an Azure Registry object from a Registry spec
func New(r v1alpha3.RegistrySpec) containerregistry.Registry {
	return containerregistry.Registry{
		Sku: &containerregistry.Sku{
			Name: containerregistry.SkuName(r.Sku),
			Tier: containerregistry.SkuTier(r.Sku),
		},
		RegistryProperties: &containerregistry.RegistryProperties{
			AdminUserEnabled: azure.ToBoolPtr(r.AdminUserEnabled),
		},
		Location: azure.ToStringPtr(r.Location),
	}
}

// NewUpdateParams returns an Azure Registry object from a Registry spec
func NewUpdateParams(r v1alpha3.RegistrySpec) containerregistry.RegistryUpdateParameters {
	return containerregistry.RegistryUpdateParameters{
		Sku: &containerregistry.Sku{
			Name: containerregistry.SkuName(r.Sku),
		},
		RegistryPropertiesUpdateParameters: &containerregistry.RegistryPropertiesUpdateParameters{
			AdminUserEnabled: azure.ToBoolPtr(r.AdminUserEnabled),
		},
	}
}

// UpToDate determines if a Registry is up to date
func UpToDate(r *v1alpha3.RegistrySpec, az *containerregistry.Registry) bool {
	if r.Sku != "" {
		if r.Sku != string(az.Sku.Name) {
			return false
		}
		if r.Sku != string(az.Sku.Tier) {
			return false
		}
	}
	if r.AdminUserEnabled != azure.ToBool(az.AdminUserEnabled) {
		return false
	}
	return true
}

// Initialized determines if a Registry has been initialized
func Initialized(r *v1alpha3.Registry) bool {
	if r.Status.State == "Succeeded" || r.Status.State == "Failed" || r.Status.State == "Canceled" {
		return true
	}
	return false
}

// Update updates the status related to the external
// Azure Registry in the RegistryStatus
func Update(r *v1alpha3.Registry, az *containerregistry.Registry) {
	r.Status.State = string(az.ProvisioningState)
	r.Status.ProviderID = azure.ToString(az.ID)
	r.Status.LoginServer = azure.ToString(az.LoginServer)
	if az.Status != nil {
		r.Status.Status = azure.ToString(az.Status.DisplayStatus)
		r.Status.StatusMessage = azure.ToString(az.Status.Message)
	} else {
		r.Status.Status = ""
		r.Status.StatusMessage = ""
	}
}
