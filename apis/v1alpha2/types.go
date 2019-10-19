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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
)

// A ProviderSpec defines the desired state of a Provider.
type ProviderSpec struct {
	// Azure service principal credentials json secret key reference
	// A Secret containing JSON encoded credentials for an Azure Service
	// Principal that will be used to authenticate to this Azure Provider.
	Secret runtimev1alpha1.SecretKeySelector `json:"credentialsSecretRef"`
}

// +kubebuilder:object:root=true

// A Provider configures an Azure 'provider', i.e. a connection to a particular
// Azure account using a particular Azure Service Principal.
// +kubebuilder:printcolumn:name="SECRET-NAME",type="string",JSONPath=".spec.credentialsSecretRef.name",priority=1
// +kubebuilder:resource:scope=Cluster
type Provider struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ProviderSpec `json:"spec,omitempty"`
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
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// Name of the resource group.
	Name string `json:"name,omitempty"`

	// Location of the resource group. See the  official list of valid regions -
	// https://azure.microsoft.com/en-us/global-infrastructure/regions/
	Location string `json:"location,omitempty"`
}

// A ResourceGroupStatus represents the observed status of a ResourceGroup.
type ResourceGroupStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// TODO(negz): Do we really need the name in both spec and status?

	// Name of the resource group.
	Name string `json:"name"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// A ResourceGroup is a managed resource that represents an Azure Resource
// Group.
type ResourceGroup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceGroupSpec   `json:"spec,omitempty"`
	Status ResourceGroupStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ResourceGroupList contains a list of Resource Groups
type ResourceGroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ResourceGroup `json:"items"`
}
