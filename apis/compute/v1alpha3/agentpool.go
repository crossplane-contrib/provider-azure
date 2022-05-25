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

// AgentPoolParameters define the desired state of an Azure Kubernetes Engine
// agent pool.
type AgentPoolParameters struct {
	// ResourceGroupName is the name of the resource group that the cluster will
	// be created in
	// +immutable
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup to retrieve its
	// name
	// +immutable
	// +optional
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to a ResourceGroup to
	// retrieve its name
	// +optional
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// AKSClusterName is the name of the AKSCluster that the cluster will
	// be created in
	// +immutable
	AKSClusterName string `json:"aksClusterName,omitempty"`

	// AKSClusterNameRef - A reference to a AKSCluster to retrieve its
	// id
	// +immutable
	// +optional
	AKSClusterNameRef *xpv1.Reference `json:"aksClusterNameRef,omitempty"`

	// AKSClusterNameSelector - Select a reference to a AKSCluster to
	// retrieve its ids
	// +optional
	AKSClusterNameSelector *xpv1.Selector `json:"aksClusterNameSelector,omitempty"`

	// VnetSubnetID is the subnet to which the cluster will be deployed.
	// +immutable
	// +optional
	VnetSubnetID string `json:"vnetSubnetID,omitempty"`

	// VnetSubnetIDRef - A reference to a Subnet to retrieve its ID
	// +immutable
	// +optional
	VnetSubnetIDRef *xpv1.Reference `json:"vnetSubnetIDRef,omitempty"`

	// VnetSubnetIDSelector - Select a reference to a Subnet to retrieve
	// its ID
	// +optional
	VnetSubnetIDSelector *xpv1.Selector `json:"vnetSubnetIDSelector,omitempty"`

	// NodeCount is the number of nodes that the cluster will initially be
	// created with.  This can be scaled over time and defaults to 1.
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=0
	// +optional
	NodeCount *int32 `json:"nodeCount,omitempty"`

	// MaxNodeCount - Maximum number of nodes for auto-scaling
	// +optional
	MaxNodeCount *int32 `json:"maxNodeCount,omitempty"`

	// MinNodeCount - Minimum number of nodes for auto-scaling
	// +optional
	MinNodeCount *int32 `json:"minNodeCount,omitempty"`

	// EnableAutoScaling - Whether to enable auto-scaler
	EnableAutoScaling *bool `json:"enableAutoScaling,omitempty"`

	// Mode - AgentPoolMode represents mode of an agent pool.
	// Possible values include: 'System', 'User'.
	// Default: 'User'
	// +optional
	// +kubebuilder:validation:Enum=System;User
	Mode string `json:"mode,omitempty"`

	// AvailabilityZones - Availability zones for nodes.
	// +optional
	// +immutable
	AvailabilityZones []string `json:"availabilityZones,omitempty"`

	// NodeVMSize is the name of the worker node VM size, e.g., Standard_B2s,
	// Standard_F2s_v2, etc.
	// +kubebuilder:validation:Required
	NodeVMSize string `json:"nodeVMSize"`

	// NodeTaints - Taints added to new nodes during node pool create and scale.
	// For example, key=value:NoSchedule.
	// +optional
	// +immutable
	NodeTaints []string `json:"nodeTaints,omitempty"`
}

// An AgentPoolSpec defines the desired state of a AgentPool.
type AgentPoolSpec struct {
	xpv1.ResourceSpec   `json:",inline"`
	AgentPoolParameters `json:",inline"`
}

// An AgentPoolStatus represents the observed state of an AgentPool.
type AgentPoolStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// State is the current state of the agent pool.
	State string `json:"state,omitempty"`

	// NodesCount - Number of agents (VMs) to host docker containers.
	// Allowed values must be in the range of 0 to 100 (inclusive).
	// The default value is 1.
	NodesCount int `json:"nodesCount,omitempty"`

	// AvailabilityZones - Availability zones for nodes.
	// Must use VirtualMachineScaleSets AgentPoolType.
	AvailabilityZones []string `json:"availabilityZones,omitempty"`

	// ProviderID is the external ID to identify this resource in the cloud
	// provider.
	ProviderID string `json:"providerID,omitempty"`
}

// +kubebuilder:object:root=true

// An AgentPool is a managed resource that represents an Azure AgentPool.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="COUNT",type="integer",JSONPath=".status.nodesCount"
// +kubebuilder:printcolumn:name="AZ",type="string",JSONPath=".status.availabilityZones"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
// +kubebuilder:subresource:status
type AgentPool struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AgentPoolSpec   `json:"spec"`
	Status AgentPoolStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AgentPoolList contains a list of AgentPool.
type AgentPoolList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AgentPool `json:"items"`
}
