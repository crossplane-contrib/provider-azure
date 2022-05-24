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

package v1alpha3

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// An RegistrySpec defines the desired state of a Registry.
type RegistrySpec struct {
	xpv1.ResourceSpec `json:",inline"`
	// ResourceGroupName is the name of the resource group that the cluster will
	// be created in
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup to retrieve its
	// name
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to a ResourceGroup to
	// retrieve its name
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// AdminUserEnabled - The value that indicates whether the admin user is enabled.
	AdminUserEnabled bool `json:"adminUserEnabled,omitempty"`

	// Sku - The SKU of the container registry.
	// +required
	Sku string `json:"sku"`

	// Location - The location of the resource. This cannot be changed after the resource is created.
	// +required
	Location string `json:"location"`
}

// An RegistryStatus represents the observed state of an Registry.
type RegistryStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// Status the status of an Azure resource at the time the operation was called.
	Status string `json:"status,omitempty"`

	// StatusMessage - The detailed message for the status, including alerts and error messages.
	StatusMessage string `json:"statusMessage,omitempty"`

	// State - The provisioning state of the container registry at the time the operation was called.
	// Possible values include: 'Creating', 'Updating', 'Deleting', 'Succeeded', 'Failed', 'Canceled'
	State string `json:"state,omitempty"`

	// ProviderID - The resource external ID.
	ProviderID string `json:"providerID,omitempty"`

	// LoginServer - The URL that can be used to log into the container registry.
	LoginServer string `json:"loginServer,omitempty"`
}

// +kubebuilder:object:root=true

// Registry an object that represents a container registry.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.location"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="URL",type="date",JSONPath=".status.loginServer"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
// +kubebuilder:subresource:status
type Registry struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RegistrySpec   `json:"spec"`
	Status RegistryStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RegistryList contains a list of Registry.
type RegistryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Registry `json:"items"`
}
