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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// AddressSpace contains an array of IP address ranges that can be used by
// subnets of the virtual network.
type AddressSpace struct {
	// AddressPrefixes - A list of address blocks reserved for this virtual
	// network in CIDR notation.
	AddressPrefixes []string `json:"addressPrefixes"`
}

// VirtualNetworkPropertiesFormat defines properties of a VirtualNetwork.
type VirtualNetworkPropertiesFormat struct {
	// AddressSpace - The AddressSpace that contains an array of IP address
	// ranges that can be used by subnets.
	// +optional
	AddressSpace AddressSpace `json:"addressSpace"`

	// EnableDDOSProtection - Indicates if DDoS protection is enabled for all
	// the protected resources in the virtual network. It requires a DDoS
	// protection plan associated with the resource.
	// +optional
	EnableDDOSProtection bool `json:"enableDdosProtection,omitempty"`

	// EnableVMProtection - Indicates if VM protection is enabled for all the
	// subnets in the virtual network.
	// +optional
	EnableVMProtection bool `json:"enableVmProtection,omitempty"`
}

// A VirtualNetworkSpec defines the desired state of a VirtualNetwork.
type VirtualNetworkSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// ResourceGroupName - Name of the Virtual Network's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to the the Virtual Network's resource
	// group.
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to the the Virtual
	// Network's resource group.
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// VirtualNetworkPropertiesFormat - Properties of the virtual network.
	VirtualNetworkPropertiesFormat `json:"properties"`

	// Location - Resource location.
	Location string `json:"location"`

	// Tags - Resource tags.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// A VirtualNetworkStatus represents the observed state of a VirtualNetwork.
type VirtualNetworkStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// State of this VirtualNetwork.
	State string `json:"state,omitempty"`

	// A Message providing detail about the state of this VirtualNetwork, if
	// any.
	Message string `json:"message,omitempty"`

	// ID of this VirtualNetwork.
	ID string `json:"id,omitempty"`

	// Etag - A unique read-only string that changes whenever the resource is
	// updated.
	Etag string `json:"etag,omitempty"`

	// ResourceGUID - The GUID of this VirtualNetwork.
	ResourceGUID string `json:"resourceGuid,omitempty"`

	// Type of this VirtualNetwork.
	Type string `json:"type,omitempty"`
}

// +kubebuilder:object:root=true

// A VirtualNetwork is a managed resource that represents an Azure Virtual
// Network.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.location"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type VirtualNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualNetworkSpec   `json:"spec"`
	Status VirtualNetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualNetworkList contains a list of VirtualNetwork items
type VirtualNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualNetwork `json:"items"`
}

// ServiceEndpointPropertiesFormat defines properties of a service endpoint.
type ServiceEndpointPropertiesFormat struct {
	// Service - The type of the endpoint service.
	// +optional
	Service string `json:"service,omitempty"`

	// Locations - A list of locations.
	// +optional
	Locations []string `json:"locations,omitempty"`

	// ProvisioningState - The provisioning state of the resource.
	// +optional
	ProvisioningState string `json:"provisioningState,omitempty"`
}

// SubnetPropertiesFormat defines properties of a Subnet.
type SubnetPropertiesFormat struct {
	// AddressPrefix - The address prefix for the subnet.
	AddressPrefix string `json:"addressPrefix"`

	// ServiceEndpoints - An array of service endpoints.
	ServiceEndpoints []ServiceEndpointPropertiesFormat `json:"serviceEndpoints,omitempty"`
}

// A SubnetSpec defines the desired state of a Subnet.
type SubnetSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// VirtualNetworkName - Name of the Subnet's virtual network.
	VirtualNetworkName string `json:"virtualNetworkName,omitempty"`

	// VirtualNetworkNameRef references to a VirtualNetwork to retrieve its name
	VirtualNetworkNameRef *xpv1.Reference `json:"virtualNetworkNameRef,omitempty"`

	// VirtualNetworkNameSelector selects a reference to a VirtualNetwork to
	// retrieve its name
	VirtualNetworkNameSelector *xpv1.Selector `json:"virtualNetworkNameSelector,omitempty"`

	// ResourceGroupName - Name of the Subnet's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to the the Subnets's resource group.
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Selects a reference to the the Subnets's
	// resource group.
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// SubnetPropertiesFormat - Properties of the subnet.
	SubnetPropertiesFormat `json:"properties"`
}

