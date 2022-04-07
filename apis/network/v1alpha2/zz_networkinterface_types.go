/*
Copyright 2022 The Crossplane Authors.

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

// Code generated by terrajet. DO NOT EDIT.

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type IPConfigurationObservation struct {
}

type IPConfigurationParameters struct {

	// +kubebuilder:validation:Required
	Name *string `json:"name" tf:"name,omitempty"`

	// +kubebuilder:validation:Optional
	Primary *bool `json:"primary,omitempty" tf:"primary,omitempty"`

	// +kubebuilder:validation:Optional
	PrivateIPAddress *string `json:"privateIpAddress,omitempty" tf:"private_ip_address,omitempty"`

	// +kubebuilder:validation:Required
	PrivateIPAddressAllocation *string `json:"privateIpAddressAllocation" tf:"private_ip_address_allocation,omitempty"`

	// +kubebuilder:validation:Optional
	PrivateIPAddressVersion *string `json:"privateIpAddressVersion,omitempty" tf:"private_ip_address_version,omitempty"`

	// +kubebuilder:validation:Optional
	PublicIPAddressID *string `json:"publicIpAddressId,omitempty" tf:"public_ip_address_id,omitempty"`

	// +crossplane:generate:reference:type=Subnet
	// +crossplane:generate:reference:extractor=github.com/crossplane/provider-azure/apis/rconfig.ExtractResourceID()
	// +kubebuilder:validation:Optional
	SubnetID *string `json:"subnetId,omitempty" tf:"subnet_id,omitempty"`

	// +kubebuilder:validation:Optional
	SubnetIDRef *v1.Reference `json:"subnetIdRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	SubnetIDSelector *v1.Selector `json:"subnetIdSelector,omitempty" tf:"-"`
}

type NetworkInterfaceObservation struct {
	AppliedDNSServers []*string `json:"appliedDnsServers,omitempty" tf:"applied_dns_servers,omitempty"`

	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	InternalDomainNameSuffix *string `json:"internalDomainNameSuffix,omitempty" tf:"internal_domain_name_suffix,omitempty"`

	MacAddress *string `json:"macAddress,omitempty" tf:"mac_address,omitempty"`

	PrivateIPAddress *string `json:"privateIpAddress,omitempty" tf:"private_ip_address,omitempty"`

	PrivateIPAddresses []*string `json:"privateIpAddresses,omitempty" tf:"private_ip_addresses,omitempty"`

	VirtualMachineID *string `json:"virtualMachineId,omitempty" tf:"virtual_machine_id,omitempty"`
}

type NetworkInterfaceParameters struct {

	// +kubebuilder:validation:Optional
	DNSServers []*string `json:"dnsServers,omitempty" tf:"dns_servers,omitempty"`

	// +kubebuilder:validation:Optional
	EnableAcceleratedNetworking *bool `json:"enableAcceleratedNetworking,omitempty" tf:"enable_accelerated_networking,omitempty"`

	// +kubebuilder:validation:Optional
	EnableIPForwarding *bool `json:"enableIpForwarding,omitempty" tf:"enable_ip_forwarding,omitempty"`

	// +kubebuilder:validation:Required
	IPConfiguration []IPConfigurationParameters `json:"ipConfiguration" tf:"ip_configuration,omitempty"`

	// +kubebuilder:validation:Optional
	InternalDNSNameLabel *string `json:"internalDnsNameLabel,omitempty" tf:"internal_dns_name_label,omitempty"`

	// +kubebuilder:validation:Required
	Location *string `json:"location" tf:"location,omitempty"`

	// +crossplane:generate:reference:type=github.com/crossplane/provider-azure/apis/azure/v1alpha2.ResourceGroup
	// +kubebuilder:validation:Optional
	ResourceGroupName *string `json:"resourceGroupName,omitempty" tf:"resource_group_name,omitempty"`

	// +kubebuilder:validation:Optional
	ResourceGroupNameRef *v1.Reference `json:"resourceGroupNameRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	ResourceGroupNameSelector *v1.Selector `json:"resourceGroupNameSelector,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	Tags map[string]*string `json:"tags,omitempty" tf:"tags,omitempty"`
}

// NetworkInterfaceSpec defines the desired state of NetworkInterface
type NetworkInterfaceSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     NetworkInterfaceParameters `json:"forProvider"`
}

// NetworkInterfaceStatus defines the observed state of NetworkInterface.
type NetworkInterfaceStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        NetworkInterfaceObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkInterface is the Schema for the NetworkInterfaces API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azurejet}
type NetworkInterface struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              NetworkInterfaceSpec   `json:"spec"`
	Status            NetworkInterfaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkInterfaceList contains a list of NetworkInterfaces
type NetworkInterfaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkInterface `json:"items"`
}

// Repository type metadata.
var (
	NetworkInterface_Kind             = "NetworkInterface"
	NetworkInterface_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: NetworkInterface_Kind}.String()
	NetworkInterface_KindAPIVersion   = NetworkInterface_Kind + "." + CRDGroupVersion.String()
	NetworkInterface_GroupVersionKind = CRDGroupVersion.WithKind(NetworkInterface_Kind)
)

func init() {
	SchemeBuilder.Register(&NetworkInterface{}, &NetworkInterfaceList{})
}
