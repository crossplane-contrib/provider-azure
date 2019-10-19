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
	"github.com/Azure/azure-storage-blob-go/azblob"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/util"
)

// AccountParameters define the desired state of an Azure Blob Storage Account.
type AccountParameters struct {
	// ResourceGroupName specifies the resource group for this Account.
	ResourceGroupName string `json:"resourceGroupName"`

	// StorageAccountName specifies the name for this Account.
	// +kubebuilder:validation:MaxLength=24
	StorageAccountName string `json:"storageAccountName"`

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
// +kubebuilder:resource:scope=Cluster
type Account struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AccountSpec   `json:"spec,omitempty"`
	Status            AccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AccountList contains a list of Account.
type AccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Account `json:"items"`
}

// An AccountClassSpecTemplate is a template for the spec of a dynamically
// provisioned Account.
type AccountClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	AccountParameters                 `json:",inline"`
}

// +kubebuilder:object:root=true

// An AccountClass is a non-portable resource class. It defines the desired spec
// of resource claims that use it to dynamically provision a managed resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type AccountClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// Account.
	SpecTemplate AccountClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// AccountClassList contains a list of AccountClass.
type AccountClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AccountClass `json:"items"`
}

// ContainerParameters define the desired state of an Azure Blob Storage
// Container.
type ContainerParameters struct {
	// NameFormat specifies the name of the external Container. The first
	// instance of the string '%s' will be replaced with the Kubernetes
	// UID of this Container.
	NameFormat string `json:"nameFormat"`

	// Metadata for this Container.
	// +optional
	Metadata azblob.Metadata `json:"metadata,omitempty"`

	// PublicAccessType for this container; either "blob" or "container".
	// +optional
	PublicAccessType azblob.PublicAccessType `json:"publicAccessType,omitempty"`

	// AccountReference to the Azure Blob Storage Account this Container will
	// reside within.
	AccountReference corev1.LocalObjectReference `json:"accountReference"`
}

// A ContainerSpec defines the desired state of a Container.
type ContainerSpec struct {
	ContainerParameters `json:",inline"`

	// NOTE(negz): Container is the only Crossplane type that does not use a
	// Provider (it reads credentials from its associated Account instead). This
	// means we can't embed a coreruntimev1alpha1.ResourceSpec, as doing so would
	// require a redundant providerRef be specified. Instead we duplicate
	// most of that struct here; the below values should be kept in sync with
	// coreruntimev1alpha1.ResourceSpec.

	// WriteConnectionSecretToReference specifies the name of a Secret, in the
	// same namespace as this managed resource, to which any connection details
	// for this managed resource should be written. Connection details
	// frequently include the endpoint, username, and password required to
	// connect to the managed resource.
	// +optional
	WriteConnectionSecretToReference *runtimev1alpha1.SecretReference `json:"writeConnectionSecretToRef,omitempty"`

	// ClaimReference specifies the resource claim to which this managed
	// resource will be bound. ClaimReference is set automatically during
	// dynamic provisioning. Crossplane does not currently support setting this
	// field manually, per https://github.com/crossplaneio/crossplane-runtime/issues/19
	// +optional
	ClaimReference *corev1.ObjectReference `json:"claimRef,omitempty"`

	// ClassReference specifies the non-portable resource class that was used to
	// dynamically provision this managed resource, if any. Crossplane does not
	// currently support setting this field manually, per
	// https://github.com/crossplaneio/crossplane-runtime/issues/20
	// +optional
	ClassReference *corev1.ObjectReference `json:"classRef,omitempty"`

	// ReclaimPolicy specifies what will happen to the external resource this
	// managed resource manages when the managed resource is deleted. "Delete"
	// deletes the external resource, while "Retain" (the default) does not.
	// Note this behaviour is subtly different from other uses of the
	// ReclaimPolicy concept within the Kubernetes ecosystem per
	// https://github.com/crossplaneio/crossplane-runtime/issues/21
	// +optional
	ReclaimPolicy runtimev1alpha1.ReclaimPolicy `json:"reclaimPolicy,omitempty"`
}

// A ContainerStatus represents the observed status of a Container.
type ContainerStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// Name of this Container.
	Name string `json:"name,omitempty"`
}

// +kubebuilder:object:root=true

// A Container is a managed resource that represents an Azure Blob Storage
// Container.
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="STORAGE_ACCOUNT",type="string",JSONPath=".spec.accountRef.name"
// +kubebuilder:printcolumn:name="PUBLIC_ACCESS_TYPE",type="string",JSONPath=".spec.publicAccessType"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="RECLAIM_POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type Container struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ContainerSpec   `json:"spec,omitempty"`
	Status            ContainerStatus `json:"status,omitempty"`
}

// SetBindingPhase of this Container.
func (c *Container) SetBindingPhase(p runtimev1alpha1.BindingPhase) {
	c.Status.SetBindingPhase(p)
}

// GetBindingPhase of this Container.
func (c *Container) GetBindingPhase() runtimev1alpha1.BindingPhase {
	return c.Status.GetBindingPhase()
}

// SetConditions of this Container.
func (c *Container) SetConditions(cd ...runtimev1alpha1.Condition) {
	c.Status.SetConditions(cd...)
}

// SetClaimReference of this Container.
func (c *Container) SetClaimReference(r *corev1.ObjectReference) {
	c.Spec.ClaimReference = r
}

// GetClaimReference of this Container.
func (c *Container) GetClaimReference() *corev1.ObjectReference {
	return c.Spec.ClaimReference
}

// SetClassReference of this Container.
func (c *Container) SetClassReference(r *corev1.ObjectReference) {
	c.Spec.ClassReference = r
}

// GetClassReference of this Container.
func (c *Container) GetClassReference() *corev1.ObjectReference {
	return c.Spec.ClassReference
}

// SetWriteConnectionSecretToReference of this Container.
func (c *Container) SetWriteConnectionSecretToReference(r *runtimev1alpha1.SecretReference) {
	c.Spec.WriteConnectionSecretToReference = r
}

// GetWriteConnectionSecretToReference of this Container.
func (c *Container) GetWriteConnectionSecretToReference() *runtimev1alpha1.SecretReference {
	return c.Spec.WriteConnectionSecretToReference
}

// GetReclaimPolicy of this Container.
func (c *Container) GetReclaimPolicy() runtimev1alpha1.ReclaimPolicy {
	return c.Spec.ReclaimPolicy
}

// SetReclaimPolicy of this Container.
func (c *Container) SetReclaimPolicy(p runtimev1alpha1.ReclaimPolicy) {
	c.Spec.ReclaimPolicy = p
}

// GetContainerName based on the NameFormat spec value,
// If name format is not provided, container name defaults to UID
// If name format provided with '%s' value, container name will result in formatted string + UID,
//   NOTE: only single %s substitution is supported
// If name format does not contain '%s' substitution, i.e. a constant string, the
// constant string value is returned back
//
// Examples:
//   For all examples assume "UID" = "test-uid"
//   1. NameFormat = "", ContainerName = "test-uid"
//   2. NameFormat = "%s", ContainerName = "test-uid"
//   3. NameFormat = "foo", ContainerName = "foo"
//   4. NameFormat = "foo-%s", ContainerName = "foo-test-uid"
//   5. NameFormat = "foo-%s-bar-%s", ContainerName = "foo-test-uid-bar-%!s(MISSING)"
func (c *Container) GetContainerName() string {
	return util.ConditionalStringFormat(c.Spec.NameFormat, string(c.GetUID()))
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
// +kubebuilder:resource:scope=Cluster
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
