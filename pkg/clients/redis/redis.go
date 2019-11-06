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

package redis

import (
	"context"
	"encoding/json"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/services/redis/mgmt/2018-03-01/redis"
	"github.com/Azure/azure-sdk-for-go/services/redis/mgmt/2018-03-01/redis/redisapi"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/pkg/errors"

	"github.com/crossplaneio/stack-azure/apis/cache/v1alpha3"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
)

// Resource states
const (
	ProvisioningStateCreating  = string(redis.Creating)
	ProvisioningStateDeleting  = string(redis.Deleting)
	ProvisioningStateFailed    = string(redis.Failed)
	ProvisioningStateSucceeded = string(redis.Succeeded)
	ProvisioningStateUpdating  = string(redis.Updating)
)

// A Client handles CRUD operations for Azure Cache resources. This interface is
// compatible with the upstream Azure redis client.
type Client redisapi.ClientAPI

// NewClient returns a new Azure Cache for Redis client. Credentials must be
// passed as JSON encoded data.
func NewClient(_ context.Context, credentials []byte) (redisapi.ClientAPI, error) {
	c := azure.Credentials{}
	if err := json.Unmarshal(credentials, &c); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal Azure client secret data")
	}
	client := redis.NewClient(c.SubscriptionID)

	cfg := auth.ClientCredentialsConfig{
		ClientID:     c.ClientID,
		ClientSecret: c.ClientSecret,
		TenantID:     c.TenantID,
		AADEndpoint:  c.ActiveDirectoryEndpointURL,
		Resource:     c.ResourceManagerEndpointURL,
	}
	a, err := cfg.Authorizer()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create Azure authorizer from credentials config")
	}
	client.Authorizer = a
	if err := client.AddToUserAgent(azure.UserAgent); err != nil {
		return nil, errors.Wrap(err, "cannot add to Azure client user agent")
	}

	return client, nil
}

// NewCreateParameters returns Redis resource creation parameters suitable for
// use with the Azure API.
func NewCreateParameters(cr *v1alpha3.Redis) redis.CreateParameters {
	return redis.CreateParameters{
		Location: azure.ToStringPtr(cr.Spec.ForProvider.Location),
		Zones:    azure.ToStringArrayPtr(cr.Spec.ForProvider.Zones),
		Tags:     azure.ToStringPtrMap(cr.Spec.ForProvider.Tags),
		CreateProperties: &redis.CreateProperties{
			Sku:                NewSKU(cr.Spec.ForProvider.SKU),
			SubnetID:           cr.Spec.ForProvider.SubnetID,
			StaticIP:           cr.Spec.ForProvider.StaticIP,
			EnableNonSslPort:   cr.Spec.ForProvider.EnableNonSSLPort,
			RedisConfiguration: azure.ToStringPtrMap(cr.Spec.ForProvider.RedisConfiguration),
			TenantSettings:     azure.ToStringPtrMap(cr.Spec.ForProvider.TenantSettings),
			ShardCount:         azure.ToInt32(cr.Spec.ForProvider.ShardCount),
			MinimumTLSVersion:  redis.TLSVersion(azure.ToString(cr.Spec.ForProvider.MinimumTLSVersion)),
		},
	}
}

// NewUpdateParameters returns Redis resource update parameters suitable for use
// with the Azure API.
func NewUpdateParameters(cr *v1alpha3.Redis) redis.UpdateParameters {
	return redis.UpdateParameters{
		Tags: azure.ToStringPtrMap(cr.Spec.ForProvider.Tags),
		UpdateProperties: &redis.UpdateProperties{
			Sku:                NewSKU(cr.Spec.ForProvider.SKU),
			RedisConfiguration: azure.ToStringPtrMap(cr.Spec.ForProvider.RedisConfiguration),
			EnableNonSslPort:   cr.Spec.ForProvider.EnableNonSSLPort,
			ShardCount:         azure.ToInt32(cr.Spec.ForProvider.ShardCount),
			TenantSettings:     azure.ToStringPtrMap(cr.Spec.ForProvider.TenantSettings),
			MinimumTLSVersion:  redis.TLSVersion(azure.ToString(cr.Spec.ForProvider.MinimumTLSVersion)),
		},
	}
}

