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

type IOTHubDPSSharedAccessPolicyObservation struct {
	ID *string `json:"id,omitempty" tf:"id,omitempty"`
}

type IOTHubDPSSharedAccessPolicyParameters struct {

	// +kubebuilder:validation:Optional
	EnrollmentRead *bool `json:"enrollmentRead,omitempty" tf:"enrollment_read,omitempty"`

	// +kubebuilder:validation:Optional
	EnrollmentWrite *bool `json:"enrollmentWrite,omitempty" tf:"enrollment_write,omitempty"`

	// +crossplane:generate:reference:type=IOTHubDPS
	// +kubebuilder:validation:Optional
	IOTHubDPSName *string `json:"iothubDpsName,omitempty" tf:"iothub_dps_name,omitempty"`

	// +kubebuilder:validation:Optional
	IOTHubDPSNameRef *v1.Reference `json:"iotHubDpsNameRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	IOTHubDPSNameSelector *v1.Selector `json:"iotHubDpsNameSelector,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	RegistrationRead *bool `json:"registrationRead,omitempty" tf:"registration_read,omitempty"`

	// +kubebuilder:validation:Optional
	RegistrationWrite *bool `json:"registrationWrite,omitempty" tf:"registration_write,omitempty"`

	// +crossplane:generate:reference:type=github.com/crossplane/provider-azure/apis/azure/v1alpha2.ResourceGroup
	// +kubebuilder:validation:Optional
	ResourceGroupName *string `json:"resourceGroupName,omitempty" tf:"resource_group_name,omitempty"`

	// +kubebuilder:validation:Optional
	ResourceGroupNameRef *v1.Reference `json:"resourceGroupNameRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	ResourceGroupNameSelector *v1.Selector `json:"resourceGroupNameSelector,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	ServiceConfig *bool `json:"serviceConfig,omitempty" tf:"service_config,omitempty"`
}

// IOTHubDPSSharedAccessPolicySpec defines the desired state of IOTHubDPSSharedAccessPolicy
type IOTHubDPSSharedAccessPolicySpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     IOTHubDPSSharedAccessPolicyParameters `json:"forProvider"`
}

// IOTHubDPSSharedAccessPolicyStatus defines the observed state of IOTHubDPSSharedAccessPolicy.
type IOTHubDPSSharedAccessPolicyStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        IOTHubDPSSharedAccessPolicyObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// IOTHubDPSSharedAccessPolicy is the Schema for the IOTHubDPSSharedAccessPolicys API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azurejet}
type IOTHubDPSSharedAccessPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              IOTHubDPSSharedAccessPolicySpec   `json:"spec"`
	Status            IOTHubDPSSharedAccessPolicyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// IOTHubDPSSharedAccessPolicyList contains a list of IOTHubDPSSharedAccessPolicys
type IOTHubDPSSharedAccessPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []IOTHubDPSSharedAccessPolicy `json:"items"`
}

// Repository type metadata.
var (
	IOTHubDPSSharedAccessPolicy_Kind             = "IOTHubDPSSharedAccessPolicy"
	IOTHubDPSSharedAccessPolicy_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: IOTHubDPSSharedAccessPolicy_Kind}.String()
	IOTHubDPSSharedAccessPolicy_KindAPIVersion   = IOTHubDPSSharedAccessPolicy_Kind + "." + CRDGroupVersion.String()
	IOTHubDPSSharedAccessPolicy_GroupVersionKind = CRDGroupVersion.WithKind(IOTHubDPSSharedAccessPolicy_Kind)
)

func init() {
	SchemeBuilder.Register(&IOTHubDPSSharedAccessPolicy{}, &IOTHubDPSSharedAccessPolicyList{})
}
