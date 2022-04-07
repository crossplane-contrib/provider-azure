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
	"testing"

	networkmgmt "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	azure "github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients"

	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/network/v1alpha3"
)

var (
	uid                  = types.UID("definitely-a-uuid")
	location             = "cool-location"
	enableDDOSProtection = true
	enableVMProtection   = true
	addressPrefixes      = []string{"10.0.0.0/16"}
	addressPrefix        = "10.0.0.0/16"
	serviceEndpoint      = "Microsoft.Sql"
	tags                 = map[string]string{"one": "test", "two": "test"}

	id           = "a-very-cool-id"
	etag         = "a-very-cool-etag"
	resourceType = "resource-type"
	purpose      = "cool-purpose"
	address      = "20.46.134.23"

	ipVersion           = networkmgmt.IPv6
	skuName             = string(networkmgmt.PublicIPAddressSkuNameStandard)
	ipAllocMethod       = networkmgmt.Static
	prefixID            = "/test-prefix-id"
	dnsLabel            = "test-label"
	fqdn                = "test.fqdn"
	reverseFQDN         = "fqdn.reverse"
	timeout       int32 = 1
	ipTagType           = "FirstPartyUsage"
	ipTag               = "SQL"
	ipTagType2          = "FirstPartyUsage2"
	ipTag2              = "SQL2"
	tagKey              = "tagKey"
	tagVal              = "tagValue"
)

