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

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
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
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// ResourceGroupName - Name of the Virtual Network's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to the the Virtual Network's resource
	// group.
	ResourceGroupNameRef *runtimev1alpha1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to the the Virtual
	// Network's resource group.
	ResourceGroupNameSelector *runtimev1alpha1.Selector `json:"resourceGroupNameSelector,omitempty"`

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
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// State of this VirtualNetwork.
	State string `json:"state,omitempty"`

	// A Message providing detail about the state of this VirtualNetwork, if
	// any.
	Message string `json:"message,omitempty"`

	// ID of this VirtualNetwork.
	ID string `json:"id,omitempty"`

	// Etag - A unique string that changes whenever the resource is
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
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// VirtualNetworkName - Name of the Subnet's virtual network.
	VirtualNetworkName string `json:"virtualNetworkName,omitempty"`

	// VirtualNetworkNameRef references to a VirtualNetwork to retrieve its name
	VirtualNetworkNameRef *runtimev1alpha1.Reference `json:"virtualNetworkNameRef,omitempty"`

	// VirtualNetworkNameSelector selects a reference to a VirtualNetwork to
	// retrieve its name
	VirtualNetworkNameSelector *runtimev1alpha1.Selector `json:"virtualNetworkNameSelector,omitempty"`

	// ResourceGroupName - Name of the Subnet's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to the the Subnets's resource group.
	ResourceGroupNameRef *runtimev1alpha1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Selects a reference to the the Subnets's
	// resource group.
	ResourceGroupNameSelector *runtimev1alpha1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// SubnetPropertiesFormat - Properties of the subnet.
	SubnetPropertiesFormat `json:"properties"`
}

// A SubnetStatus represents the observed state of a Subnet.
type SubnetStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

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
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.location"
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

//Azure Firewall Structs
// +kubebuilder:object:root=true
// A AzureFirewall is a managed resource that represents an Azure Firewall
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.location"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type AzureFirewall struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureFirewallSpec   `json:"spec"`
	Status AzureFirewallStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true
// AzureFirewallList contains a list of Security Groups
type AzureFirewallList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureFirewall `json:"items"`
}

// A AzureFirewallSpec defines the desired state of a AzureFirewall.
type AzureFirewallSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// ResourceGroupName - Name of the SecurityGroup's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to the the SecurityGroup's resource
	// group.
	ResourceGroupNameRef *runtimev1alpha1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to the the Azure Firewall
	// resource group.
	ResourceGroupNameSelector *runtimev1alpha1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// Location - Resource location.
	Location string `json:"location"`

	// AzureFirewallPropertiesFormat - Properties of AzureFirewall
	AzureFirewallPropertiesFormat `json:"properties,omitempty"`

	// Tags - Resource tags.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`

	// Zones - A list of availability zones denoting where the resource needs to come from.
	Zones []string `json:"zones,omitempty"`

	// Etag - Gets a unique string that changes whenever the resource is updated.
	Etag string `json:"etag,omitempty"`

	// ID - Resource ID.
	ID string `json:"id,omitempty"`

	// Name - Resource name.
	Name string `json:"name,omitempty"`

	// Type - Resource type.
	Type string `json:"type,omitempty"`
}

// AzureFirewallPropertiesFormat properties of the Azure Firewall.
type AzureFirewallPropertiesFormat struct {
	// ApplicationRuleCollections - Collection of application rule collections used by Azure Firewall.
	ApplicationRuleCollections *[]AzureFirewallApplicationRuleCollection `json:"applicationRuleCollections,omitempty"`
	// NatRuleCollections - Collection of NAT rule collections used by Azure Firewall.
	NatRuleCollections *[]AzureFirewallNatRuleCollection `json:"natRuleCollections,omitempty"`
	// NetworkRuleCollections - Collection of network rule collections used by Azure Firewall.
	NetworkRuleCollections *[]AzureFirewallNetworkRuleCollection `json:"networkRuleCollections,omitempty"`
	// IPConfigurations - IP configuration of the Azure Firewall resource.
	IPConfigurations *[]AzureFirewallIPConfiguration `json:"ipConfigurations,omitempty"`
	// ProvisioningState - The provisioning state of the resource. Possible values include: 'Succeeded', 'Updating', 'Deleting', 'Failed'
	ProvisioningState string `json:"provisioningState,omitempty"`
	// ThreatIntelMode - The operation mode for Threat Intelligence. Possible values include: 'AzureFirewallThreatIntelModeAlert', 'AzureFirewallThreatIntelModeDeny', 'AzureFirewallThreatIntelModeOff'
	ThreatIntelMode string `json:"threatIntelMode,omitempty"`
	// VirtualHub - The virtualHub to which the firewall belongs.
	VirtualHub *SubResource `json:"virtualHub,omitempty"`
	// FirewallPolicy - The firewallPolicy associated with this azure firewall.
	FirewallPolicy *SubResource `json:"firewallPolicy,omitempty"`
	// HubIPAddresses - IP addresses associated with AzureFirewall.
	HubIPAddresses *HubIPAddresses `json:"hubIpAddresses,omitempty"`
}

