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

package v1beta1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"

	apisv1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
)

// Possible state strings for SQL types.
const (
	StateDisabled = "Disabled"
	StateDropping = "Dropping"
	StateReady    = "Ready"
)

// +kubebuilder:object:root=true

// A MySQLServer is a managed resource that represents an Azure MySQL Database
// Server.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.userVisibleState"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type MySQLServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SQLServerSpec   `json:"spec"`
	Status SQLServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MySQLServerList contains a list of MySQLServer.
type MySQLServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQLServer `json:"items"`
}

// +kubebuilder:object:root=true

// A PostgreSQLServer is a managed resource that represents an Azure PostgreSQL
// Database Server.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.userVisibleState"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type PostgreSQLServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SQLServerSpec   `json:"spec"`
	Status SQLServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PostgreSQLServerList contains a list of PostgreSQLServer.
type PostgreSQLServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PostgreSQLServer `json:"items"`
}

// A SQLServerClassSpecTemplate is a template for the spec of a dynamically
// provisioned MySQLServer or PostgreSQLServer.
type SQLServerClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	ForProvider                       SQLServerParameters `json:"forProvider"`
}

// +kubebuilder:object:root=true

// A SQLServerClass is a non-portable resource class. It defines the desired
// spec of resource claims that use it to dynamically provision a managed
// resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type SQLServerClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// SQLServer.
	SpecTemplate SQLServerClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// SQLServerClassList contains a list of SQLServerClass.
type SQLServerClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SQLServerClass `json:"items"`
}

// SKU billing information related properties of a server.
type SKU struct {
	// Tier - The tier of the particular SKU.
	// Possible values include: 'Basic', 'GeneralPurpose', 'MemoryOptimized'
	// +kubebuilder:validation:Enum=Basic;GeneralPurpose;MemoryOptimized
	Tier string `json:"tier"`

	// Capacity - The scale up/out capacity, representing server's compute units.
	Capacity int `json:"capacity"`

	// Size - The size code, to be interpreted by resource as appropriate.
	// +optional
	Size *string `json:"size,omitempty"`

	// Family - The family of hardware.
	Family string `json:"family"`
}

// StorageProfile storage Profile properties of a server
type StorageProfile struct {
	// BackupRetentionDays - Backup retention days for the server.
	// +optional
	BackupRetentionDays *int `json:"backupRetentionDays,omitempty"`

	// GeoRedundantBackup - Enable Geo-redundant or not for server backup.
	// Possible values include: 'Enabled', 'Disabled'
	// +kubebuilder:validation:Enum=Enabled;Disabled
	// +optional
	GeoRedundantBackup *string `json:"geoRedundantBackup,omitempty"`

	// StorageMB - Max storage allowed for a server.
	StorageMB int `json:"storageMB"`

	// StorageAutogrow - Enable Storage Auto Grow.
	// Possible values include: 'Enabled', 'Disabled'
	// +kubebuilder:validation:Enum=Enabled;Disabled
	// +optional
	StorageAutogrow *string `json:"storageAutogrow,omitempty"`
}

// SQLServerParameters define the desired state of an Azure SQL Database, either
// PostgreSQL or MySQL.
type SQLServerParameters struct {
	// ResourceGroupName specifies the name of the resource group that should
	// contain this SQLServer.
	// +immutable
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	// +immutable
	ResourceGroupNameRef *runtimev1alpha1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - A selector for a ResourceGroup object to
	// retrieve its name
	// +immutable
	ResourceGroupNameSelector *runtimev1alpha1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// SKU is the billing information related properties of the server.
	SKU SKU `json:"sku"`

	// Location specifies the location of this SQLServer.
	// +immutable
	Location string `json:"location"`

	// AdministratorLogin - The administrator's login name of a server. Can only be specified when the server is being created (and is required for creation).
	// +immutable
	AdministratorLogin string `json:"administratorLogin"`

	// TODO(hasheddan): support AdministratorLoginPassword

	// TODO(hasheddan): support MinimalTLSVersion

	// TODO(hasheddan): support InfrastructureEncryption

	// TODO(hasheddan): support PublicNetworkAccess

	// TODO(hasheddan): support CreateMode

	// Tags - Application-specific metadata in the form of key-value pairs.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`

	// Version - Server version.
	Version string `json:"version"`

	// SSLEnforcement - Enable ssl enforcement or not when connect to server. Possible values include: 'Enabled', 'Disabled'
	// +kubebuilder:validation:Enum=Enabled;Disabled
	SSLEnforcement string `json:"sslEnforcement"`

	// StorageProfile - Storage profile of a server.
	StorageProfile StorageProfile `json:"storageProfile"`
}

// A SQLServerSpec defines the desired state of a SQLServer.
type SQLServerSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  SQLServerParameters `json:"forProvider"`
}

// SQLServerObservation represents the current state of Azure SQL resource.
type SQLServerObservation struct {
	// ID - Resource ID
	ID string `json:"id,omitempty"`

	// Name - Resource name.
	Name string `json:"name,omitempty"`

	// Type - Resource type.
	Type string `json:"type,omitempty"`

	// UserVisibleState - A state of a server that is visible to user.
	UserVisibleState string `json:"userVisibleState,omitempty"`

	// FullyQualifiedDomainName - The fully qualified domain name of a server.
	FullyQualifiedDomainName string `json:"fullyQualifiedDomainName,omitempty"`

	// MasterServerID - The master server id of a replica server.
	MasterServerID string `json:"masterServerId,omitempty"`

	// LastOperation represents the state of the last operation started by the
	// controller.
	LastOperation apisv1alpha3.AsyncOperation `json:"lastOperation,omitempty"`
}

// A SQLServerStatus represents the observed state of a SQLServer.
type SQLServerStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     SQLServerObservation `json:"atProvider,omitempty"`
}
