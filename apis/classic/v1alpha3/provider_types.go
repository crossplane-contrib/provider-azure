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

// A ProvisioningState of a resource group.
type ProvisioningState string

// Provisioning states.
const (
	ProvisioningStateSucceeded ProvisioningState = "Succeeded"
	ProvisioningStateDeleting  ProvisioningState = "Deleting"
)

// A ProviderSpec defines the desired state of a Provider.
type ProviderSpec struct {
	// CredentialsSecretRef references a specific secret's key that contains
	// the credentials that are used to connect to the Azure API.
	CredentialsSecretRef xpv1.SecretKeySelector `json:"credentialsSecretRef"`
}

// +kubebuilder:object:root=true

// A Provider configures an Azure 'provider', i.e. a connection to a particular
// Azure account using a particular Azure Service Principal.
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentialsSecretRef.name",priority=1
// +kubebuilder:resource:scope=Cluster,categories={crossplane,provider,azure}
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProviderSpec `json:"spec"`
}

// +kubebuilder:object:root=true

// ProviderList contains a list of Provider
type ProviderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Provider `json:"items"`
}

// A ResourceGroupSpec defines the desired state of a ResourceGroup.
type ResourceGroupSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// Location of the resource group. See the  official list of valid regions -
	// https://azure.microsoft.com/en-us/global-infrastructure/regions/
	Location string `json:"location"`
}

// A ResourceGroupStatus represents the observed status of a ResourceGroup.
type ResourceGroupStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// ProvisioningState - The provisioning state of the resource group.
	ProvisioningState ProvisioningState `json:"provisioningState,omitempty"`
}

// A ResourceGroup is a managed resource that represents an Azure Resource
// Group.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type ResourceGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceGroupSpec   `json:"spec"`
	Status ResourceGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceGroupList contains a list of Resource Groups
type ResourceGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceGroup `json:"items"`
}

// AsyncOperation is used to save Azure Async operation details.
type AsyncOperation struct {
	// Method is HTTP method that the initial request is made with.
	Method string `json:"method,omitempty"`

	// PollingURL is used to fetch the status of the given operation.
	PollingURL string `json:"pollingUrl,omitempty"`

	// Status represents the status of the operation.
	Status string `json:"status,omitempty"`

	// ErrorMessage represents the error that occurred during the operation.
	ErrorMessage string `json:"errorMessage,omitempty"`
}
