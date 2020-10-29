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

// SecurityRuleProtocol enumerates the values for security rule protocol.
type SecurityRuleProtocol string

// ApplicationSecurityGroupPropertiesFormat application security group properties.
type ApplicationSecurityGroupPropertiesFormat struct {
	// ResourceGUID - READ-ONLY; The resource GUID property of the application security group resource. It uniquely identifies a resource, even if the user changes its name or migrate the resource across subscriptions or resource groups.
	ResourceGUID string `json:"resourceGuid,omitempty"`
	// ProvisioningState - READ-ONLY; The provisioning state of the application security group resource. Possible values are: 'Succeeded', 'Updating', 'Deleting', and 'Failed'.
	ProvisioningState string `json:"provisioningState,omitempty"`
}

// ApplicationSecurityGroup an application security group in a resource group.
type ApplicationSecurityGroup struct {
	// ApplicationSecurityGroupPropertiesFormat - Properties of the application security group.
	Properties ApplicationSecurityGroupPropertiesFormat  `json:"properties,omitempty"`
	// Etag - READ-ONLY; A unique read-only string that changes whenever the resource is updated.
	Etag string `json:"etag,omitempty"`
	// ID - Resource ID.
	ID string `json:"id,omitempty"`
	// Name - READ-ONLY; Resource name.
	Name string `json:"name,omitempty"`
	// Type - READ-ONLY; Resource type.
	Type string `json:"type,omitempty"`
	// Location - Resource location.
	Location string `json:"location,omitempty"`

}

// SecurityRuleAccess enumerates the values for security rule access.
type SecurityRuleAccess string

// SecurityRuleDirection enumerates the values for security rule direction.
type SecurityRuleDirection string

// SecurityRulePropertiesFormat security rule resource.
type SecurityRulePropertiesFormat struct {
	// Description - A description for this rule. Restricted to 140 chars.
	Description string `json:"description,omitempty"`
	// Protocol - Network protocol this rule applies to. Possible values include: 'SecurityRuleProtocolTCP', 'SecurityRuleProtocolUDP', 'SecurityRuleProtocolIcmp', 'SecurityRuleProtocolEsp', 'SecurityRuleProtocolAsterisk'
	Protocol SecurityRuleProtocol `json:"protocol,omitempty"`
	// SourcePortRange - The source port or range. Integer or range between 0 and 65535. Asterisk '*' can also be used to match all ports.
	SourcePortRange string `json:"sourcePortRange,omitempty"`
	// DestinationPortRange - The destination port or range. Integer or range between 0 and 65535. Asterisk '*' can also be used to match all ports.
	DestinationPortRange string `json:"destinationPortRange,omitempty"`
	// SourceAddressPrefix - The CIDR or source IP range. Asterisk '*' can also be used to match all source IPs. Default tags such as 'VirtualNetwork', 'AzureLoadBalancer' and 'Internet' can also be used. If this is an ingress rule, specifies where network traffic originates from.
	SourceAddressPrefix string `json:"sourceAddressPrefix,omitempty"`
	// SourceAddressPrefixes - The CIDR or source IP ranges.
	SourceAddressPrefixes []string `json:"sourceAddressPrefixes,omitempty"`
	// SourceApplicationSecurityGroups - The application security group specified as source.
	SourceApplicationSecurityGroups []ApplicationSecurityGroup `json:"sourceApplicationSecurityGroups,omitempty"`
	// DestinationAddressPrefix - The destination address prefix. CIDR or destination IP range. Asterisk '*' can also be used to match all source IPs. Default tags such as 'VirtualNetwork', 'AzureLoadBalancer' and 'Internet' can also be used.
	DestinationAddressPrefix string `json:"destinationAddressPrefix,omitempty"`
	// DestinationAddressPrefixes - The destination address prefixes. CIDR or destination IP ranges.
	DestinationAddressPrefixes []string `json:"destinationAddressPrefixes,omitempty"`
	// DestinationApplicationSecurityGroups - The application security group specified as destination.
	DestinationApplicationSecurityGroups []ApplicationSecurityGroup `json:"destinationApplicationSecurityGroups,omitempty"`
	// SourcePortRanges - The source port ranges.
	SourcePortRanges []string `json:"sourcePortRanges,omitempty"`
	// DestinationPortRanges - The destination port ranges.
	DestinationPortRanges []string `json:"destinationPortRanges,omitempty"`
	// Access - The network traffic is allowed or denied. Possible values include: 'SecurityRuleAccessAllow', 'SecurityRuleAccessDeny'
	Access SecurityRuleAccess `json:"access,omitempty"`
	// Priority - The priority of the rule. The value can be between 100 and 4096. The priority number must be unique for each rule in the collection. The lower the priority number, the higher the priority of the rule.
	Priority int32 `json:"priority,omitempty"`
	// Direction - The direction of the rule. The direction specifies if rule will be evaluated on incoming or outgoing traffic. Possible values include: 'SecurityRuleDirectionInbound', 'SecurityRuleDirectionOutbound'
	Direction SecurityRuleDirection `json:"direction,omitempty"`
	// ProvisioningState - The provisioning state of the public IP resource. Possible values are: 'Updating', 'Deleting', and 'Failed'.
	ProvisioningState string `json:"provisioningState,omitempty"`
}

