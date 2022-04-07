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
	"sort"
	"strings"

	networkmgmt "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	azure "github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients"

	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/network/v1alpha3"
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
	p := s.Spec.ForProvider
	return networkmgmt.PublicIPAddress{
		Sku: NewPublicIPAddressSKU(s.Spec.ForProvider.SKU),
		PublicIPAddressPropertiesFormat: &networkmgmt.PublicIPAddressPropertiesFormat{
			PublicIPAllocationMethod: networkmgmt.IPAllocationMethod(p.PublicIPAllocationMethod),
			PublicIPAddressVersion:   networkmgmt.IPVersion(p.PublicIPAddressVersion),
			DNSSettings:              newDNSSettings(p.PublicIPAddressDNSSettings),
			PublicIPPrefix:           newPublicIPPrefixRef(p.PublicIPPrefixID),
			IdleTimeoutInMinutes:     p.TCPIdleTimeoutInMinutes,
			IPTags:                   newIPTags(p.IPTags),
		},
		Location: &p.Location,
		Tags:     azure.ToStringPtrMap(p.Tags),
	}
}

func newPublicIPPrefixRef(ref *string) *networkmgmt.SubResource {
	if ref == nil {
		return nil
	}
	return &networkmgmt.SubResource{
		ID: ref,
	}
}

func newIPTags(t []v1alpha3.IPTag) *[]networkmgmt.IPTag {
	if len(t) == 0 {
		return nil
	}
	result := make([]networkmgmt.IPTag, len(t))
	for i, tag := range t {
		tag := tag
		result[i] = networkmgmt.IPTag{
			IPTagType: &tag.IPTagType,
			Tag:       &tag.Tag,
		}
	}
	return &result
}

