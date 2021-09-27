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

package network

import (
	"reflect"

	networkmgmt "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"

	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

// NewVirtualNetworkParameters returns an Azure VirtualNetwork object from a virtual network spec
func NewVirtualNetworkParameters(v *v1alpha3.VirtualNetwork) networkmgmt.VirtualNetwork {
	return networkmgmt.VirtualNetwork{
		Location: azure.ToStringPtr(v.Spec.Location),
		Tags:     azure.ToStringPtrMap(v.Spec.Tags),
		VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
			EnableDdosProtection: azure.ToBoolPtr(v.Spec.VirtualNetworkPropertiesFormat.EnableDDOSProtection, azure.FieldRequired),
			EnableVMProtection:   azure.ToBoolPtr(v.Spec.VirtualNetworkPropertiesFormat.EnableVMProtection),
			AddressSpace: &networkmgmt.AddressSpace{
				AddressPrefixes: &v.Spec.VirtualNetworkPropertiesFormat.AddressSpace.AddressPrefixes,
			},
		},
	}
}

// VirtualNetworkNeedsUpdate determines if a virtual network need to be updated
func VirtualNetworkNeedsUpdate(kube *v1alpha3.VirtualNetwork, az networkmgmt.VirtualNetwork) bool {
	up := NewVirtualNetworkParameters(kube)

	switch {
	case !reflect.DeepEqual(up.VirtualNetworkPropertiesFormat.AddressSpace, az.VirtualNetworkPropertiesFormat.AddressSpace):
		return true
	case !reflect.DeepEqual(up.VirtualNetworkPropertiesFormat.EnableDdosProtection, az.VirtualNetworkPropertiesFormat.EnableDdosProtection):
		return true
	case !reflect.DeepEqual(up.VirtualNetworkPropertiesFormat.EnableVMProtection, az.VirtualNetworkPropertiesFormat.EnableVMProtection):
		return true
	case !reflect.DeepEqual(up.Tags, az.Tags):
		return true
	}

	return false
}

// UpdateVirtualNetworkStatusFromAzure updates the status related to the external
// Azure virtual network in the VirtualNetworkStatus
func UpdateVirtualNetworkStatusFromAzure(v *v1alpha3.VirtualNetwork, az networkmgmt.VirtualNetwork) {
	v.Status.State = azure.ToString(az.ProvisioningState)
	v.Status.ID = azure.ToString(az.ID)
	v.Status.Etag = azure.ToString(az.Etag)
	v.Status.ResourceGUID = azure.ToString(az.ResourceGUID)
	v.Status.Type = azure.ToString(az.Type)
}

// NewSubnetParameters returns an Azure Subnet object from a subnet spec
func NewSubnetParameters(s *v1alpha3.Subnet) networkmgmt.Subnet {
	return networkmgmt.Subnet{
		SubnetPropertiesFormat: &networkmgmt.SubnetPropertiesFormat{
			AddressPrefix:    azure.ToStringPtr(s.Spec.SubnetPropertiesFormat.AddressPrefix),
			ServiceEndpoints: NewServiceEndpoints(s.Spec.SubnetPropertiesFormat.ServiceEndpoints),
		},
	}
}

// NewPublicIPAddressParameters returns an Azure PublicIPAddress object from a public ip address spec
func NewPublicIPAddressParameters(s *v1alpha3.PublicIPAddress) networkmgmt.PublicIPAddress {
	return networkmgmt.PublicIPAddress{
		Sku: NewPublicIPAddressSKU(s.Spec.PublicIPAddressFormat.SKU),
		PublicIPAddressPropertiesFormat: &networkmgmt.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: networkmgmt.IPAllocationMethod(s.Spec.PublicIPAddressFormat.PublicIPAllocationMethod),
			PublicIPAddressVersion:   networkmgmt.IPVersion(s.Spec.PublicIPAddressFormat.PublicIPAddressVersion),
		},
		Location: s.Spec.PublicIPAddressFormat.Location,
	}
}

// NewPublicIPAddressSKU returns an Azure PublicIPAddressSku object from a public ip address sku
func NewPublicIPAddressSKU(s *v1alpha3.SKU) *networkmgmt.PublicIPAddressSku {
	if s == nil {
		return nil
	}
	return &networkmgmt.PublicIPAddressSku{
		Name: networkmgmt.PublicIPAddressSkuName(s.Name),
	}
}