// A SubnetStatus represents the observed state of a Subnet.
type SubnetStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// State of this Subnet.
	State string `json:"state,omitempty"`

	// A Message providing detail about the state of this Subnet, if any.
	Message string `json:"message,omitempty"`

	// Etag - A unique string that changes whenever the resource is updated.
	Etag string `json:"etag,omitempty"`

	// ID of this Subnet.
	ID string `json:"id,omitempty"`

	// Purpose - A string identifying the intention of use for this subnet based
	// on delegations and other user-defined properties.
	Purpose string `json:"purpose,omitempty"`
}

// +kubebuilder:object:root=true

// A Subnet is a managed resource that represents an Azure Subnet.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type Subnet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SubnetSpec   `json:"spec"`
	Status SubnetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubnetList contains a list of Subnet items
type SubnetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Subnet `json:"items"`
}

// A PublicIPAddressSpec defines the desired state of a PublicIPAddress.
type PublicIPAddressSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       PublicIPAddressProperties `json:"forProvider"`
}

// PublicIPAddressDNSSettings contains FQDN of the DNS record associated with the public IP address.
type PublicIPAddressDNSSettings struct {
	// DomainNameLabel -the Domain name label.
	// The concatenation of the domain name label and the regionalized DNS zone
	// make up the fully qualified domain name associated with
	// the public IP address. If a domain name label is specified,
	// an A DNS record is created for the public IP in
	// the Microsoft Azure DNS system.
	// +kubebuilder:validation:MinLength:=1
	DomainNameLabel string `json:"domainNameLabel"`
	// ReverseFQDN - Gets or Sets the Reverse FQDN.
	// A user-visible, fully qualified domain name that
	// resolves to this public IP address. If the reverseFqdn
	// is specified, then a PTR DNS record is created pointing
	// from the IP address in the in-addr.arpa domain to
	// the reverse FQDN.
	// +optional
	ReverseFQDN *string `json:"reverseFqdn,omitempty"`
}

// IPTag list of tags to be assigned to this public IP
type IPTag struct {
	// IPTagType - Type of the IP tag. Example: FirstPartyUsage.
	IPTagType string `json:"ipTagType"`
	// Tag - Value of the IpTag associated with the public IP. Example SQL, Storage etc.
	Tag string `json:"tag"`
}

