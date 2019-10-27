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

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"

	networkv1alpha3 "github.com/crossplaneio/stack-azure/apis/network/v1alpha3"
	apisv1alpha3 "github.com/crossplaneio/stack-azure/apis/v1alpha3"
)

const (
	// OperationCreateServer is the operation type for creating a new mysql
	// server.
	OperationCreateServer = "createServer"

	// OperationCreateFirewallRules is the operation type for creating a
	// firewall rule.
	OperationCreateFirewallRules = "createFirewallRules"
)

// Error strings
const (
	errResourceIsNotSQLServer                                  = "the managed resource is not a MySQLServer or PostgreSQLServer"
	errResourceIsNotPostgreSQLServerVirtualNetworkRule         = "the managed resource is not a PostgreSQLServerVirtualNetworkRule"
	errResourceIsNotMySQLServerVirtualNetworkRule              = "the managed resource is not a MySQLServerVirtualNetworkRule"
	errResourceIsNotPostgreSQLNorMySQLServerVirtualNetworkRule = "the managed resource is not MySQLServerVirtualNetworkRule or PostgreSQLServerVirtualNetworkRule"
)

// SubnetIDReferencerForVirtualNetworkRule is an attribute
// referencer that resolves id from a referenced Subnet and assigns it to a
// PostgreSQLServer or MySQL server object
type SubnetIDReferencerForVirtualNetworkRule struct {
	networkv1alpha3.SubnetIDReferencer `json:",inline"`
}

// Assign assigns the retrieved group name to the managed resource
func (v *SubnetIDReferencerForVirtualNetworkRule) Assign(res resource.CanReference, value string) error {
	switch sql := res.(type) {
	case *MySQLServerVirtualNetworkRule:
		sql.Spec.VirtualNetworkRuleProperties.VirtualNetworkSubnetID = value
	case *PostgreSQLServerVirtualNetworkRule:
		sql.Spec.VirtualNetworkRuleProperties.VirtualNetworkSubnetID = value
	default:
		return errors.New(errResourceIsNotPostgreSQLNorMySQLServerVirtualNetworkRule)
	}

	return nil
}

// ResourceGroupNameReferencerForVirtualNetworkRule is an attribute referencer
// that resolves the name of a the ResourceGroup.
type ResourceGroupNameReferencerForVirtualNetworkRule struct {
	apisv1alpha3.ResourceGroupNameReferencer `json:",inline"`
}

// Assign assigns the retrieved group name to the managed resource
func (v *ResourceGroupNameReferencerForVirtualNetworkRule) Assign(res resource.CanReference, value string) error {
	switch sql := res.(type) {
	case *MySQLServerVirtualNetworkRule:
		sql.Spec.ResourceGroupName = value
	case *PostgreSQLServerVirtualNetworkRule:
		sql.Spec.ResourceGroupName = value
	default:
		return errors.New(errResourceIsNotPostgreSQLNorMySQLServerVirtualNetworkRule)
	}
	return nil
}

// ResourceGroupNameReferencerForSQLServer is an attribute referencer that
// resolves the name of a the ResourceGroup.
type ResourceGroupNameReferencerForSQLServer struct {
	apisv1alpha3.ResourceGroupNameReferencer `json:",inline"`
}

// Assign assigns the retrieved group name to the managed resource
func (v *ResourceGroupNameReferencerForSQLServer) Assign(res resource.CanReference, value string) error {
	switch sql := res.(type) {
	case *MySQLServer:
		sql.Spec.ResourceGroupName = value
	case *PostgreSQLServer:
		sql.Spec.ResourceGroupName = value
	default:
		return errors.New(errResourceIsNotSQLServer)
	}
	return nil
}

// ServerNameReferencerForPostgreSQLServerVirtualNetworkRule is an attribute
// referencer that resolves the name of a PostgreSQLServer.
type ServerNameReferencerForPostgreSQLServerVirtualNetworkRule struct {
	PostgreSQLServerNameReferencer `json:",inline"`
}

// Assign assigns the retrieved group name to the managed resource
func (v *ServerNameReferencerForPostgreSQLServerVirtualNetworkRule) Assign(res resource.CanReference, value string) error {
	vnet, ok := res.(*PostgreSQLServerVirtualNetworkRule)
	if !ok {
		return errors.Errorf(errResourceIsNotPostgreSQLServerVirtualNetworkRule)
	}

	vnet.Spec.ServerName = value
	return nil
}

// ServerNameReferencerForMySQLServerVirtualNetworkRule is an attribute
// referencer that resolves the name of a MySQLServer.
type ServerNameReferencerForMySQLServerVirtualNetworkRule struct {
	MySQLServerNameReferencer `json:",inline"`
}

// Assign assigns the retrieved group name to the managed resource
func (v *ServerNameReferencerForMySQLServerVirtualNetworkRule) Assign(res resource.CanReference, value string) error {
	vnet, ok := res.(*MySQLServerVirtualNetworkRule)
	if !ok {
		return errors.Errorf(errResourceIsNotMySQLServerVirtualNetworkRule)
	}

	vnet.Spec.ServerName = value
	return nil
}

// +kubebuilder:object:root=true

// A MySQLServer is a managed resource that represents an Azure MySQL Database
// Server.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type MySQLServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SQLServerSpec   `json:"spec,omitempty"`
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
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.version"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type PostgreSQLServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SQLServerSpec   `json:"spec,omitempty"`
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
	SQLServerParameters               `json:",inline"`
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

