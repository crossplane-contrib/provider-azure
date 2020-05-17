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

// FirewallRuleProperties defines the properties of an Azure SQL firewall rule.
type FirewallRuleProperties struct {
	// StartIPAddress of the IP range this firewall rule allows.
	StartIPAddress string `json:"startIpAddress"`

	// EndIPAddress of the IP range this firewall rule allows.
	EndIPAddress string `json:"endIpAddress"`
}

// A FirewallRuleObservation represents the observed state of an Azure SQL
// firewall rule.
type FirewallRuleObservation struct {
	// ID - Resource ID
	ID string `json:"id,omitempty"`

	// Type - Resource type.
	Type string `json:"type,omitempty"`
}

// A FirewallRuleStatus represents the status of an Azure SQL firewall rule.
type FirewallRuleStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     FirewallRuleObservation `json:"atProvider,omitempty"`
}

// FirewallRuleParameters define the desired state of an Azure SQL firewall
// rule.
type FirewallRuleParameters struct {
	// ServerName - Name of the Firewall Rule's server.
	ServerName string `json:"serverName,omitempty"`

	// ServerNameRef - A reference to the Firewall Rule's MySQLServer.
	ServerNameRef *runtimev1alpha1.Reference `json:"serverNameRef,omitempty"`

	// ServerNameSelector - Selects a MySQLServer to reference.
	ServerNameSelector *runtimev1alpha1.Selector `json:"serverNameSelector,omitempty"`

	// ResourceGroupName - Name of the Firewall Rule's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	ResourceGroupNameRef *runtimev1alpha1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Selects a ResourceGroup to reference.
	ResourceGroupNameSelector *runtimev1alpha1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// FirewallRuleProperties - Resource properties.
	FirewallRuleProperties `json:"properties"`
}

// A FirewallRuleSpec defines the desired state of an Azure SQL firewall rule.
type FirewallRuleSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  FirewallRuleParameters `json:"forProvider"`
}

// +kubebuilder:object:root=true

// A MySQLServerFirewallRule is a managed resource that represents an Azure
// MySQL firewall rule.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type MySQLServerFirewallRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallRuleSpec   `json:"spec"`
	Status FirewallRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MySQLServerFirewallRuleList contains a list of MySQLServerFirewallRule.
type MySQLServerFirewallRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQLServerFirewallRule `json:"items"`
}

// +kubebuilder:object:root=true

// A PostgreSQLServerFirewallRule is a managed resource that represents an Azure
// PostgreSQL firewall rule.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type PostgreSQLServerFirewallRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FirewallRuleSpec   `json:"spec"`
	Status FirewallRuleStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PostgreSQLServerFirewallRuleList contains a list of
// PostgreSQLServerFirewallRule.
type PostgreSQLServerFirewallRuleList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQLServerFirewallRule `json:"items"`
}