// NewSKU returns a Redis resource SKU suitable for use with the Azure API.
func NewSKU(s v1alpha3.SKU) *redis.Sku {
	return &redis.Sku{
		Name:     redis.SkuName(s.Name),
		Family:   redis.SkuFamily(s.Family),
		Capacity: azure.ToInt32Ptr(s.Capacity, azure.FieldRequired),
	}
}

// NeedsUpdate returns true if the supplied Kubernetes resource differs from the
// supplied Azure resource. It considers only fields that can be modified in
// place without deleting and recreating the instance.
func NeedsUpdate(kube *v1alpha3.Redis, az redis.ResourceType) bool {
	if az.Properties == nil {
		return true
	}
	up := NewUpdateParameters(kube)

	switch {
	case !reflect.DeepEqual(up.Sku, az.Sku):
		return true
	case !reflect.DeepEqual(up.RedisConfiguration, az.RedisConfiguration):
		return true
	case !reflect.DeepEqual(up.EnableNonSslPort, az.EnableNonSslPort):
		return true
	case !reflect.DeepEqual(up.ShardCount, az.ShardCount):
		return true
	case !reflect.DeepEqual(up.TenantSettings, az.TenantSettings):
		return true
	case !reflect.DeepEqual(up.MinimumTLSVersion, az.MinimumTLSVersion):
		return true
	case !reflect.DeepEqual(up.Tags, az.Tags):
		return true
	}

	return false
}

// GenerateObservation produces a RedisObservation object from the redis.ResourceType
// received from Azure.
func GenerateObservation(az redis.ResourceType) v1alpha3.RedisObservation {
	o := v1alpha3.RedisObservation{
		ID:   azure.ToString(az.ID),
		Name: azure.ToString(az.Name),
	}
	if az.Properties == nil {
		return o
	}
	o.RedisVersion = azure.ToString(az.RedisVersion)
	o.ProvisioningState = string(az.ProvisioningState)
	o.HostName = azure.ToString(az.HostName)
	o.Port = azure.ToInt(az.Port)
	o.SSLPort = azure.ToInt(az.SslPort)
	if az.LinkedServers != nil {
		o.LinkedServers = make([]string, len(*az.LinkedServers))
		for i, val := range *az.LinkedServers {
			o.LinkedServers[i] = azure.ToString(val.ID)
		}
	}
	return o
}

// LateInitialize fills the spec values that user did not fill with their
// corresponding value in the Azure, if there is any.
func LateInitialize(spec *v1alpha3.RedisParameters, az redis.ResourceType) {
	spec.Zones = azure.LateInitializeStringValArrFromArrPtr(spec.Zones, az.Zones)
	spec.Tags = azure.LateInitializeStringMap(spec.Tags, az.Tags)
	if az.Properties == nil {
		return
	}
	spec.SubnetID = azure.LateInitializeStringPtrFromPtr(spec.SubnetID, az.Properties.SubnetID)
	spec.StaticIP = azure.LateInitializeStringPtrFromPtr(spec.StaticIP, az.Properties.StaticIP)
	spec.RedisConfiguration = azure.LateInitializeStringMap(spec.RedisConfiguration, az.Properties.RedisConfiguration)
	spec.EnableNonSSLPort = azure.LateInitializeBoolPtrFromPtr(spec.EnableNonSSLPort, az.Properties.EnableNonSslPort)
	spec.TenantSettings = azure.LateInitializeStringMap(spec.TenantSettings, az.Properties.TenantSettings)
	spec.ShardCount = azure.LateInitializeIntPtrFromInt32Ptr(spec.ShardCount, az.Properties.ShardCount)
	minTLS := string(az.Properties.MinimumTLSVersion)
	spec.MinimumTLSVersion = azure.LateInitializeStringPtrFromPtr(spec.MinimumTLSVersion, &minTLS)

}
