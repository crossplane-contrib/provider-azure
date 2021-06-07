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

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
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
			EnableVMProtection:   azure.ToBoolPtr(v.Spec.VirtualNetworkPropertiesFormat.EnableVMProtection, azure.FieldRequired),
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

// NewExpressRouteCircuitsParameters returns an Azure ExpressRouteCircuits object from a expressroutecircuits spec
func NewExpressRouteCircuitsParameters(d *v1alpha3.ExpressRouteCircuits) networkmgmt.ExpressRouteCircuit {
	return networkmgmt.ExpressRouteCircuit{
		Location: azure.ToStringPtr(d.Spec.Location),
		Tags:     azure.ToStringPtrMap(d.Spec.Tags),
		Sku: &networkmgmt.ExpressRouteCircuitSku{
			Name:   azure.ToStringPtr(d.Spec.Sku.Tier + "_" + d.Spec.Sku.Family),
			Tier:   network.ExpressRouteCircuitSkuTier(d.Spec.Sku.Tier),
			Family: network.ExpressRouteCircuitSkuFamily(d.Spec.Sku.Family),
		},
		ExpressRouteCircuitPropertiesFormat: &networkmgmt.ExpressRouteCircuitPropertiesFormat{
			ServiceProviderProperties: &networkmgmt.ExpressRouteCircuitServiceProviderProperties{
				ServiceProviderName: d.Spec.ExpressRouteCircuitsPropertiesFormat.ServiceProviderName,
				BandwidthInMbps:     d.Spec.ExpressRouteCircuitsPropertiesFormat.BandwidthInMbps,
				PeeringLocation:     d.Spec.ExpressRouteCircuitsPropertiesFormat.PeeringLocation,
			},
			GlobalReachEnabled:     d.Spec.ExpressRouteCircuitsPropertiesFormat.GlobalReachEnabled,
			AllowClassicOperations: d.Spec.ExpressRouteCircuitsPropertiesFormat.AllowClassicOperations,
		},
	}
}

// UpdateExpressRouteCircuitStatusFromAzure updates the status related to the external
func UpdateExpressRouteCircuitStatusFromAzure(v *v1alpha3.ExpressRouteCircuits, az networkmgmt.ExpressRouteCircuit) {
	v.Status.State = azure.ToString(az.ExpressRouteCircuitPropertiesFormat.ProvisioningState)
	v.Status.Etag = azure.ToString(az.Etag)
	v.Status.ID = azure.ToString(az.ID)
}

// ExpressRouteCircuitNeedsUpdate determines if a express route circuit needs to be updated
func ExpressRouteCircuitNeedsUpdate(kube *v1alpha3.ExpressRouteCircuits, az networkmgmt.ExpressRouteCircuit) bool {
	up := NewExpressRouteCircuitsParameters(kube)
	switch {
	case !reflect.DeepEqual(up.ExpressRouteCircuitPropertiesFormat.AllowClassicOperations, az.ExpressRouteCircuitPropertiesFormat.AllowClassicOperations):
		return true
	case !reflect.DeepEqual(up.Sku.Tier, az.Sku.Tier):
		return true
	case !reflect.DeepEqual(up.Sku.Family, az.Sku.Family):
		return true
	case !reflect.DeepEqual(up.ExpressRouteCircuitPropertiesFormat.ServiceProviderProperties.BandwidthInMbps, az.ExpressRouteCircuitPropertiesFormat.ServiceProviderProperties.BandwidthInMbps):
		return true
	}
	return false
}