func newDNSSettings(s *v1alpha3.PublicIPAddressDNSSettings) *networkmgmt.PublicIPAddressDNSSettings {
	if s == nil {
		return nil
	}
	return &networkmgmt.PublicIPAddressDNSSettings{
		DomainNameLabel: &s.DomainNameLabel,
		ReverseFqdn:     s.ReverseFQDN,
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

// GeneratePublicIPAddressObservation returns the observation object related to the external
// Azure public IP address in the PublicIPAddressStatus
func GeneratePublicIPAddressObservation(az networkmgmt.PublicIPAddress) *v1alpha3.PublicIPAddressObservation {
	v := &v1alpha3.PublicIPAddressObservation{}
	v.State = azure.ToString(az.ProvisioningState)
	v.Etag = azure.ToString(az.Etag)
	v.ID = azure.ToString(az.ID)
	v.Address = azure.ToString(az.IPAddress)
	v.Version = string(az.PublicIPAddressVersion)
	if az.IPConfiguration != nil {
		v.IPConfiguration = &v1alpha3.IPConfiguration{
			PrivateIPAllocationMethod: string(az.IPConfiguration.PrivateIPAllocationMethod),
			PrivateIPAddress:          az.IPConfiguration.PrivateIPAddress,
			ProvisioningState:         azure.ToString(az.IPConfiguration.ProvisioningState),
		}
	}
	if az.DNSSettings != nil {
		v.DNSSettings = &v1alpha3.PublicIPAddressDNSSettingsObservation{
			DomainNameLabel: az.DNSSettings.DomainNameLabel,
			FQDN:            az.DNSSettings.Fqdn,
			ReverseFQDN:     az.DNSSettings.ReverseFqdn,
		}
	}
	return v
}

// LateInitializePublicIPAddress late-initilizes a PublicIPAddress resource
func LateInitializePublicIPAddress(p *v1alpha3.PublicIPAddressProperties, in *networkmgmt.PublicIPAddress) {
	p.PublicIPAddressDNSSettings = lateInitializeDNSSettings(p.PublicIPAddressDNSSettings, in.DNSSettings)
	p.Tags = azure.LateInitializeStringMap(p.Tags, in.Tags)
	if p.SKU == nil && in.Sku != nil {
		p.SKU = &v1alpha3.SKU{
			Name: string(in.Sku.Name),
		}
	}
	if p.PublicIPPrefixID == nil && in.PublicIPPrefix != nil && in.PublicIPPrefix.ID != nil {
		p.PublicIPPrefixID = in.PublicIPPrefix.ID
	}
	p.TCPIdleTimeoutInMinutes = azure.LateInitializeInt32PtrFromInt32Ptr(p.TCPIdleTimeoutInMinutes, in.IdleTimeoutInMinutes)
	p.IPTags = lateInitializeIPTags(p.IPTags, in.IPTags)
}

func lateInitializeIPTags(t []v1alpha3.IPTag, from *[]networkmgmt.IPTag) []v1alpha3.IPTag {
	if len(t) != 0 || from == nil || len(*from) == 0 {
		return t
	}
	t = make([]v1alpha3.IPTag, len(*from))
	for i, tag := range *from {
		tag := tag
		t[i] = v1alpha3.IPTag{
			IPTagType: *tag.IPTagType,
			Tag:       *tag.Tag,
		}
	}
	return t
}

func lateInitializeDNSSettings(d *v1alpha3.PublicIPAddressDNSSettings, in *networkmgmt.PublicIPAddressDNSSettings) *v1alpha3.PublicIPAddressDNSSettings {
	if in == nil {
		return d
	}
	if d == nil {
		d = &v1alpha3.PublicIPAddressDNSSettings{}
	}
	if d.DomainNameLabel == "" {
		d.DomainNameLabel = azure.ToString(in.DomainNameLabel)
	}
	d.ReverseFQDN = azure.LateInitializeStringPtrFromPtr(d.ReverseFQDN, in.ReverseFqdn)
	return d
}

// IsPublicIPAddressUpToDate is used to report whether given network.PublicIPAddress is in
// sync with the PublicIPAddressProperties that the user desires.
func IsPublicIPAddressUpToDate(p v1alpha3.PublicIPAddressProperties, in networkmgmt.PublicIPAddress) bool {
	if !cmp.Equal(p.Tags, azure.ToStringMap(in.Tags), cmpopts.EquateEmpty()) {
		return false
	}

	if !isIPPrefixIDUpToDate(p.PublicIPPrefixID, in.PublicIPPrefix) {
		return false
	}

	if azure.ToInt(p.TCPIdleTimeoutInMinutes) != azure.ToInt(in.IdleTimeoutInMinutes) {
		return false
	}

	if !isIPTagsUpToDate(p.IPTags, in.IPTags) {
		return false
	}

	if !isSKUUpToDate(p.SKU, in.Sku) {
		return false
	}

	return isDNSSettingsUpToDate(p.PublicIPAddressDNSSettings, in.PublicIPAddressPropertiesFormat.DNSSettings)
}

func isDNSSettingsUpToDate(d *v1alpha3.PublicIPAddressDNSSettings, in *networkmgmt.PublicIPAddressDNSSettings) bool {
	if d == nil {
		d = &v1alpha3.PublicIPAddressDNSSettings{}
	}
	if in == nil {
		in = &networkmgmt.PublicIPAddressDNSSettings{}
	}
	return d.DomainNameLabel == azure.ToString(in.DomainNameLabel) &&
		azure.ToString(d.ReverseFQDN) == azure.ToString(in.ReverseFqdn)
}

func isIPPrefixIDUpToDate(p *string, in *networkmgmt.SubResource) bool {
	if in == nil {
		in = &networkmgmt.SubResource{}
	}
	return azure.ToString(p) == azure.ToString(in.ID)
}

func isIPTagsUpToDate(t []v1alpha3.IPTag, in *[]networkmgmt.IPTag) bool {
	if in == nil {
		in = &[]networkmgmt.IPTag{}
	}
	if len(t) != len(*in) {
		return false
	}
	ct := make([]v1alpha3.IPTag, len(t))
	copy(ct, t)
	sort.Slice(ct, func(i, j int) bool {
		result := strings.Compare(ct[i].IPTagType, ct[j].IPTagType)
		if result != 0 {
			return result == -1
		}
		// then compare tags
		return strings.Compare(ct[i].Tag, ct[j].Tag) == -1
	})
	sort.Slice(*in, func(i, j int) bool {
		result := strings.Compare(*(*in)[i].IPTagType, *(*in)[j].IPTagType)
		if result != 0 {
			return result == -1
		}
		// then compare tags
		return strings.Compare(*(*in)[i].Tag, *(*in)[j].Tag) == -1
	})
	for i, tag := range *in {
		if *tag.IPTagType != ct[i].IPTagType || *tag.Tag != ct[i].Tag {
			return false
		}
	}
	return true
}

func isSKUUpToDate(s *v1alpha3.SKU, in *networkmgmt.PublicIPAddressSku) bool {
	if in == nil {
		in = &networkmgmt.PublicIPAddressSku{}
	}
	if s == nil {
		s = &v1alpha3.SKU{}
	}
	return s.Name == string(in.Name)
}
