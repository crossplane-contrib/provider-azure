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
)

const (
	// SupportedRedisVersion is the only minor version of Redis currently
	// supported by Azure Cache for Redis. The version cannot be specified at
	// creation time.
	SupportedRedisVersion = "3.2"
)

// An SKU represents the performance and cost oriented properties of a
// Redis.
type SKU struct {
	//TODO: all three of them required? they might set defaults when sent as empty

	// Name specifies what type of Redis cache to deploy. Valid values: (Basic,
	// Standard, Premium). Possible values include: 'Basic', 'Standard',
	// 'Premium'
	// +kubebuilder:validation:Enum=Basic;Standard;Premium
	Name string `json:"name"`

	// Family specifies which family to use. Valid values: (C, P). Possible
	// values include: 'C', 'P'
	// +kubebuilder:validation:Enum=C;P
	Family string `json:"family"`

	// Capacity specifies the size of Redis cache to deploy. Valid values: for C
	// family (0, 1, 2, 3, 4, 5, 6), for P family (1, 2, 3, 4).
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=6
	Capacity int `json:"capacity"`
}

// TODO: trial-error the optional fields.

// RedisParameters define the desired state of an Azure Redis cluster.
// https://docs.microsoft.com/en-us/rest/api/redis/redis/create#redisresource
type RedisParameters struct {
	// TODO(muvaf): not included in received object.
	// ResourceGroupName in which to create this resource.
	// +immutable
	ResourceGroupName string `json:"resourceGroupName"`

	// Sku - The SKU of the Redis cache to deploy.
	SKU SKU `json:"sku"`

	// Location in which to create this resource.
	// +immutable
	Location string `json:"location"`

	// SubnetID specifies the full resource ID of a subnet in a virtual network
	// to deploy the Redis cache in. Example format:
	// /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/Microsoft.{Network|ClassicNetwork}/VirtualNetworks/vnet1/subnets/subnet1
	// +immutable
	// +optional
	SubnetID *string `json:"subnetId,omitempty"`

	// StaticIP address. Required when deploying a Redis cache inside an
	// existing Azure Virtual Network.
	// +immutable
	// +optional
	StaticIP *string `json:"staticIp,omitempty"`

	// RedisConfiguration specifies Redis Settings.
	// +optional
	RedisConfiguration map[string]string `json:"redisConfiguration,omitempty"`

	// EnableNonSSLPort specifies whether the non-ssl Redis server port (6379)
	// is enabled.
	// +optional
	EnableNonSSLPort *bool `json:"enableNonSslPort,omitempty"`

	// TenantSettings - A dictionary of tenant settings
	// +optional
	TenantSettings map[string]string `json:"tenantSettings,omitempty"`

	// ShardCount specifies the number of shards to be created on a Premium
	// Cluster Cache.
	// +optional
	ShardCount *int `json:"shardCount,omitempty"`

	// MinimumTLSVersion - Optional: requires clients to use a specified TLS
	// version (or higher) to connect (e,g, '1.0', '1.1', '1.2'). Possible
	// values include: 'OneFullStopZero', 'OneFullStopOne', 'OneFullStopTwo'
	// +optional
	MinimumTLSVersion *string `json:"minimumTlsVersion,omitempty"`

	// Zones - A list of availability zones denoting where the resource needs to come from.
	// +immutable
	// +optional
	Zones []string `json:"zones,omitempty"`

	// Tags - Resource tags.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`
}

// A RedisSpec defines the desired state of a Redis.
type RedisSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  RedisParameters `json:"forProvider,omitempty"`
}

// RedisObservation represents the observed state of the Redis object in Azure.
type RedisObservation struct {
	// RedisVersion - Redis version.
	RedisVersion string `json:"redisVersion,omitempty"`

	// ProvisioningState - Redis instance provisioning status. Possible values
	// include: 'Creating', 'Deleting', 'Disabled', 'Failed', 'Linking',
	// 'Provisioning', 'RecoveringScaleFailure', 'Scaling', 'Succeeded',
	// 'Unlinking', 'Unprovisioning', 'Updating'
	ProvisioningState string `json:"provisioningState,omitempty"`

	// HostName - Redis host name.
	HostName string `json:"hostName,omitempty"`

	// Port - Redis non-SSL port.
	Port int `json:"port,omitempty"`

	// SSLPort - Redis SSL port.
	SSLPort int `json:"sslPort,omitempty"`

	// TODO: keep in connection secret.
	// AccessKeys - The keys of the Redis cache - not set if this object is not
	// the response to Create or Update redis cache
	// AccessKeys *AccessKeys `json:"accessKeys,omitempty"`

	// LinkedServers - List of the linked servers associated with the cache
	LinkedServers []string `json:"linkedServers,omitempty"`

	// RedisConfiguration - All Redis Settings. Few possible keys:
	// rdb-backup-enabled,rdb-storage-connection-string,rdb-backup-frequency
	// maxmemory-delta,maxmemory-policy,notify-keyspace-events,maxmemory-samples,
	// slowlog-log-slower-than,slowlog-max-len,list-max-ziplist-entries,
	// list-max-ziplist-value,hash-max-ziplist-entries,hash-max-ziplist-value,
	// set-max-intset-entries,zset-max-ziplist-entries,zset-max-ziplist-value etc.
	RedisConfiguration map[string]string `json:"redisConfiguration"`

	// EnableNonSSLPort - Specifies whether the non-ssl Redis server port (6379) is enabled.
	EnableNonSSLPort bool `json:"enableNonSslPort,omitempty"`

	// TenantSettings - A dictionary of tenant settings
	TenantSettings map[string]string `json:"tenantSettings"`

	// ShardCount - The number of shards to be created on a Premium Cluster Cache.
	ShardCount int `json:"shardCount,omitempty"`

	// MinimumTLSVersion - Requires clients to use a specified TLS version (or higher) to connect (e,g, '1.0', '1.1', '1.2'). Possible values include: 'OneFullStopZero', 'OneFullStopOne', 'OneFullStopTwo'
	MinimumTLSVersion string `json:"minimumTlsVersion,omitempty"`
}

// A RedisStatus represents the observed state of a Redis.
type RedisStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
	AtProvider                     RedisObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Redis is a managed resource that represents an Azure Redis cluster.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.bindingPhase"
// +kubebuilder:printcolumn:name="STATE",type="string",JSONPath=".status.state"
// +kubebuilder:printcolumn:name="CLASS",type="string",JSONPath=".spec.classRef.name"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".status.redisVersion"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type Redis struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RedisSpec   `json:"spec,omitempty"`
	Status RedisStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RedisList contains a list of Redis.
type RedisList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Redis `json:"items"`
}

// A RedisClassSpecTemplate is a template for the spec of a dynamically
// provisioned Redis.
type RedisClassSpecTemplate struct {
	runtimev1alpha1.ClassSpecTemplate `json:",inline"`
	ForProvider                       RedisParameters `json:"forProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A RedisClass is a non-portable resource class. It defines the desired spec of
// resource claims that use it to dynamically provision a managed resource.
// +kubebuilder:printcolumn:name="PROVIDER-REF",type="string",JSONPath=".specTemplate.providerRef.name"
// +kubebuilder:printcolumn:name="RECLAIM-POLICY",type="string",JSONPath=".specTemplate.reclaimPolicy"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster
type RedisClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// SpecTemplate is a template for the spec of a dynamically provisioned
	// Redis.
	SpecTemplate RedisClassSpecTemplate `json:"specTemplate"`
}

// +kubebuilder:object:root=true

// RedisClassList contains a list of cloud memorystore resource classes.
type RedisClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RedisClass `json:"items"`
}
