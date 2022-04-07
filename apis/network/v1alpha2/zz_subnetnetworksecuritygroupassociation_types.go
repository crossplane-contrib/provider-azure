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

type SubnetNetworkSecurityGroupAssociationObservation struct {
	ID *string `json:"id,omitempty" tf:"id,omitempty"`
}

type SubnetNetworkSecurityGroupAssociationParameters struct {

	// +kubebuilder:validation:Required
	NetworkSecurityGroupID *string `json:"networkSecurityGroupId" tf:"network_security_group_id,omitempty"`

	// +crossplane:generate:reference:type=Subnet
	// +crossplane:generate:reference:extractor=github.com/crossplane-contrib/provider-jet-azure/apis/rconfig.ExtractResourceID()
	// +kubebuilder:validation:Optional
	SubnetID *string `json:"subnetId,omitempty" tf:"subnet_id,omitempty"`

	// +kubebuilder:validation:Optional
	SubnetIDRef *v1.Reference `json:"subnetIdRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	SubnetIDSelector *v1.Selector `json:"subnetIdSelector,omitempty" tf:"-"`
}

// SubnetNetworkSecurityGroupAssociationSpec defines the desired state of SubnetNetworkSecurityGroupAssociation
type SubnetNetworkSecurityGroupAssociationSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     SubnetNetworkSecurityGroupAssociationParameters `json:"forProvider"`
}

// SubnetNetworkSecurityGroupAssociationStatus defines the observed state of SubnetNetworkSecurityGroupAssociation.
type SubnetNetworkSecurityGroupAssociationStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        SubnetNetworkSecurityGroupAssociationObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// SubnetNetworkSecurityGroupAssociation is the Schema for the SubnetNetworkSecurityGroupAssociations API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azurejet}
type SubnetNetworkSecurityGroupAssociation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              SubnetNetworkSecurityGroupAssociationSpec   `json:"spec"`
	Status            SubnetNetworkSecurityGroupAssociationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SubnetNetworkSecurityGroupAssociationList contains a list of SubnetNetworkSecurityGroupAssociations
type SubnetNetworkSecurityGroupAssociationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SubnetNetworkSecurityGroupAssociation `json:"items"`
}

// Repository type metadata.
var (
	SubnetNetworkSecurityGroupAssociation_Kind             = "SubnetNetworkSecurityGroupAssociation"
	SubnetNetworkSecurityGroupAssociation_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: SubnetNetworkSecurityGroupAssociation_Kind}.String()
	SubnetNetworkSecurityGroupAssociation_KindAPIVersion   = SubnetNetworkSecurityGroupAssociation_Kind + "." + CRDGroupVersion.String()
	SubnetNetworkSecurityGroupAssociation_GroupVersionKind = CRDGroupVersion.WithKind(SubnetNetworkSecurityGroupAssociation_Kind)
)

func init() {
	SchemeBuilder.Register(&SubnetNetworkSecurityGroupAssociation{}, &SubnetNetworkSecurityGroupAssociationList{})
}