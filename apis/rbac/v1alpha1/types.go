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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ApplicationParameters parameters for an application.
type ApplicationParameters struct {
	// AvailableToOtherTenants - Whether the application is available to other tenants.
	// +optional
	// +immutable
	AvailableToOtherTenants *bool `json:"availableToOtherTenants,omitempty"`

	// DisplayName - The display name of the application.
	// +optional
	// +immutable
	DisplayName *string `json:"displayName,omitempty"`

	// Homepage - The home page of the application.
	// +optional
	// +immutable
	Homepage *string `json:"homepage,omitempty"`

	// IdentifierUris - A collection of URIs for the application.
	// +optional
	// +immutable
	IdentifierURIs []string `json:"identifierUris,omitempty"`
}

// An ApplicationSpec defines the desired state of an Application.
type ApplicationSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ApplicationParameters `json:"forProvider"`
}

// An ApplicationStatus represents the observed state of an Application.
type ApplicationStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// ApplicationID - The application ID.
	ApplicationID string `json:"applicationID,omitempty"`
}

// +kubebuilder:object:root=true

// A Application is a managed resource that represents an Application.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type Application struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ApplicationSpec   `json:"spec"`
	Status ApplicationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ApplicationList contains a list of Application items
type ApplicationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Application `json:"items"`
}

// An ServicePrincipalParameters defines the desired state of an ServicePrincipal.
type ServicePrincipalParameters struct {
	// ApplicationID - The application ID.
	// +optional
	ApplicationID string `json:"applicationID,omitempty"`

	// ApplicationIDRef - A reference to the Application id.
	// +optional
	// +immutable
	ApplicationIDRef *xpv1.Reference `json:"applicationIDRef,omitempty"`

	// ApplicationIDSelector - Select a reference to the Application id.
	// +immutable
	ApplicationIDSelector *xpv1.Selector `json:"applicationIDSelector,omitempty"`

	// AccountEnabled - whether or not the service principal account is enabled
	// +optional
	// +immutable
	AccountEnabled *bool `json:"accountEnabled,omitempty"`
}

// An ServicePrincipalSpec defines the desired state of an ServicePrincipal.
type ServicePrincipalSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ServicePrincipalParameters `json:"forProvider"`
}

// An ServicePrincipalStatus represents the observed state of an ServicePrincipal.
type ServicePrincipalStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A ServicePrincipal is a managed resource that represents an ServicePrincipal.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type ServicePrincipal struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServicePrincipalSpec   `json:"spec"`
	Status ServicePrincipalStatus `json:"status,omitempty"`
}

// An RoleAssignmentParameters defines the desired state of an RoleAssignment.
type RoleAssignmentParameters struct {
	// PrincipalID - The principal ID assigned to the role.
	// This maps to the ID inside the Active Directory.
	// It can point to a user, service principal, or security group.
	// +optional
	PrincipalID string `json:"principalID,omitempty"`

	// PrincipalIDRef - A reference to the Principal id.
	// +optional
	// +immutable
	PrincipalIDRef *xpv1.Reference `json:"principalIDRef,omitempty"`

	// PrincipalIDRef - Select a reference to the Principal id.
	// +immutable
	PrincipalIDSelector *xpv1.Selector `json:"principalIDSelector,omitempty"`

	// RoleID - The role definition ID.
	// +immutable
	// +kubebuilder:validation:Required
	RoleID string `json:"roleID"`

	// Scope - The role assignment scope.
	// +immutable
	// +kubebuilder:validation:Required
	Scope string `json:"scope"`
}

// +kubebuilder:object:root=true

// ServicePrincipalList contains a list of ServicePrincipal items
type ServicePrincipalList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServicePrincipal `json:"items"`
}

// An RoleAssignmentSpec defines the desired state of an RoleAssignment.
type RoleAssignmentSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RoleAssignmentParameters `json:"forProvider"`
}

// An RoleAssignmentStatus represents the observed state of an RoleAssignment.
type RoleAssignmentStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A RoleAssignment is a managed resource that represents an RoleAssignment.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type RoleAssignment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RoleAssignmentSpec   `json:"spec"`
	Status RoleAssignmentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RoleAssignmentList contains a list of RoleAssignment items
type RoleAssignmentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RoleAssignment `json:"items"`
}