// AzureFirewallApplicationRuleCollection application rule collection resource.
type AzureFirewallApplicationRuleCollection struct {
	// AzureFirewallApplicationRuleCollectionPropertiesFormat - Properties of the azure firewall application rule collection.
	Properties AzureFirewallApplicationRuleCollectionPropertiesFormat `json:"properties,omitempty"`
	// Name - Gets name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name string `json:"name,omitempty"`
	// Etag - READ-ONLY; Gets a unique read-only string that changes whenever the resource is updated.
	Etag string `json:"etag,omitempty"`
	// ID - Resource ID.
	ID string `json:"id,omitempty"`
}

// AzureFirewallApplicationRuleCollectionPropertiesFormat properties of the application rule collection.
type AzureFirewallApplicationRuleCollectionPropertiesFormat struct {
	// Priority - Priority of the application rule collection resource.
	Priority int32 `json:"priority,omitempty"`
	// Action - The action type of a rule collection.
	Action string `json:"action,omitempty"`
	// Rules - Collection of rules used by a application rule collection.
	Rules []AzureFirewallApplicationRule `json:"rules,omitempty"`
	// ProvisioningState - The provisioning state of the resource. Possible values include: 'Succeeded', 'Updating', 'Deleting', 'Failed'
	ProvisioningState string `json:"provisioningState,omitempty"`
}

// AzureFirewallApplicationRule properties of an application rule.
type AzureFirewallApplicationRule struct {
	// Name - Name of the application rule.
	Name string `json:"name,omitempty"`
	// Description - Description of the rule.
	Description string `json:"description,omitempty"`
	// SourceAddresses - List of source IP addresses for this rule.
	SourceAddresses []string `json:"sourceAddresses,omitempty"`
	// Protocols - Array of ApplicationRuleProtocols.
	Protocols []AzureFirewallApplicationRuleProtocol `json:"protocols,omitempty"`
	// TargetFqdns - List of FQDNs for this rule.
	TargetFqdns []string `json:"targetFqdns,omitempty"`
	// FqdnTags - List of FQDN Tags for this rule.
	FqdnTags []string `json:"fqdnTags,omitempty"`
}

// AzureFirewallApplicationRuleProtocol properties of the application rule protocol.
type AzureFirewallApplicationRuleProtocol struct {
	// ProtocolType - Protocol type. Possible values include: 'AzureFirewallApplicationRuleProtocolTypeHTTP', 'AzureFirewallApplicationRuleProtocolTypeHTTPS'
	ProtocolType string `json:"protocolType,omitempty"`
	// Port - Port number for the protocol, cannot be greater than 64000. This field is optional.
	Port int32 `json:"port,omitempty"`
}

// AzureFirewallIPConfiguration IP configuration of an Azure Firewall.
type AzureFirewallIPConfiguration struct {
	// AzureFirewallIPConfigurationPropertiesFormat - Properties of the azure firewall IP configuration.
	AzureFirewallIPConfigurationPropertiesFormat AzureFirewallIPConfigurationPropertiesFormat `json:"properties,omitempty"`
	// Name - Name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name *string `json:"name,omitempty"`
	// Etag - A unique string that changes whenever the resource is updated.
	Etag *string `json:"etag,omitempty"`
	// ID - Resource ID.
	ID *string `json:"id,omitempty"`
}

// AzureFirewallIPConfigurationPropertiesFormat properties of IP configuration of an Azure Firewall.
type AzureFirewallIPConfigurationPropertiesFormat struct {
	// PrivateIPAddress - The Firewall Internal Load Balancer IP to be used as the next hop in User Defined Routes.
	PrivateIPAddress *string `json:"privateIPAddress,omitempty"`
	// Subnet - Reference of the subnet resource. This resource must be named 'AzureFirewallSubnet'.
	Subnet *SubResource `json:"subnet,omitempty"`
	// PublicIPAddress - Reference of the PublicIP resource. This field is a mandatory input if subnet is not null.
	PublicIPAddress *SubResource `json:"publicIPAddress,omitempty"`
	// ProvisioningState - The provisioning state of the resource. Possible values include: 'Succeeded', 'Updating', 'Deleting', 'Failed'
	ProvisioningState *string `json:"provisioningState,omitempty"`
}

// SubResource reference to another subresource.
type SubResource struct {
	// ID - Resource ID.
	ID string `json:"id,omitempty"`
}

// HubIPAddresses IP addresses associated with azure firewall.
type HubIPAddresses struct {
	// PublicIPAddresses - List of Public IP addresses associated with azure firewall.
	PublicIPAddresses []AzureFirewallPublicIPAddress `json:"publicIPAddresses,omitempty"`
	// PrivateIPAddress - Private IP Address associated with azure firewall.
	PrivateIPAddress string `json:"privateIPAddress,omitempty"`
}

// AzureFirewallPublicIPAddress public IP Address associated with azure firewall.
type AzureFirewallPublicIPAddress struct {
	// Address - Public IP Address value.
	Address *string `json:"address,omitempty"`
}

