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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"

	apisv1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
)

// Possible state strings for SQL types.
const (
	StateDisabled = "Disabled"
	StateDropping = "Dropping"
	StateReady    = "Ready"
)

// PostgreSQLServerPort is the port PostgreSQLServer listens to.
const PostgreSQLServerPort = "5432"

// +kubebuilder:object:root=true

// A MySQLServer is a managed resource that represents an Azure MySQL Database
// Server.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.version"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
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
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.version"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
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
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - A selector for a ResourceGroup object to
	// retrieve its name
	// +immutable
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// SKU is the billing information related properties of the server.
	SKU SKU `json:"sku"`

	// Location specifies the location of this SQLServer.
	// +immutable
	Location string `json:"location"`

	// AdministratorLogin - The administrator's login name of a server. Can only be specified when the server is being created (and is required for creation).
	// +immutable
	AdministratorLogin string `json:"administratorLogin"`

	// TODO(hasheddan): support AdministratorLoginPassword

	// MinimalTLSVersion - control TLS connection policy
	MinimalTLSVersion string `json:"minimalTlsVersion,omitempty"`

	// TODO(hasheddan): support InfrastructureEncryption

	// TODO(hasheddan): support PublicNetworkAccess

	// CreateMode - Possible values include: 'CreateModeDefault', 'CreateModePointInTimeRestore', 'CreateModeGeoRestore', 'CreateModeReplica'
	// +optional
	CreateMode *CreateMode `json:"createMode,omitempty"`

	// RestorePointInTime - Restore point creation time (RFC3339 format), specifying the time to restore from.
	// +optional
	RestorePointInTime *metav1.Time `json:"restorePointInTime,omitempty"`

	// SourceServerID - The server to restore from when restoring or creating replicas
	// +optional
	SourceServerID *string `json:"sourceServerID,omitempty"`

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

// CreateMode controls the creation behaviour
// Keep synced with "github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql".MinimalTLSVersionEnum
// +kubebuilder:validation:Enum=Default;GeoRestore;PointInTimeRestore;Replica
type CreateMode string

// All valid values of CreateMode
const (
	CreateModeDefault            CreateMode = "Default"
	CreateModeReplica            CreateMode = "Replica"
	CreateModeGeoRestore         CreateMode = "GeoRestore"
	CreateModePointInTimeRestore CreateMode = "PointInTimeRestore"
)

// A SQLServerSpec defines the desired state of a SQLServer.
type SQLServerSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       SQLServerParameters `json:"forProvider"`
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
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          SQLServerObservation `json:"atProvider,omitempty"`
}
