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

package v1alpha2

import (
	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"

	"github.com/Azure/azure-sdk-for-go/services/redis/mgmt/2018-03-01/redis"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SKU options.
const (
	SKUNameBasic    = string(redis.Basic)
	SKUNamePremium  = string(redis.Premium)
	SKUNameStandard = string(redis.Standard)

	SKUFamilyC = string(redis.C)
	SKUFamilyP = string(redis.P)
)

// Resource states
const (
	ProvisioningStateCreating               = string(redis.Creating)
	ProvisioningStateDeleting               = string(redis.Deleting)
	ProvisioningStateDisabled               = string(redis.Disabled)
	ProvisioningStateFailed                 = string(redis.Failed)
	ProvisioningStateLinking                = string(redis.Linking)
	ProvisioningStateProvisioning           = string(redis.Provisioning)
	ProvisioningStateRecoveringScaleFailure = string(redis.RecoveringScaleFailure)
	ProvisioningStateScaling                = string(redis.Scaling)
	ProvisioningStateSucceeded              = string(redis.Succeeded)
	ProvisioningStateUnlinking              = string(redis.Unlinking)
	ProvisioningStateUnprovisioning         = string(redis.Unprovisioning)
	ProvisioningStateUpdating               = string(redis.Updating)
)

const (
	// SupportedRedisVersion is the only minor version of Redis currently
	// supported by Azure Cache for Redis. The version cannot be specified at
	// creation time.
	SupportedRedisVersion = "3.2"
)

// RedisParameters define the desired state of an Azure Redis cluster. Most
// fields map directly to an Azure Redis resource:
// https://docs.microsoft.com/en-us/rest/api/redis/redis/create#redisresource
type RedisParameters struct {
	// ResourceGroupName in which to create this resource.
	ResourceGroupName string `json:"resourceGroupName"`

	// Location in which to create this resource.
	Location string `json:"location"`

	// SKU of the Redis cache to deploy.
	SKU SKUSpec `json:"sku"`

	// EnableNonSSLPort specifies whether the non-ssl Redis server port (6379)
	// is enabled.
	// +optional
	EnableNonSSLPort bool `json:"enableNonSslPort,omitempty"`

	// ShardCount specifies the number of shards to be created on a Premium
	// Cluster Cache.
	// +optional
	ShardCount int `json:"shardCount,omitempty"`

	// StaticIP address. Required when deploying a Redis cache inside an
	// existing Azure Virtual Network.
	// +optional
	StaticIP string `json:"staticIP,omitempty"`

	// SubnetID specifies the full resource ID of a subnet in a virtual network
	// to deploy the Redis cache in. Example format:
	// /subscriptions/{subscriptionId}/resourceGroups/{resourceGroupName}/Microsoft.{Network|ClassicNetwork}/VirtualNetworks/vnet1/subnets/subnet1
	// +optional
	SubnetID string `json:"subnetId,omitempty"`

	// RedisConfiguration specifies Redis Settings.
	// +optional
	RedisConfiguration map[string]string `json:"redisConfiguration,omitempty"`
}

// TODO(negz): Rename SKU to PricingTier? Both SQL databases and Redis caches
// call this an 'SKU' in their API, but we call it a PricingTier in our Azure
// SQL database CRD.

// An SKUSpec represents the performance and cost oriented properties of a
// Redis.
type SKUSpec struct {
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

// A RedisSpec defines the desired state of a Redis.
type RedisSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	RedisParameters              `json:",inline"`
}

// A RedisStatus represents the observed state of a Redis.
type RedisStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`

	// State represents the state of an Azure Redis.
	State string `json:"state,omitempty"`

	// ProviderID is the external ID to identify this resource in the cloud
	// provider.
	ProviderID string `json:"providerID,omitempty"`

	// Endpoint of the Redis resource used in connection strings.
	Endpoint string `json:"endpoint,omitempty"`

	// Port at which the Redis endpoint is listening.
	Port int `json:"port,omitempty"`

	// SSLPort at which the Redis endpoint is listening.
	SSLPort int `json:"sslPort,omitempty"`

	// RedisVersion the Redis endpoint is running.
	RedisVersion string `json:"redisVersion,omitempty"`

	// ResourceName of the Redis cache resource.
	ResourceName string `json:"resourceName,omitempty"`
}

// +kubebuilder:object:root=true

// A Redis is a managed resource that represents an Azure Redis cluster.
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.state"
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
	RedisParameters                   `json:",inline"`
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