// NewNetworkInterfaceParameters returns an Azure NetworkInterface object from a network interface
func NewNetworkInterfaceParameters(s *v1alpha3.NetworkInterface) networkmgmt.Interface {
	return networkmgmt.Interface{
		InterfacePropertiesFormat: &networkmgmt.InterfacePropertiesFormat{
			Primary:          azure.ToBoolPtr(true),
			IPConfigurations: NewInterfaceIPConfiguration(s),
		},
		Location: azure.ToStringPtr(s.Spec.NetworkInterfaceFormat.Location),
		Tags:     s.Spec.Tags,
	}
}

// NewServiceEndpoints converts to Azure ServiceEndpointPropertiesFormat
func NewServiceEndpoints(e []v1alpha3.ServiceEndpointPropertiesFormat) *[]networkmgmt.ServiceEndpointPropertiesFormat {
	endpoints := make([]networkmgmt.ServiceEndpointPropertiesFormat, len(e))

	for i, end := range e {
		endpoints[i] = networkmgmt.ServiceEndpointPropertiesFormat{
			Service: azure.ToStringPtr(end.Service),
		}
	}

	return &endpoints
}

// NewInterfaceIPConfiguration converts to Azure InterfaceIPConfiguration
func NewInterfaceIPConfiguration(s *v1alpha3.NetworkInterface) *[]networkmgmt.InterfaceIPConfiguration {
	ifaces := s.Spec.NetworkInterfaceFormat.IPConfigurations
	interfaces := make([]networkmgmt.InterfaceIPConfiguration, len(ifaces))

	for i, iface := range ifaces {
		var publicIP *networkmgmt.PublicIPAddress
		if iface.PublicIPAddressID != "" {
			publicIP = &networkmgmt.PublicIPAddress{ID: azure.ToStringPtr(iface.PublicIPAddressID)}
		}
		interfaces[i] = networkmgmt.InterfaceIPConfiguration{
			Name: azure.ToStringPtr(iface.Name),
			InterfaceIPConfigurationPropertiesFormat: &networkmgmt.InterfaceIPConfigurationPropertiesFormat{
				Primary:         iface.Primary,
				Subnet:          &networkmgmt.Subnet{ID: azure.ToStringPtr(iface.SubnetID)},
				PublicIPAddress: publicIP,
			},
		}
	}

	return &interfaces
}

// SubnetNeedsUpdate determines if a virtual network need to be updated
func SubnetNeedsUpdate(kube *v1alpha3.Subnet, az networkmgmt.Subnet) bool {
	up := NewSubnetParameters(kube)

	return !reflect.DeepEqual(up.SubnetPropertiesFormat.AddressPrefix, az.SubnetPropertiesFormat.AddressPrefix)
}

// UpdateSubnetStatusFromAzure updates the status related to the external
// Azure subnet in the SubnetStatus
func UpdateSubnetStatusFromAzure(v *v1alpha3.Subnet, az networkmgmt.Subnet) {
	v.Status.State = azure.ToString(az.ProvisioningState)
	v.Status.Etag = azure.ToString(az.Etag)
	v.Status.ID = azure.ToString(az.ID)
	v.Status.Purpose = azure.ToString(az.Purpose)
}

// UpdatePublicIPAddressStatusFromAzure updates the status related to the external
// Azure public ip address in the PublicIPAddressStatus
func UpdatePublicIPAddressStatusFromAzure(v *v1alpha3.PublicIPAddress, az networkmgmt.PublicIPAddress) {
	v.Status.State = azure.ToString(az.ProvisioningState)
	v.Status.Etag = azure.ToString(az.Etag)
	v.Status.ID = azure.ToString(az.ID)
	v.Status.Address = azure.ToString(az.IPAddress)
}

// UpdateNetworkInterfaceStatusFromAzure updates the status related to the external
// Azure network interface in the NetworkInterfaceStatus
func UpdateNetworkInterfaceStatusFromAzure(v *v1alpha3.NetworkInterface, az networkmgmt.Interface) {
	v.Status.State = azure.ToString(az.InterfacePropertiesFormat.ProvisioningState)
	v.Status.Etag = azure.ToString(az.Etag)
	v.Status.ID = azure.ToString(az.ID)
}