// A AzureFirewallStatus represents the observed status of a AzureFirewall.
type AzureFirewallStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// State of this SecurityGroup.
	State string `json:"state,omitempty"`

	// A Message providing detail about the state of this AzureFirewall, if
	// any.
	Message string `json:"message,omitempty"`

	// ID of this AzureFirewall.
	ID string `json:"id,omitempty"`

	// Etag - A unique string that changes whenever the resource is
	// updated.
	Etag string `json:"etag,omitempty"`

	// ResourceGUID - The GUID of this AzureFirewall.
	ResourceGUID string `json:"resourceGuid,omitempty"`

	// Type of this AzureFirewall.
	Type string `json:"type,omitempty"`
}

//Rules Structs
// AzureFirewallNatRule properties of a NAT rule.
type AzureFirewallNatRule struct {
	// Name - Name of the NAT rule.
	Name string `json:"name,omitempty"`
	// Description - Description of the rule.
	Description string `json:"description,omitempty"`
	// SourceAddresses - List of source IP addresses for this rule.
	SourceAddresses []string `json:"sourceAddresses,omitempty"`
	// DestinationAddresses - List of destination IP addresses for this rule. Supports IP ranges, prefixes, and service tags.
	DestinationAddresses []string `json:"destinationAddresses,omitempty"`
	// DestinationPorts - List of destination ports.
	DestinationPorts []string `json:"destinationPorts,omitempty"`
	// Protocols - Array of AzureFirewallNetworkRuleProtocols applicable to this NAT rule.
	Protocols []string `json:"protocols,omitempty"`
	// TranslatedAddress - The translated address for this NAT rule.
	TranslatedAddress string `json:"translatedAddress,omitempty"`
	// TranslatedPort - The translated port for this NAT rule.
	TranslatedPort string `json:"translatedPort,omitempty"`
}

// AzureFirewallNatRuleCollectionProperties properties of the NAT rule collection.
type AzureFirewallNatRuleCollectionProperties struct {
	// Priority - Priority of the NAT rule collection resource.
	Priority int32 `json:"priority,omitempty"`
	// Action - The action type of a NAT rule collection.
	Action string `json:"action,omitempty"`
	// Rules - Collection of rules used by a NAT rule collection.
	Rules []AzureFirewallNatRule `json:"rules,omitempty"`
	// ProvisioningState - The provisioning state of the resource. Possible values include: 'Succeeded', 'Updating', 'Deleting', 'Failed'
	ProvisioningState string `json:"provisioningState,omitempty"`
}

// AzureFirewallNatRuleCollection NAT rule collection resource.
type AzureFirewallNatRuleCollection struct {
	// AzureFirewallNatRuleCollectionProperties - Properties of the azure firewall NAT rule collection.
	Properties AzureFirewallNatRuleCollectionProperties `json:"properties,omitempty"`
	// Name - Gets name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name string `json:"name,omitempty"`
	// Etag - Gets a unique string that changes whenever the resource is updated.
	Etag string `json:"etag,omitempty"`
	// ID - Resource ID.
	ID string `json:"id,omitempty"`
}

// AzureFirewallNetworkRuleCollection network rule collection resource.
type AzureFirewallNetworkRuleCollection struct {
	// AzureFirewallNetworkRuleCollectionPropertiesFormat - Properties of the azure firewall network rule collection.
	Properties AzureFirewallNetworkRuleCollectionPropertiesFormat `json:"properties,omitempty"`
	// Name - Gets name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name string `json:"name,omitempty"`
	// Etag - Gets a unique string that changes whenever the resource is updated.
	Etag string `json:"etag,omitempty"`
	// ID - Resource ID.
	ID string `json:"id,omitempty"`
}

// AzureFirewallNetworkRuleCollectionPropertiesFormat properties of the network rule collection.
type AzureFirewallNetworkRuleCollectionPropertiesFormat struct {
	// Priority - Priority of the network rule collection resource.
	Priority int32 `json:"priority,omitempty"`
	// Action - The action type of a rule collection.
	Action string `json:"action,omitempty"`
	// Rules - Collection of rules used by a network rule collection.
	Rules []AzureFirewallNetworkRule `json:"rules,omitempty"`
	// ProvisioningState - The provisioning state of the resource. Possible values include: 'Succeeded', 'Updating', 'Deleting', 'Failed'
	ProvisioningState string `json:"provisioningState,omitempty"`
}

// AzureFirewallNetworkRule properties of the network rule.
type AzureFirewallNetworkRule struct {
	// Name - Name of the network rule.
	Name string `json:"name,omitempty"`
	// Description - Description of the rule.
	Description string `json:"description,omitempty"`
	// Protocols - Array of AzureFirewallNetworkRuleProtocols.
	Protocols []string `json:"protocols,omitempty"`
	// SourceAddresses - List of source IP addresses for this rule.
	SourceAddresses []string `json:"sourceAddresses,omitempty"`
	// DestinationAddresses - List of destination IP addresses.
	DestinationAddresses []string `json:"destinationAddresses,omitempty"`
	// DestinationPorts - List of destination ports.
	DestinationPorts []string `json:"destinationPorts,omitempty"`
}
