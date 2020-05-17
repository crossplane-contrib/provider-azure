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
	"github.com/Azure/azure-storage-blob-go/azblob"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

// AccountParameters define the desired state of an Azure Blob Storage Account.
type AccountParameters struct {
	// ResourceGroupName specifies the resource group for this Account.
	ResourceGroupName string `json:"resourceGroupName"`

	// StorageAccountSpec specifies the desired state of this Account.
	StorageAccountSpec *StorageAccountSpec `json:"storageAccountSpec"`
}

// An AccountSpec defines the desired state of an Account.
type AccountSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	AccountParameters            `json:",inline"`
}

// An AccountStatus represents the observed state of an Account.
type AccountStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	*StorageAccountStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// An Account is a managed resource that represents an Azure Blob Service
// Account.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="RESOURCE_GROUP",type="string",JSONPath=".spec.resourceGroupName"
// +kubebuilder:printcolumn:name="ACCOUNT_NAME",type="string",JSONPath=".spec.storageAccountName"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="RECLAIM_POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type Account struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AccountSpec   `json:"spec"`
	Status            AccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccountList contains a list of Account.
type AccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Account `json:"items"`
}

// ContainerParameters define the desired state of an Azure Blob Storage
// Container.
type ContainerParameters struct {
	// Metadata for this Container.
	// +optional
	Metadata azblob.Metadata `json:"metadata,omitempty"`

	// PublicAccessType for this container; either "blob" or "container".
	// +optional
	PublicAccessType azblob.PublicAccessType `json:"publicAccessType,omitempty"`
}

// A ContainerSpec defines the desired state of a Container.
type ContainerSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ContainerParameters          `json:",inline"`
}

// A ContainerStatus represents the observed status of a Container.
type ContainerStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// A Container is a managed resource that represents an Azure Blob Storage
// Container.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STORAGE_ACCOUNT",type="string",JSONPath=".spec.accountRef.name"
// +kubebuilder:printcolumn:name="PUBLIC_ACCESS_TYPE",type="string",JSONPath=".spec.publicAccessType"
// +kubebuilder:printcolumn:name="RECLAIM_POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type Container struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ContainerSpec   `json:"spec"`
	Status            ContainerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ContainerList - list of the container objects
type ContainerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Container `json:"items"`
}

// A ContainerClassSpecTemplate is a template for the spec of a dynamically
// provisioned Container.
type ContainerClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	ContainerParameters               `json:",inline"`
}

// +kubebuilder:object:root=true

// A ContainerClass is a non-portable resource class. It defines the desired
// spec of resource claims that use it to dynamically provision a managed
// resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,class,azure}
type ContainerClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// Container.
	SpecTemplate ContainerClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// ContainerClassList contains a list of cloud memorystore resource classes.
type ContainerClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ContainerClass `json:"items"`
}

func parsePublicAccessType(s string) azblob.PublicAccessType {
	if s == "" {
		return azblob.PublicAccessNone
	}
	return azblob.PublicAccessType(s)
}