func TestNewVirtualNetworkParameters(t *testing.T) {
	cases := []struct {
		name string
		r    *v1alpha3.VirtualNetwork
		want networkmgmt.VirtualNetwork
	}{
		{
			name: "SuccessfulFull",
			r: &v1alpha3.VirtualNetwork{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.VirtualNetworkSpec{
					Location: location,
					VirtualNetworkPropertiesFormat: v1alpha3.VirtualNetworkPropertiesFormat{
						AddressSpace: v1alpha3.AddressSpace{
							AddressPrefixes: addressPrefixes,
						},
						EnableDDOSProtection: enableDDOSProtection,
						EnableVMProtection:   enableVMProtection,
					},
				},
			},
			want: networkmgmt.VirtualNetwork{
				Location: azure.ToStringPtr(location),
				Tags:     azure.ToStringPtrMap(nil),
				VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
					EnableDdosProtection: to.BoolPtr(enableDDOSProtection),
					EnableVMProtection:   to.BoolPtr(enableVMProtection),
					AddressSpace: &networkmgmt.AddressSpace{
						AddressPrefixes: &addressPrefixes,
					},
				},
			},
		},
		{
			name: "SuccessfulPartial",
			r: &v1alpha3.VirtualNetwork{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.VirtualNetworkSpec{
					Location: location,
					VirtualNetworkPropertiesFormat: v1alpha3.VirtualNetworkPropertiesFormat{
						AddressSpace: v1alpha3.AddressSpace{
							AddressPrefixes: addressPrefixes,
						},
						EnableDDOSProtection: enableDDOSProtection,
					},
				},
			},
			want: networkmgmt.VirtualNetwork{
				Location: azure.ToStringPtr(location),
				Tags:     azure.ToStringPtrMap(nil),
				VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
					EnableDdosProtection: to.BoolPtr(enableDDOSProtection),
					EnableVMProtection:   nil,
					AddressSpace: &networkmgmt.AddressSpace{
						AddressPrefixes: &addressPrefixes,
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewVirtualNetworkParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewVirtualNetworkParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestVirtualNetworkNeedsUpdate(t *testing.T) {
	cases := []struct {
		name string
		kube *v1alpha3.VirtualNetwork
		az   networkmgmt.VirtualNetwork
		want bool
	}{
		{
			name: "NeedsUpdateAddressSpace",
			kube: &v1alpha3.VirtualNetwork{
				Spec: v1alpha3.VirtualNetworkSpec{
					VirtualNetworkPropertiesFormat: v1alpha3.VirtualNetworkPropertiesFormat{
						AddressSpace: v1alpha3.AddressSpace{
							AddressPrefixes: []string{"10.3.0.0/16"},
						},
						EnableDDOSProtection: enableDDOSProtection,
						EnableVMProtection:   enableVMProtection,
					},
					Tags: tags,
				},
			},
			az: networkmgmt.VirtualNetwork{
				VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
					AddressSpace: &networkmgmt.AddressSpace{
						AddressPrefixes: &addressPrefixes,
					},
					EnableDdosProtection: to.BoolPtr(enableDDOSProtection),
					EnableVMProtection:   to.BoolPtr(enableVMProtection),
				},
				Tags: azure.ToStringPtrMap(tags),
			},
			want: true,
		},
		{
			name: "NeedsUpdateDdosProtection",
			kube: &v1alpha3.VirtualNetwork{
				Spec: v1alpha3.VirtualNetworkSpec{
					VirtualNetworkPropertiesFormat: v1alpha3.VirtualNetworkPropertiesFormat{
						AddressSpace: v1alpha3.AddressSpace{
							AddressPrefixes: addressPrefixes,
						},
						EnableDDOSProtection: !enableDDOSProtection,
						EnableVMProtection:   enableVMProtection,
					},
					Tags: tags,
				},
			},
			az: networkmgmt.VirtualNetwork{
				VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
					AddressSpace: &networkmgmt.AddressSpace{
						AddressPrefixes: &addressPrefixes,
					},
					EnableDdosProtection: to.BoolPtr(enableDDOSProtection),
					EnableVMProtection:   to.BoolPtr(enableVMProtection),
				},
				Tags: azure.ToStringPtrMap(tags),
			},
			want: true,
		},
		{
			name: "NeedsUpdateVMProtection",
			kube: &v1alpha3.VirtualNetwork{
				Spec: v1alpha3.VirtualNetworkSpec{
					VirtualNetworkPropertiesFormat: v1alpha3.VirtualNetworkPropertiesFormat{
						AddressSpace: v1alpha3.AddressSpace{
							AddressPrefixes: addressPrefixes,
						},
						EnableDDOSProtection: enableDDOSProtection,
						EnableVMProtection:   !enableVMProtection,
					},
					Tags: tags,
				},
			},
			az: networkmgmt.VirtualNetwork{
				VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
					AddressSpace: &networkmgmt.AddressSpace{
						AddressPrefixes: &addressPrefixes,
					},
					EnableDdosProtection: to.BoolPtr(enableDDOSProtection),
					EnableVMProtection:   to.BoolPtr(enableVMProtection),
				},
				Tags: azure.ToStringPtrMap(tags),
			},
			want: true,
		},
		{
			name: "NeedsUpdateTags",
			kube: &v1alpha3.VirtualNetwork{
				Spec: v1alpha3.VirtualNetworkSpec{
					VirtualNetworkPropertiesFormat: v1alpha3.VirtualNetworkPropertiesFormat{
						AddressSpace: v1alpha3.AddressSpace{
							AddressPrefixes: addressPrefixes,
						},
						EnableDDOSProtection: enableDDOSProtection,
						EnableVMProtection:   enableVMProtection,
					},
					Tags: map[string]string{"three": "test"},
				},
			},
			az: networkmgmt.VirtualNetwork{
				VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
					AddressSpace: &networkmgmt.AddressSpace{
						AddressPrefixes: &addressPrefixes,
					},
					EnableDdosProtection: to.BoolPtr(enableDDOSProtection),
					EnableVMProtection:   to.BoolPtr(enableVMProtection),
				},
				Tags: azure.ToStringPtrMap(tags),
			},
			want: true,
		},
		{
			name: "NoUpdate",
			kube: &v1alpha3.VirtualNetwork{
				Spec: v1alpha3.VirtualNetworkSpec{
					VirtualNetworkPropertiesFormat: v1alpha3.VirtualNetworkPropertiesFormat{
						AddressSpace: v1alpha3.AddressSpace{
							AddressPrefixes: addressPrefixes,
						},
						EnableDDOSProtection: enableDDOSProtection,
						EnableVMProtection:   enableVMProtection,
					},
					Tags: tags,
				},
			},
			az: networkmgmt.VirtualNetwork{
				VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
					AddressSpace: &networkmgmt.AddressSpace{
						AddressPrefixes: &addressPrefixes,
					},
					EnableDdosProtection: to.BoolPtr(enableDDOSProtection),
					EnableVMProtection:   to.BoolPtr(enableVMProtection),
				},
				Tags: azure.ToStringPtrMap(tags),
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := VirtualNetworkNeedsUpdate(tc.kube, tc.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("VirtualNetworkNeedsUpdate(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestUpdateVirtualNetworkStatusFromAzure(t *testing.T) {
	mockCondition := xpv1.Condition{Message: "mockMessage"}
	resourceStatus := xpv1.ResourceStatus{
		ConditionedStatus: xpv1.ConditionedStatus{
			Conditions: []xpv1.Condition{mockCondition},
		},
	}

	cases := []struct {
		name string
		r    networkmgmt.VirtualNetwork
		want v1alpha3.VirtualNetworkStatus
	}{
		{
			name: "SuccessfulFull",
			r: networkmgmt.VirtualNetwork{
				Location: azure.ToStringPtr(location),
				Etag:     azure.ToStringPtr(etag),
				ID:       azure.ToStringPtr(id),
				Type:     azure.ToStringPtr(resourceType),
				Tags:     azure.ToStringPtrMap(nil),
				VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
					EnableDdosProtection: azure.ToBoolPtr(enableDDOSProtection),
					EnableVMProtection:   azure.ToBoolPtr(enableVMProtection),
					AddressSpace: &networkmgmt.AddressSpace{
						AddressPrefixes: &addressPrefixes,
					},
					ProvisioningState: azure.ToStringPtr("Succeeded"),
					ResourceGUID:      azure.ToStringPtr(string(uid)),
				},
			},
			want: v1alpha3.VirtualNetworkStatus{
				State:        string(networkmgmt.Succeeded),
				ID:           id,
				Etag:         etag,
				Type:         resourceType,
				ResourceGUID: string(uid),
			},
		},
		{
			name: "SuccessfulPartial",
			r: networkmgmt.VirtualNetwork{
				Location: azure.ToStringPtr(location),
				Type:     azure.ToStringPtr(resourceType),
				Tags:     azure.ToStringPtrMap(nil),
				VirtualNetworkPropertiesFormat: &networkmgmt.VirtualNetworkPropertiesFormat{
					EnableDdosProtection: azure.ToBoolPtr(enableDDOSProtection),
					EnableVMProtection:   azure.ToBoolPtr(enableVMProtection),
					AddressSpace: &networkmgmt.AddressSpace{
						AddressPrefixes: &addressPrefixes,
					},
					ProvisioningState: azure.ToStringPtr("Succeeded"),
					ResourceGUID:      azure.ToStringPtr(string(uid)),
				},
			},
			want: v1alpha3.VirtualNetworkStatus{
				State:        string(networkmgmt.Succeeded),
				ResourceGUID: string(uid),
				Type:         resourceType,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			v := &v1alpha3.VirtualNetwork{
				Status: v1alpha3.VirtualNetworkStatus{
					ResourceStatus: resourceStatus,
				},
			}

			UpdateVirtualNetworkStatusFromAzure(v, tc.r)

			// make sure that internal resource status hasn't changed
			if diff := cmp.Diff(mockCondition, v.Status.ResourceStatus.Conditions[0]); diff != "" {
				t.Errorf("UpdateVirtualNetworkStatusFromAzure(...): -want, +got\n%s", diff)
			}

			// make sure that other resource parameters are updated
			tc.want.ResourceStatus = resourceStatus
			if diff := cmp.Diff(tc.want, v.Status); diff != "" {
				t.Errorf("UpdateVirtualNetworkStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestNewSubnetParameters(t *testing.T) {
	cases := []struct {
		name string
		r    *v1alpha3.Subnet
		want networkmgmt.Subnet
	}{
		{
			name: "Successful",
			r: &v1alpha3.Subnet{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.SubnetSpec{
					SubnetPropertiesFormat: v1alpha3.SubnetPropertiesFormat{
						AddressPrefix: addressPrefix,
					},
				},
			},
			want: networkmgmt.Subnet{
				SubnetPropertiesFormat: &networkmgmt.SubnetPropertiesFormat{
					AddressPrefix:    azure.ToStringPtr(addressPrefix),
					ServiceEndpoints: NewServiceEndpoints(nil),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewSubnetParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewSubnetParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestNewPublicIPAddressParameters(t *testing.T) {
	cases := []struct {
		name string
		r    *v1alpha3.PublicIPAddress
		want networkmgmt.PublicIPAddress
	}{
		{
			name: "Successful",
			r: &v1alpha3.PublicIPAddress{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.PublicIPAddressSpec{
					ForProvider: v1alpha3.PublicIPAddressProperties{
						PublicIPAllocationMethod: string(ipAllocMethod),
						PublicIPAddressVersion:   string(ipVersion),
						Location:                 location,
						SKU: &v1alpha3.SKU{
							Name: skuName,
						},
						PublicIPPrefixID: &prefixID,
						PublicIPAddressDNSSettings: &v1alpha3.PublicIPAddressDNSSettings{
							DomainNameLabel: dnsLabel,
							ReverseFQDN:     &reverseFQDN,
						},
						TCPIdleTimeoutInMinutes: &timeout,
						IPTags: []v1alpha3.IPTag{
							{
								IPTagType: ipTagType,
								Tag:       ipTag,
							},
						},
						Tags: map[string]string{
							tagKey: tagVal,
						},
					},
				},
			},
			want: networkmgmt.PublicIPAddress{
				PublicIPAddressPropertiesFormat: &networkmgmt.PublicIPAddressPropertiesFormat{
					PublicIPAllocationMethod: ipAllocMethod,
					PublicIPAddressVersion:   ipVersion,
					PublicIPPrefix: &networkmgmt.SubResource{
						ID: &prefixID,
					},
					DNSSettings: &networkmgmt.PublicIPAddressDNSSettings{
						DomainNameLabel: &dnsLabel,
						ReverseFqdn:     &reverseFQDN,
					},
					IdleTimeoutInMinutes: &timeout,
					IPTags: &[]networkmgmt.IPTag{
						{
							IPTagType: &ipTagType,
							Tag:       &ipTag,
						},
					},
				},
				Location: &location,
				Sku: &networkmgmt.PublicIPAddressSku{
					Name: networkmgmt.PublicIPAddressSkuName(skuName),
				},
				Tags: map[string]*string{
					tagKey: &tagVal,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewPublicIPAddressParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewSubnetParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestNewServiceEndpoints(t *testing.T) {
	cases := []struct {
		name string
		r    []v1alpha3.ServiceEndpointPropertiesFormat
		want *[]networkmgmt.ServiceEndpointPropertiesFormat
	}{
		{
			name: "SuccessfulNotSet",
			r:    []v1alpha3.ServiceEndpointPropertiesFormat{},
			want: &[]networkmgmt.ServiceEndpointPropertiesFormat{},
		},
		{
			name: "SuccessfulSet",
			r: []v1alpha3.ServiceEndpointPropertiesFormat{
				{Service: serviceEndpoint},
			},
			want: &[]networkmgmt.ServiceEndpointPropertiesFormat{
				{Service: &serviceEndpoint},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewServiceEndpoints(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewServiceEndpoints(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestSubnetNeedsUpdate(t *testing.T) {
	cases := []struct {
		name string
		kube *v1alpha3.Subnet
		az   networkmgmt.Subnet
		want bool
	}{
		{
			name: "NeedsUpdate",
			kube: &v1alpha3.Subnet{
				Spec: v1alpha3.SubnetSpec{
					SubnetPropertiesFormat: v1alpha3.SubnetPropertiesFormat{
						AddressPrefix: "10.1.0.0/16",
					},
				},
			},
			az: networkmgmt.Subnet{
				SubnetPropertiesFormat: &networkmgmt.SubnetPropertiesFormat{
					AddressPrefix: &addressPrefix,
				},
			},
			want: true,
		},
		{
			name: "NoUpdate",
			kube: &v1alpha3.Subnet{
				Spec: v1alpha3.SubnetSpec{
					SubnetPropertiesFormat: v1alpha3.SubnetPropertiesFormat{
						AddressPrefix: addressPrefix,
					},
				},
			},
			az: networkmgmt.Subnet{
				SubnetPropertiesFormat: &networkmgmt.SubnetPropertiesFormat{
					AddressPrefix: &addressPrefix,
				},
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SubnetNeedsUpdate(tc.kube, tc.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("SubnetNeedsUpdate(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestUpdateSubnetStatusFromAzure(t *testing.T) {
	mockCondition := xpv1.Condition{Message: "mockMessage"}
	resourceStatus := xpv1.ResourceStatus{
		ConditionedStatus: xpv1.ConditionedStatus{
			Conditions: []xpv1.Condition{mockCondition},
		},
	}

	cases := []struct {
		name string
		r    networkmgmt.Subnet
		want v1alpha3.SubnetStatus
	}{
		{
			name: "SuccessfulFull",
			r: networkmgmt.Subnet{
				Etag: azure.ToStringPtr(etag),
				ID:   azure.ToStringPtr(id),
				SubnetPropertiesFormat: &networkmgmt.SubnetPropertiesFormat{
					Purpose:           azure.ToStringPtr(purpose),
					ProvisioningState: azure.ToStringPtr("Succeeded"),
				},
			},
			want: v1alpha3.SubnetStatus{
				State:   string(networkmgmt.Succeeded),
				ID:      id,
				Etag:    etag,
				Purpose: purpose,
			},
		},
		{
			name: "SuccessfulPartial",
			r: networkmgmt.Subnet{
				ID: azure.ToStringPtr(id),
				SubnetPropertiesFormat: &networkmgmt.SubnetPropertiesFormat{
					ProvisioningState: azure.ToStringPtr("Succeeded"),
				},
			},
			want: v1alpha3.SubnetStatus{
				State: string(networkmgmt.Succeeded),
				ID:    id,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			v := &v1alpha3.Subnet{
				Status: v1alpha3.SubnetStatus{
					ResourceStatus: resourceStatus,
				},
			}

			UpdateSubnetStatusFromAzure(v, tc.r)

			// make sure that internal resource status hasn't changed
			if diff := cmp.Diff(mockCondition, v.Status.ResourceStatus.Conditions[0]); diff != "" {
				t.Errorf("UpdateSubnetStatusFromAzure(...): -want, +got\n%s", diff)
			}

			// make sure that other resource parameters are updated
			tc.want.ResourceStatus = resourceStatus
			if diff := cmp.Diff(tc.want, v.Status); diff != "" {
				t.Errorf("UpdateSubnetStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestUpdatePublicIPAddressStatusFromAzure(t *testing.T) {
	mockCondition := xpv1.Condition{Message: "mockMessage"}
	resourceStatus := xpv1.ResourceStatus{
		ConditionedStatus: xpv1.ConditionedStatus{
			Conditions: []xpv1.Condition{mockCondition},
		},
	}

	cases := []struct {
		name string
		r    networkmgmt.PublicIPAddress
		want v1alpha3.PublicIPAddressStatus
	}{
		{
			name: "SuccessfulFull",
			r: networkmgmt.PublicIPAddress{
				Etag: azure.ToStringPtr(etag),
				ID:   azure.ToStringPtr(id),
				PublicIPAddressPropertiesFormat: &networkmgmt.PublicIPAddressPropertiesFormat{
					IPAddress:         azure.ToStringPtr(address),
					ProvisioningState: azure.ToStringPtr("Succeeded"),
				},
			},
			want: v1alpha3.PublicIPAddressStatus{
				AtProvider: v1alpha3.PublicIPAddressObservation{
					State:   string(networkmgmt.Succeeded),
					ID:      id,
					Etag:    etag,
					Address: address,
				},
			},
		},
		{
			name: "SuccessfulPartial",
			r: networkmgmt.PublicIPAddress{
				ID: azure.ToStringPtr(id),
				PublicIPAddressPropertiesFormat: &networkmgmt.PublicIPAddressPropertiesFormat{
					ProvisioningState: azure.ToStringPtr("Succeeded"),
				},
			},
			want: v1alpha3.PublicIPAddressStatus{
				AtProvider: v1alpha3.PublicIPAddressObservation{
					State: string(networkmgmt.Succeeded),
					ID:    id,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			v := &v1alpha3.PublicIPAddress{
				Status: v1alpha3.PublicIPAddressStatus{
					ResourceStatus: resourceStatus,
				},
			}

			v.Status.AtProvider = *GeneratePublicIPAddressObservation(tc.r)

			// make sure that internal resource status hasn't changed
			if diff := cmp.Diff(mockCondition, v.Status.ResourceStatus.Conditions[0]); diff != "" {
				t.Errorf("UpdateSubnetStatusFromAzure(...): -want, +got\n%s", diff)
			}

			// make sure that other resource parameters are updated
			tc.want.ResourceStatus = resourceStatus
			if diff := cmp.Diff(tc.want, v.Status); diff != "" {
				t.Errorf("UpdateSubnetStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestIsPublicIPAddressUpToDate(t *testing.T) {
	tagValue2 := "value2"
	type args struct {
		p  v1alpha3.PublicIPAddressProperties
		in networkmgmt.PublicIPAddress
	}
	tests := map[string]struct {
		args args
		want bool
	}{
		"NoUpdatesBothEmpty": {
			args: args{
				in: newSDKPublicIPAddress(),
			},
			want: true,
		},
		"NoUpdatesSpecEmptyTags": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					Tags: map[string]string{},
				},
				in: newSDKPublicIPAddress(),
			},
			want: true,
		},
		"NoUpdatesSDKEmptyTags": {
			args: args{
				in: newSDKPublicIPAddress(withTags(map[string]*string{})),
			},
			want: true,
		},
		"SameNonEmptyTags": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					Tags: map[string]string{tagKey: tagVal},
				},
				in: newSDKPublicIPAddress(withTags(map[string]*string{tagKey: &tagVal})),
			},
			want: true,
		},
		"DifferingTags": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					Tags: map[string]string{tagKey: tagVal},
				},
				in: newSDKPublicIPAddress(withTags(map[string]*string{tagKey: &tagValue2})),
			},
		},
		"SameNonEmptyIPTagsInOrder": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					IPTags: []v1alpha3.IPTag{
						{
							IPTagType: ipTagType2,
							Tag:       ipTag2,
						},
						{
							IPTagType: ipTagType,
							Tag:       ipTag,
						},
					},
				},
				in: newSDKPublicIPAddress(withIPTags([]networkmgmt.IPTag{
					{
						IPTagType: &ipTagType2,
						Tag:       &ipTag2,
					},
					{
						IPTagType: &ipTagType,
						Tag:       &ipTag,
					},
				})),
			},
			want: true,
		},
		"SameNonEmptyIPTagsSpecNotInOrder": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					IPTags: []v1alpha3.IPTag{
						{
							IPTagType: ipTagType,
							Tag:       ipTag,
						},
						{
							IPTagType: ipTagType2,
							Tag:       ipTag2,
						},
					},
				},
				in: newSDKPublicIPAddress(withIPTags([]networkmgmt.IPTag{
					{
						IPTagType: &ipTagType2,
						Tag:       &ipTag2,
					},
					{
						IPTagType: &ipTagType,
						Tag:       &ipTag,
					},
				})),
			},
			want: true,
		},
		"DifferingNonEmptyIPTags": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					IPTags: []v1alpha3.IPTag{
						{
							IPTagType: ipTagType,
							Tag:       ipTag,
						},
						{
							IPTagType: ipTagType2,
							Tag:       ipTag2,
						},
						{
							IPTagType: ipTagType2,
							Tag:       ipTag2,
						},
					},
				},
				in: newSDKPublicIPAddress(withIPTags([]networkmgmt.IPTag{
					{
						IPTagType: &ipTagType,
						Tag:       &ipTag,
					},
					{
						IPTagType: &ipTagType2,
						Tag:       &ipTag2,
					},
				})),
			},
			want: false,
		},
		"SameKeyIPTags": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					IPTags: []v1alpha3.IPTag{
						{
							IPTagType: ipTagType,
							Tag:       ipTag,
						},
						{
							IPTagType: ipTagType,
							Tag:       ipTag2,
						},
					},
				},
				in: newSDKPublicIPAddress(withIPTags([]networkmgmt.IPTag{
					{
						IPTagType: &ipTagType,
						Tag:       &ipTag2,
					},
					{
						IPTagType: &ipTagType,
						Tag:       &ipTag,
					},
				})),
			},
			want: true,
		},
		"SameIdleTimeout": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					TCPIdleTimeoutInMinutes: &timeout,
				},
				in: newSDKPublicIPAddress(withIdleTimeout(timeout)),
			},
			want: true,
		},
		"DifferingTimeout": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					TCPIdleTimeoutInMinutes: &timeout,
				},
				in: newSDKPublicIPAddress(withIdleTimeout(2)),
			},
			want: false,
		},
		"SameDNSSettings": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					PublicIPAddressDNSSettings: &v1alpha3.PublicIPAddressDNSSettings{
						DomainNameLabel: dnsLabel,
						ReverseFQDN:     &reverseFQDN,
					},
				},
				in: newSDKPublicIPAddress(withDNSSettings(dnsLabel, fqdn, reverseFQDN)),
			},
			want: true,
		},
		"DifferingDNSSettings": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					PublicIPAddressDNSSettings: &v1alpha3.PublicIPAddressDNSSettings{
						DomainNameLabel: dnsLabel,
						ReverseFQDN:     &reverseFQDN,
					},
				},
				in: newSDKPublicIPAddress(withDNSSettings("test-label2", fqdn, reverseFQDN)),
			},
			want: false,
		},
		"SameIPPrefixID": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					PublicIPPrefixID: &prefixID,
				},
				in: newSDKPublicIPAddress(withIPPrefix(prefixID)),
			},
			want: true,
		},
		"DifferingIPPrefixID": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					PublicIPPrefixID: &prefixID,
				},
				in: newSDKPublicIPAddress(withIPPrefix(id)),
			},
			want: false,
		},
		"SameSKU": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					SKU: &v1alpha3.SKU{
						Name: skuName,
					},
				},
				in: newSDKPublicIPAddress(withSKU(skuName)),
			},
			want: true,
		},
		"DifferingSKU": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					SKU: &v1alpha3.SKU{
						Name: skuName,
					},
				},
				in: newSDKPublicIPAddress(withSKU("another-sku-basic")),
			},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			if got := IsPublicIPAddressUpToDate(tt.args.p, tt.args.in); got != tt.want {
				t.Errorf("IsPublicIPAddressUpToDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

type publicIPAddressOption func(*networkmgmt.PublicIPAddress)

func newSDKPublicIPAddress(opts ...publicIPAddressOption) networkmgmt.PublicIPAddress {
	result := networkmgmt.PublicIPAddress{
		PublicIPAddressPropertiesFormat: &networkmgmt.PublicIPAddressPropertiesFormat{},
	}
	for _, o := range opts {
		o(&result)
	}
	return result
}

func withTags(tags map[string]*string) publicIPAddressOption {
	return func(in *networkmgmt.PublicIPAddress) {
		in.Tags = tags
	}
}

func withIPTags(ipTags []networkmgmt.IPTag) publicIPAddressOption {
	return func(in *networkmgmt.PublicIPAddress) {
		if ipTags == nil {
			return
		}
		in.IPTags = &ipTags
	}
}

func withIdleTimeout(to int32) publicIPAddressOption {
	return func(in *networkmgmt.PublicIPAddress) {
		in.IdleTimeoutInMinutes = &to
	}
}

func withIPPrefix(id string) publicIPAddressOption {
	return func(in *networkmgmt.PublicIPAddress) {
		in.PublicIPPrefix = &networkmgmt.SubResource{ID: &id}
	}
}

func withDNSSettings(label, fqdn, revFQDN string) publicIPAddressOption {
	return func(in *networkmgmt.PublicIPAddress) {
		in.DNSSettings = &networkmgmt.PublicIPAddressDNSSettings{
			DomainNameLabel: &label,
			Fqdn:            &fqdn,
			ReverseFqdn:     &revFQDN,
		}
	}
}

func withSKU(name string) publicIPAddressOption {
	return func(in *networkmgmt.PublicIPAddress) {
		in.Sku = &networkmgmt.PublicIPAddressSku{
			Name: networkmgmt.PublicIPAddressSkuName(name),
		}
	}
}

func TestLateInitializePublicIPAddress(t *testing.T) {
	tagVal2 := "tagValue2"
	type args struct {
		p  v1alpha3.PublicIPAddressProperties
		in networkmgmt.PublicIPAddress
	}
	tests := map[string]struct {
		args args
		want v1alpha3.PublicIPAddressProperties
	}{
		"LateInitializeEmptySpec": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{},
				in: newSDKPublicIPAddress(
					withTags(map[string]*string{tagKey: &tagVal}),
					withIPTags([]networkmgmt.IPTag{
						{
							IPTagType: &ipTagType,
							Tag:       &ipTag,
						},
					}),
					withIdleTimeout(timeout),
					withIPPrefix(prefixID),
					withDNSSettings(dnsLabel, fqdn, reverseFQDN),
					withSKU(skuName)),
			},
			want: v1alpha3.PublicIPAddressProperties{
				PublicIPPrefixID: &prefixID,
				PublicIPAddressDNSSettings: &v1alpha3.PublicIPAddressDNSSettings{
					DomainNameLabel: dnsLabel,
					ReverseFQDN:     &reverseFQDN,
				},
				TCPIdleTimeoutInMinutes: &timeout,
				IPTags: []v1alpha3.IPTag{
					{
						IPTagType: ipTagType,
						Tag:       ipTag,
					},
				},
				Tags: map[string]string{
					tagKey: tagVal,
				},
				SKU: &v1alpha3.SKU{
					Name: skuName,
				},
			},
		},
		"LateInitializeNonEmptySpec": {
			args: args{
				p: v1alpha3.PublicIPAddressProperties{
					// SKU:                        nil,
					PublicIPPrefixID: &prefixID,
					PublicIPAddressDNSSettings: &v1alpha3.PublicIPAddressDNSSettings{
						DomainNameLabel: dnsLabel,
						ReverseFQDN:     &reverseFQDN,
					},
					TCPIdleTimeoutInMinutes: &timeout,
					IPTags: []v1alpha3.IPTag{
						{
							IPTagType: ipTagType,
							Tag:       ipTag,
						},
					},
					Tags: map[string]string{
						tagKey: tagVal,
					},
					SKU: &v1alpha3.SKU{
						Name: skuName,
					},
				},
				in: newSDKPublicIPAddress(
					withTags(map[string]*string{"tagKey2": &tagVal2}),
					withIPTags([]networkmgmt.IPTag{
						{
							IPTagType: &ipTagType2,
							Tag:       &ipTag2,
						},
					}),
					withIdleTimeout(3),
					withIPPrefix(id),
					withDNSSettings("label2", "fqdn2", "fqdn.reverse2"),
					withSKU("another-sku-basic")),
			},
			want: v1alpha3.PublicIPAddressProperties{
				// SKU:                        nil,
				PublicIPPrefixID: &prefixID,
				PublicIPAddressDNSSettings: &v1alpha3.PublicIPAddressDNSSettings{
					DomainNameLabel: dnsLabel,
					ReverseFQDN:     &reverseFQDN,
				},
				TCPIdleTimeoutInMinutes: &timeout,
				IPTags: []v1alpha3.IPTag{
					{
						IPTagType: ipTagType,
						Tag:       ipTag,
					},
				},
				Tags: map[string]string{
					tagKey: tagVal,
				},
				SKU: &v1alpha3.SKU{
					Name: skuName,
				},
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			LateInitializePublicIPAddress(&tt.args.p, &tt.args.in)
			if diff := cmp.Diff(tt.want, tt.args.p); diff != "" {
				t.Errorf("LateInitializePublicIPAddress(tt.args.p, tt.args.in): -want, +got\n%s", diff)
			}
		})
	}
}