// SQLServerParameters define the desired state of an Azure SQL Database, either
// PostgreSQL or MySQL.
type SQLServerParameters struct {

	// ResourceGroupName specifies the name of the resource group that should
	// contain this SQLServer.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	ResourceGroupNameRef *ResourceGroupNameReferencerForSQLServer `json:"resourceGroupNameRef,omitempty" resource:"attributereferencer"`

	// Location specifies the location of this SQLServer.
	Location string `json:"location"`

	// PricingTier specifies the pricing tier (aka SKU) for this SQLServer.
	PricingTier PricingTierSpec `json:"pricingTier"`

	// StorageProfile configures the storage profile of this SQLServer.
	StorageProfile StorageProfileSpec `json:"storageProfile"`

	// AdminLoginName specifies the administrator login name for this SQLServer.
	AdminLoginName string `json:"adminLoginName"`

	// Version specifies the version of this server, for
	// example "5.6", or "9.6".
	Version string `json:"version"`

	// SSLEnforced specifies whether SSL is required to connect to this
	// SQLServer.
	// +optional
	SSLEnforced bool `json:"sslEnforced,omitempty"`
}

// A SQLServerSpec defines the desired state of a SQLServer.
type SQLServerSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	SQLServerParameters          `json:",inline"`
}

// A SQLServerStatus represents the observed state of a SQLServer.
type SQLServerStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// State of this SQLServer.
	State string `json:"state,omitempty"`

	// A Message containing detail on the state of this SQLServer, if any.
	Message string `json:"message,omitempty"`

	// ProviderID is the external ID to identify this resource in the cloud
	// provider.
	ProviderID string `json:"providerID,omitempty"`

	// Endpoint of the MySQL Server instance used in connection strings.
	Endpoint string `json:"endpoint,omitempty"`
}

// PricingTierSpec represents the performance and cost oriented properties of a
// SQLServer.
type PricingTierSpec struct {
	// Tier of the particular SKU, e.g. Basic. Possible values include: 'Basic',
	// 'GeneralPurpose', 'MemoryOptimized'
	Tier string `json:"tier"`

	// VCores (aka Capacity) specifies how many virtual cores this SQLServer
	// requires.
	VCores int `json:"vcores"`

	// Family of hardware.
	Family string `json:"family"`
}

// A StorageProfileSpec represents storage related properties of a SQLServer.
type StorageProfileSpec struct {
	// StorageGB configures the maximum storage allowed.
	StorageGB int `json:"storageGB"`

	// BackupRetentionDays configures how many days backups will be retained.
	BackupRetentionDays int `json:"backupRetentionDays,omitempty"`

	// GeoRedundantBackup enables geo-redunndant backups.
	GeoRedundantBackup bool `json:"geoRedundantBackup,omitempty"`
}

// ValidMySQLVersionValues returns the valid set of engine version values.
func ValidMySQLVersionValues() []string {
	return []string{"5.6", "5.7"}
}

// ValidPostgreSQLVersionValues returns the valid set of engine version values.
func ValidPostgreSQLVersionValues() []string {
	return []string{"9.5", "9.6", "10", "10.0", "10.2"}
}

// VirtualNetworkRuleProperties defines the properties of a VirtualNetworkRule.
type VirtualNetworkRuleProperties struct {
	// VirtualNetworkSubnetID - The ARM resource id of the virtual network
	// subnet.
	VirtualNetworkSubnetID string `json:"virtualNetworkSubnetId,omitempty"`

	// VirtualNetworkSubnetIDRef - A reference to a Subnet to retrieve its ID
	VirtualNetworkSubnetIDRef *SubnetIDReferencerForVirtualNetworkRule `json:"virtualNetworkSubnetIdRef,omitempty" resource:"attributereferencer"`

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

	// Name - Name of the Virtual Network Rule.
	Name string `json:"name"`

	// ServerName - Name of the Virtual Network Rule's PostgreSQLServer.
	ServerName string `json:"serverName,omitempty"`

	// ServerNameRef - A reference to the Virtual Network Rule's PostgreSQLServer.
	ServerNameRef *PostgreSQLServerNameReferencer `json:"serverNameRef,omitempty" resource:"attributereferencer"`

	// ResourceGroupName - Name of the Virtual Network Rule's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	ResourceGroupNameRef *ResourceGroupNameReferencerForVirtualNetworkRule `json:"resourceGroupNameRef,omitempty" resource:"attributereferencer"`

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
// +kubebuilder:resource:scope=Cluster
type PostgreSQLServerVirtualNetworkRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PostgreSQLVirtualNetworkRuleSpec `json:"spec,omitempty"`
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

	// Name - Name of the Virtual Network Rule.
	Name string `json:"name"`

	// ServerName - Name of the Virtual Network Rule's server.
	ServerName string `json:"serverName,omitempty"`

	// ServerNameRef - A reference to the Virtual Network Rule's MySQLServer.
	ServerNameRef *MySQLServerNameReferencer `json:"serverNameRef,omitempty" resource:"attributereferencer"`

	// ResourceGroupName - Name of the Virtual Network Rule's resource group.
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	ResourceGroupNameRef *ResourceGroupNameReferencerForVirtualNetworkRule `json:"resourceGroupNameRef,omitempty" resource:"attributereferencer"`

	// VirtualNetworkRuleProperties - Resource properties.
	VirtualNetworkRuleProperties `json:"properties"`
}

// +kubebuilder:object:root=true

// A MySQLServerVirtualNetworkRule is a managed resource that represents an
// Azure MySQL Database virtual network rule.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type MySQLServerVirtualNetworkRule struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLVirtualNetworkRuleSpec `json:"spec,omitempty"`
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
