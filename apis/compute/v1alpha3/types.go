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
	"github.com/pkg/errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	networkv1alpha3 "github.com/crossplaneio/stack-azure/apis/network/v1alpha3"
	apisv1alpha3 "github.com/crossplaneio/stack-azure/apis/v1alpha3"
)

const (
	// ClusterProvisioningStateSucceeded is the state for a cluster that has
	// succeeded provisioning.
	ClusterProvisioningStateSucceeded = "Succeeded"

	// DefaultReclaimPolicy is the default reclaim policy to use.
	DefaultReclaimPolicy = runtimev1alpha1.ReclaimRetain

	// DefaultNodeCount is the default node count for a cluster.
	DefaultNodeCount = 1
)

// Error strings
const (
	errResourceIsNotAKSCluster = "the managed resource is not an AKSCluster"
)

// ResourceGroupNameReferencerForAKSCluster is an attribute referencer that resolves name from a referenced ResourceGroup
type ResourceGroupNameReferencerForAKSCluster struct {
	apisv1alpha3.ResourceGroupNameReferencer `json:",inline"`
}

// Assign assigns the retrieved group name to the managed resource
func (v *ResourceGroupNameReferencerForAKSCluster) Assign(res resource.CanReference, value string) error {
	aks, ok := res.(*AKSCluster)
	if !ok {
		return errors.Errorf(errResourceIsNotAKSCluster)
	}

	aks.Spec.ResourceGroupName = value
	return nil
}

// SubnetIDReferencerForAKSCluster is an attribute referencer that resolves name from a referenced ResourceGroup
type SubnetIDReferencerForAKSCluster struct {
	networkv1alpha3.SubnetIDReferencer `json:",inline"`
}

// Assign assigns the retrieved group name to the managed resource
func (v *SubnetIDReferencerForAKSCluster) Assign(res resource.CanReference, value string) error {
	aks, ok := res.(*AKSCluster)
	if !ok {
		return errors.Errorf(errResourceIsNotAKSCluster)
	}

	aks.Spec.VnetSubnetID = value
	return nil
}

// AKSClusterParameters define the desired state of an Azure Kubernetes Engine
// cluster.
type AKSClusterParameters struct {
	// ResourceGroupName is the name of the resource group that the cluster will
	// be created in
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	ResourceGroupNameRef *ResourceGroupNameReferencerForAKSCluster `json:"resourceGroupNameRef,omitempty" resource:"attributereferencer"`

	// Location is the Azure location that the cluster will be created in
	Location string `json:"location"`

	// Version is the Kubernetes version that will be deployed to the cluster
	Version string `json:"version"`

	// VnetSubnetID is the subnet to which the cluster will be deployed.
	// +optional
	VnetSubnetID string `json:"vnetSubnetID,omitempty"`

	// ResourceGroupNameRef - A reference to a VnetSubnet object to retrieve
	// its ID
	VnetSubnetIDRef *SubnetIDReferencerForAKSCluster `json:"vnetSubnetIDRef,omitempty" resource:"attributereferencer"`

	// NodeCount is the number of nodes that the cluster will initially be
	// created with.  This can be scaled over time and defaults to 1.
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=0
	// +optional
	NodeCount *int `json:"nodeCount,omitempty"`

	// NodeVMSize is the name of the worker node VM size, e.g., Standard_B2s,
	// Standard_F2s_v2, etc.
	// +optional
	NodeVMSize string `json:"nodeVMSize"`

	// DNSNamePrefix is the DNS name prefix to use with the hosted Kubernetes
	// API server FQDN. You will use this to connect to the Kubernetes API when
	// managing containers after creating the cluster.
	// +optional
	DNSNamePrefix string `json:"dnsNamePrefix"`

	// DisableRBAC determines whether RBAC will be disabled or enabled in the
	// cluster.
	// +optional
	DisableRBAC bool `json:"disableRBAC,omitempty"`

	// WriteServicePrincipalSecretTo the specified Secret. The service principal
	// is automatically generated and used by the AKS cluster to interact with
	// other Azure resources.
	WriteServicePrincipalSecretTo runtimev1alpha1.SecretReference `json:"writeServicePrincipalTo"`
}

// An AKSClusterSpec defines the desired state of a AKSCluster.
type AKSClusterSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	AKSClusterParameters         `json:",inline"`
}

// An AKSClusterStatus represents the observed state of an AKSCluster.
type AKSClusterStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// ClusterName is the name of the cluster as registered with the cloud
	// provider.
	ClusterName string `json:"clusterName,omitempty"`

	// State is the current state of the cluster.
	State string `json:"state,omitempty"`

	// ProviderID is the external ID to identify this resource in the cloud
	// provider.
	ProviderID string `json:"providerID,omitempty"`

	// Endpoint is the endpoint where the cluster can be reached
	Endpoint string `json:"endpoint"`

	// ApplicationObjectID is the object ID of the AD application the cluster
	// uses for Azure APIs.
	ApplicationObjectID string `json:"appObjectID,omitempty"`

	// ServicePrincipalID is the ID of the service principal the AD application
	// uses.
	ServicePrincipalID string `json:"servicePrincipalID,omitempty"`

	// RunningOperation stores any current long running operation for this
	// instance across reconciliation attempts.
	RunningOperation string `json:"runningOperation,omitempty"`
}

// +kubebuilder:object:root=true

// An AKSCluster is a managed resource that represents an Azure Kubernetes
// Engine cluster.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="CLUSTER-NAME",type="string",JSONPath=".status.clusterName"
// +kubebuilder:printcolumn:name="ENDPOINT",type="string",JSONPath=".status.endpoint"
// +kubebuilder:printcolumn:name="CLUSTER-CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.location"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".spec.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type AKSCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AKSClusterSpec   `json:"spec,omitempty"`
	Status AKSClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AKSClusterList contains a list of AKSCluster.
type AKSClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AKSCluster `json:"items"`
}

// An AKSClusterClassSpecTemplate is a template for the spec of a dynamically
// provisioned AKSCluster.
type AKSClusterClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	AKSClusterParameters              `json:",inline"`
}

// +kubebuilder:object:root=true

// An AKSClusterClass is a non-portable resource class. It defines the desired
// spec of resource claims that use it to dynamically provision a managed
// resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type AKSClusterClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// AKSCluster.
	SpecTemplate AKSClusterClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// AKSClusterClassList contains a list of cloud memorystore resource classes.
type AKSClusterClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AKSClusterClass `json:"items"`
}
