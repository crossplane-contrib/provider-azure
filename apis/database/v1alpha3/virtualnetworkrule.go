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

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
)

const (
	// OperationCreateServer is the operation type for creating a new mysql
	// server.
	OperationCreateServer = "createServer"

	// OperationCreateFirewallRules is the operation type for creating a
	// firewall rule.
	OperationCreateFirewallRules = "createFirewallRules"
)

// VirtualNetworkRuleProperties defines the properties of a VirtualNetworkRule.
type VirtualNetworkRuleProperties struct {
	// VirtualNetworkSubnetID - The ARM resource id of the virtual network
	// subnet.
	VirtualNetworkSubnetID string `json:"virtualNetworkSubnetId,omitempty"`

	// VirtualNetworkSubnetIDRef - A reference to a Subnet to retrieve its ID
	VirtualNetworkSubnetIDRef *runtimev1alpha1.Reference `json:"virtualNetworkSubnetIdRef,omitempty"`

	// VirtualNetworkSubnetIDRef - A selector for a Subnet to retrieve its ID
	VirtualNetworkSubnetIDSelector *runtimev1alpha1.Selector `json:"virtualNetworkSubnetIdSelector,omitempty"`

	// IgnoreMissingVnetServiceEndpoint - Create firewall rule before the
	// virtual network has vnet service endpoint enabled.
	IgnoreMissingVnetServiceEndpoint bool `json:"ignoreMissingVnetServiceEndpoint,omitempty"`
}

// A VirtualNetworkRuleStatus represents the observed state of a
// VirtualNetworkRule.
type VirtualNetworkRuleStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// State of this virtual network rule.
	State string `json:"state,omitempty"`

	// A Message containing details about the state of this virtual network
	// rule, if any.
	Message string `json:"message,omitempty"`

	// ID - Resource ID
	ID string `json:"id,omitempty"`

	// Type - Resource type.
	Type string `json:"type,omitempty"`
}

// A PostgreSQLVirtualNetworkRuleSpec defines the desired state of a PostgreSQLVirtualNetworkRule.
type PostgreSQLVirtualNetworkRuleSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// ServerName - Name of the Virtual Network Rule's PostgreSQLServer.
	ServerName string `json:"serverName,omitempty"`

	// ServerNameRef - A reference to the Virtual Network Rule's PostgreSQLServer.
	ServerNameRef *runtimev1alpha1.Reference `json:"serverNameRef,omitempty"`

	// ServerNameSelector - A selector of the Virtual Network Rule's
	// PostgreSQLServer.
	ServerNameSelector *runtimev1alpha1.Selector `json:"serverNameSelector,omitempty"`

	// ResourceGroupName - Name of the Virtual Network Rule's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	ResourceGroupNameRef *runtimev1alpha1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - A selector for a ResourceGroup object to
	// retrieve its name
	ResourceGroupNameSelector *runtimev1alpha1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// VirtualNetworkRuleProperties - Resource properties.
	VirtualNetworkRuleProperties `json:"properties"`
}

// +kubebuilder:object:root=true

// A PostgreSQLServerVirtualNetworkRule is a managed resource that represents
// an Azure PostgreSQL Database virtual network rule.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type PostgreSQLServerVirtualNetworkRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgreSQLVirtualNetworkRuleSpec `json:"spec"`
	Status VirtualNetworkRuleStatus         `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PostgreSQLServerVirtualNetworkRuleList contains a list of PostgreSQLServerVirtualNetworkRule.
type PostgreSQLServerVirtualNetworkRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQLServerVirtualNetworkRule `json:"items"`
}

// A MySQLVirtualNetworkRuleSpec defines the desired state of a MySQLVirtualNetworkRule.
type MySQLVirtualNetworkRuleSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`

	// ServerName - Name of the Virtual Network Rule's server.
	ServerName string `json:"serverName,omitempty"`

	// ServerNameRef - A reference to the Virtual Network Rule's MySQLServer.
	ServerNameRef *runtimev1alpha1.Reference `json:"serverNameRef,omitempty"`

	// ServerNameSelector - Selects a MySQLServer to reference.
	ServerNameSelector *runtimev1alpha1.Selector `json:"serverNameSelector,omitempty"`

	// ResourceGroupName - Name of the Virtual Network Rule's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	ResourceGroupNameRef *runtimev1alpha1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Selects a ResourceGroup to reference.
	ResourceGroupNameSelector *runtimev1alpha1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// VirtualNetworkRuleProperties - Resource properties.
	VirtualNetworkRuleProperties `json:"properties"`
}

// +kubebuilder:object:root=true

// A MySQLServerVirtualNetworkRule is a managed resource that represents an
// Azure MySQL Database virtual network rule.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type MySQLServerVirtualNetworkRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLVirtualNetworkRuleSpec `json:"spec"`
	Status VirtualNetworkRuleStatus    `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MySQLServerVirtualNetworkRuleList contains a list of
// MySQLServerVirtualNetworkRule.
type MySQLServerVirtualNetworkRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQLServerVirtualNetworkRule `json:"items"`
}
