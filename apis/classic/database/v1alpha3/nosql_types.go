/*
Copyright 2020 The Crossplane Authors.

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

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// +kubebuilder:object:root=true

// A CosmosDBAccount is a managed resource that represents an Azure CosmosDB
// account with CosmosDB API.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.atProvider.state"
// +kubebuilder:printcolumn:name="KIND",type="string",JSONPath=".spec.forProvider.kind"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type CosmosDBAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CosmosDBAccountSpec   `json:"spec"`
	Status CosmosDBAccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CosmosDBAccountList contains a list of CosmosDB.
type CosmosDBAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []CosmosDBAccount `json:"items"`
}

// CosmosDBAccountParameters define the desired state of an Azure CosmosDB
// account.
type CosmosDBAccountParameters struct {
	// ResourceGroupName specifies the name of the resource group that should
	// contain this Account.
	// +immutable
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	// +immutable
	// +optional
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector to select a reference to a resource group.
	// +immutable
	// +optional
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// Kind - Indicates the type of database account.
	Kind documentdb.DatabaseAccountKind `json:"kind"`

	// Location - The location of the resource. This will be one of the
	// supported and registered Azure Geo Regions (e.g. West US, East US,
	// Southeast Asia, etc.).
	Location string `json:"location"`

	// Properties - Account properties like databaseAccountOfferType,
	// ipRangeFilters, etc.
	Properties CosmosDBAccountProperties `json:"properties"`

	// Tags - A list of key value pairs that describe the resource. These tags
	// can be used for viewing and grouping this resource (across resource
	// groups). A maximum of 15 tags can be provided for a resource. Each tag
	// must have a key with a length no greater than 128 characters and a value
	// with a length no greater than 256 characters.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// CosmosDBAccountObservation shows current state of an Azure CosmosDB account.
type CosmosDBAccountObservation struct {
	// Identity - The identity of the resource.
	ID string `json:"id"`

	// State - current state of the account in Azure.
	State string `json:"state"`
}

// CosmosDBAccountProperties define the desired properties of an Azure CosmosDB account.
type CosmosDBAccountProperties struct {
	// ConsistencyPolicy - The consistency policy for the Cosmos DB account.
	// + optional
	ConsistencyPolicy *CosmosDBAccountConsistencyPolicy `json:"consistencyPolicy,omitempty"`
	// Locations - An array that contains the georeplication locations enabled
	// for the Cosmos DB account.
	Locations []CosmosDBAccountLocation `json:"locations"`
	// DatabaseAccountOfferType - The offer type for the database
	DatabaseAccountOfferType string `json:"databaseAccountOfferType"`
	// IPRangeFilter - Cosmos DB Firewall Support: This value specifies the set
	// of IP addresses or IP address ranges in CIDR form to be included as the
	// allowed list of client IPs for a given database account. IP
	// addresses/ranges must be comma separated and must not contain any spaces.
	// + optional
	IPRangeFilter *string `json:"ipRangeFilter,omitempty"`
	// EnableAutomaticFailover - Enables automatic failover of the write region
	// in the rare event that the region is unavailable due to an outage.
	// Automatic failover will result in a new write region for the account and
	// is chosen based on the failover priorities configured for the account.
	// + optional
	EnableAutomaticFailover *bool `json:"enableAutomaticFailover,omitempty"`
	// EnableMultipleWriteLocations - Enables the account to write in multiple
	// locations
	// + optional
	EnableMultipleWriteLocations *bool `json:"enableMultipleWriteLocations,omitempty"`
	// EnableCassandraConnector - Enables the cassandra connector on the Cosmos
	// DB C* account
	// + optional
	EnableCassandraConnector *bool `json:"enableCassandraConnector,omitempty"`
}

// CosmosDBAccountConsistencyPolicy the consistency policy for the Cosmos DB
// database account.
type CosmosDBAccountConsistencyPolicy struct {
	// DefaultConsistencyLevel - The default consistency level and configuration
	// settings of the Cosmos DB account. Possible values include: 'Eventual',
	// 'Session', 'BoundedStaleness', 'Strong', 'ConsistentPrefix'
	DefaultConsistencyLevel string `json:"defaultConsistencyLevel"`
	// MaxStalenessPrefix - When used with the Bounded Staleness consistency
	// level, this value represents the number of stale requests tolerated.
	// Accepted range for this value is 1 – 2,147,483,647. Required when
	// defaultConsistencyPolicy is set to 'BoundedStaleness'.
	// + optional
	MaxStalenessPrefix *int64 `json:"maxStalenessPrefix,omitempty"`
	// MaxIntervalInSeconds - When used with the Bounded Staleness consistency
	// level, this value represents the time amount of staleness (in seconds)
	// tolerated. Accepted range for this value is 5 - 86400. Required when
	// defaultConsistencyPolicy is set to 'BoundedStaleness'.
	// + optional
	MaxIntervalInSeconds *int32 `json:"maxIntervalInSeconds,omitempty"`
}

// CosmosDBAccountLocation a region in which the Azure Cosmos DB database
// account is deployed.
type CosmosDBAccountLocation struct {
	// LocationName - The name of the region.
	LocationName string `json:"locationName"`
	// FailoverPriority - The failover priority of the region. A failover
	// priority of 0 indicates a write region. The maximum value for a failover
	// priority = (total number of regions - 1). Failover priority values must
	// be unique for each of the regions in which the database account exists.
	FailoverPriority int32 `json:"failoverPriority"`
	// IsZoneRedundant - Flag to indicate whether or not this region is an
	// AvailabilityZone region
	IsZoneRedundant bool `json:"isZoneRedundant"`
}

// A CosmosDBAccountSpec defines the desired state of a CosmosDB Account.
type CosmosDBAccountSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       CosmosDBAccountParameters `json:"forProvider"`
}

// An CosmosDBAccountStatus represents the observed state of an Account.
type CosmosDBAccountStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	// + optional
	AtProvider *CosmosDBAccountObservation `json:"atProvider,omitempty"`
}
