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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	apisv1alpha3 "github.com/crossplane-contrib/provider-azure/apis/v1alpha3"
)

// +kubebuilder:object:root=true

// A PostgreSQLServerConfiguration is a managed resource that represents an Azure
// PostgreSQL Server Configuration.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.version"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type PostgreSQLServerConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SQLServerConfigurationSpec   `json:"spec"`
	Status SQLServerConfigurationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PostgreSQLServerConfigurationList contains a list of PostgreSQLServerConfiguration.
type PostgreSQLServerConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQLServerConfiguration `json:"items"`
}

// +kubebuilder:object:root=true

// A MySQLServerConfiguration is a managed resource that represents an Azure
// MySQL Server Configuration.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.version"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type MySQLServerConfiguration struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SQLServerConfigurationSpec   `json:"spec"`
	Status SQLServerConfigurationStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MySQLServerConfigurationList contains a list of MySQLServerConfiguration.
type MySQLServerConfigurationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQLServerConfiguration `json:"items"`
}

// SQLServerConfigurationParameters define the desired state of an Azure SQL
// Database Server Configuration, either PostgreSQL or MySQL Configuration.
type SQLServerConfigurationParameters struct {
	// ResourceGroupName specifies the name of the resource group that should
	// contain this SQLServer.
	// +immutable
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	// +immutable
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - A selector for a ResourceGroup object to
	// retrieve its name
	// +immutable
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// ServerName specifies the name of the server that this
	// configuration applies to.
	// +immutable
	ServerName string `json:"serverName,omitempty"`

	// ServerNameRef - A reference to a server object to retrieve
	// its name
	// +immutable
	ServerNameRef *xpv1.Reference `json:"serverNameRef,omitempty"`

	// ServerNameSelector - A selector for a server object to
	// retrieve its name
	// +immutable
	ServerNameSelector *xpv1.Selector `json:"serverNameSelector,omitempty"`

	// Name - Configuration name to be applied
	// +kubebuilder:validation:Required
	// +immutable
	Name string `json:"name"`

	// Value - Configuration value to be applied
	// Can be left unset to read the current value
	// as a result of late-initialization.
	// +kubebuilder:validation:Optional
	Value *string `json:"value,omitempty"`
}

// A SQLServerConfigurationSpec defines the desired state of a SQLServer
// Configuration.
type SQLServerConfigurationSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SQLServerConfigurationParameters `json:"forProvider"`
}

// SQLServerConfigurationObservation represents the current state of Azure SQL resource.
type SQLServerConfigurationObservation struct {
	// ID - Resource ID
	ID string `json:"id,omitempty"`

	// Name - Resource name.
	Name string `json:"name,omitempty"`

	// Type - Resource type.
	Type string `json:"type,omitempty"`

	// DataType - Data type for the configuration
	DataType string `json:"dataType,omitempty"`

	// Value - Applied configuration value
	Value string `json:"value,omitempty"`

	// DefaultValue - Default value for this configuration
	DefaultValue string `json:"defaultValue,omitempty"`

	// Source - Applied configuration source
	Source string `json:"source,omitempty"`

	// Description - Description for the configuration
	Description string `json:"description,omitempty"`

	// LastOperation represents the state of the last operation started by the
	// controller.
	LastOperation apisv1alpha3.AsyncOperation `json:"lastOperation,omitempty"`
}

// A SQLServerConfigurationStatus represents the observed state of a
// SQLServerConfiguration.
type SQLServerConfigurationStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SQLServerConfigurationObservation `json:"atProvider,omitempty"`
}