// +kubebuilder:object:root=true
// SecurityRule network security rule.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.location"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type SecurityRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SecurityRulePropertiesFormat - Properties of the security rule.
	Properties SecurityRulePropertiesFormat `json:"properties,omitempty"`
	// Name - The name of the resource that is unique within a resource group. This name can be used to access the resource.
	Name string `json:"name,omitempty"`
	// Etag - A unique read-only string that changes whenever the resource is updated.
	Etag string `json:"etag,omitempty"`
	// ID - Resource ID.
	ID string `json:"id,omitempty"`
}


// A SecurityGroupSpec defines the desired state of a SecurityGroup.
type SecurityGroupSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// ResourceGroupName - Name of the SecurityGroup's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to the the SecurityGroup's resource
	// group.
	ResourceGroupNameRef *runtimev1alpha1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to the the Virtual
	// Network's resource group.
	ResourceGroupNameSelector *runtimev1alpha1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// Location - Resource location.
	Location string `json:"location"`

	//SecurityGroPropertiesFormat - Properties of security group
	SecurityGroupPropertiesFormat   `json:"properties,omitempty"`

	// Tags - Resource tags.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// A SecurityGroupStatus represents the observed status of a SecurityGroup.
type SecurityGroupStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// State of this SecurityGroup.
	State string `json:"state,omitempty"`

	// A Message providing detail about the state of this SecurityGroup, if
	// any.
	Message string `json:"message,omitempty"`

	// ID of this SecurityGroup.
	ID string `json:"id,omitempty"`

	// Etag - A unique read-only string that changes whenever the resource is
	// updated.
	Etag string `json:"etag,omitempty"`

	// ResourceGUID - The GUID of this SecurityGroup.
	ResourceGUID string `json:"resourceGuid,omitempty"`

	// Type of this SecurityGroup.
	Type string `json:"type,omitempty"`
}
// SecurityGroupPropertiesFormat network Security Group resource.
type SecurityGroupPropertiesFormat struct {
	// SecurityRules - A collection of security rules of the network security group.
	SecurityRules *[]SecurityRule `json:"securityRules,omitempty"`
	// DefaultSecurityRules - The default security rules of network security group.
	DefaultSecurityRules *[]SecurityRule `json:"defaultSecurityRules,omitempty"`
	// NetworkInterfaces - READ-ONLY; A collection of references to network interfaces.
	//NetworkInterfaces *[]Interface `json:"networkInterfaces,omitempty"`
	// Subnets - READ-ONLY; A collection of references to subnets.
	//Subnets *[]Subnet `json:"subnets,omitempty"`
	// ResourceGUID - The resource GUID property of the network security group resource.
	ResourceGUID *string `json:"resourceGuid,omitempty"`
	// ProvisioningState - The provisioning state of the public IP resource. Possible values are: 'Updating', 'Deleting', and 'Failed'.
	ProvisioningState *string `json:"provisioningState,omitempty"`
}
// +kubebuilder:object:root=true
// A SecurityGroup is a managed resource that represents an Azure Security
// Group.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.location"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type SecurityGroup struct{
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SecurityGroupSpec   `json:"spec"`
	Status SecurityGroupStatus `json:"status,omitempty"`
	///Properties SecurityGroupPropertiesFormat   `json:"properties,omitempty"`
}

// +kubebuilder:object:root=true
// SecurityGroupList contains a list of Security Groups
type SecurityGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityGroup `json:"items"`
}