// PublicIPAddressProperties defines properties of the PublicIPAddress.
type PublicIPAddressProperties struct {
	// ResourceGroupName - Name of the Public IP address's resource group.
	// +immutable
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to the the Public IP address's resource
	// group.
	// +immutable
	// +optional
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to the Public IP address's
	// resource group.
	// +optional
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// PublicIPAllocationMethod - The public IP address allocation method. Possible values include: 'Static', 'Dynamic'
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Static;Dynamic
	// +immutable
	PublicIPAllocationMethod string `json:"allocationMethod"`

	// PublicIPAllocationMethod - The public IP address version. Possible values include: 'IPv4', 'IPv6'
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=IPv4;IPv6
	// +immutable
	PublicIPAddressVersion string `json:"version"`

	// Location - Resource location.
	// +kubebuilder:validation:MinLength:=1
	// +immutable
	Location string `json:"location"`

	// SKU of PublicIPAddress
	// +optional
	SKU *SKU `json:"sku,omitempty"`

	// PublicIPPrefixID - The Public IP Prefix this Public IP Address should be allocated from.
	// +optional
	PublicIPPrefixID *string `json:"publicIPPrefixID,omitempty"`

	// PublicIPAddressDNSSettings - The FQDN of the DNS record associated with the public IP address.
	// +optional
	PublicIPAddressDNSSettings *PublicIPAddressDNSSettings `json:"dnsSettings,omitempty"`

	// TCPIdleTimeoutInMinutes - Timeout in minutes for idle TCP connections
	// +kubebuilder:validation:Minimum:=0
	// +optional
	TCPIdleTimeoutInMinutes *int32 `json:"tcpIdleTimeoutInMinutes,omitempty"`

	// IPTags - IP tags to be assigned to this public IP address
	// +optional
	IPTags []IPTag `json:"ipTags,omitempty"`

	// Tags - Resource tags.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// SKU of PublicIPAddress
type SKU struct {
	// Name - Name of sku. Possible values include: ['Standard', 'Basic']
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=Standard;Basic
	Name string `json:"name"`
}

// IPConfiguration properties of the observed IP configuration.
type IPConfiguration struct {
	// PrivateIPAddress - The private IP address of the IP configuration.
	PrivateIPAddress *string `json:"privateIPAddress,omitempty"`
	// PrivateIPAllocationMethod - The private IP address allocation method. Possible values include: 'Static', 'Dynamic'
	PrivateIPAllocationMethod string `json:"privateIPAllocationMethod"`
	// ProvisioningState - Gets the provisioning state of the public IP resource. Possible values are: 'Updating', 'Deleting', and 'Failed'.
	ProvisioningState string `json:"provisioningState"`
}

// PublicIPAddressDNSSettingsObservation represents observed DNS settings of
// a public IP resource
type PublicIPAddressDNSSettingsObservation struct {
	// DomainNameLabel -the Domain name label.
	// The concatenation of the domain name label and the regionalized DNS zone
	// make up the fully qualified domain name associated with
	// the public IP address. If a domain name label is specified,
	// an A DNS record is created for the public IP in
	// the Microsoft Azure DNS system.
	// +optional
	DomainNameLabel *string `json:"domainNameLabel,omitempty"`
	// ReverseFQDN - Gets or Sets the Reverse FQDN.
	// A user-visible, fully qualified domain name that
	// resolves to this public IP address. If the reverseFqdn
	// is specified, then a PTR DNS record is created pointing
	// from the IP address in the in-addr.arpa domain to
	// the reverse FQDN.
	// +optional
	ReverseFQDN *string `json:"reverseFqdn,omitempty"`
	// FQDN - Gets the FQDN, Fully qualified domain name of
	// the A DNS record associated with the public IP.
	// This is the concatenation of the domainNameLabel
	// and the regionalized DNS zone.
	FQDN *string `json:"fqdn,omitempty"`
}

// A PublicIPAddressObservation represents the observed state of a PublicIPAddress.
type PublicIPAddressObservation struct {
	// State of this PublicIPAddress.
	State string `json:"state,omitempty"`

	// A Message providing detail about the state of this PublicIPAddress, if any.
	Message string `json:"message,omitempty"`

	// Etag - A unique string that changes whenever the resource is updated.
	Etag string `json:"etag,omitempty"`

	// ID of this PublicIPAddress.
	ID string `json:"id,omitempty"`

	// Address - A string identifying address of PublicIPAddress resource
	Address string `json:"address"`

	// Version observed IP version
	Version string `json:"version"`

	// DNSSettings observed DNS settings of the IP address
	DNSSettings *PublicIPAddressDNSSettingsObservation `json:"dnsSettings,omitempty"`

	// IPConfiguration - The IP configuration associated with the public IP address
	IPConfiguration *IPConfiguration `json:"ipConfiguration,omitempty"`
}

// A PublicIPAddressStatus represents the observed state of a SQLServer.
type PublicIPAddressStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          PublicIPAddressObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A PublicIPAddress is a managed resource that represents an Azure PublicIPAddress.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ADDRESS",type="string",JSONPath=".status.atProvider.address"
// +kubebuilder:printcolumn:name="FQDN",type="string",JSONPath=".status.atProvider.dnsSettings.fqdn"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type PublicIPAddress struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PublicIPAddressSpec   `json:"spec"`
	Status PublicIPAddressStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PublicIPAddressList contains a list of PublicIPAddress items
type PublicIPAddressList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PublicIPAddress `json:"items"`
}
