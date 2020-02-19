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

	"github.com/crossplane/stack-azure/apis/cache/v1beta1"
	azure "github.com/crossplane/stack-azure/pkg/clients"
)

// Resource states
const (
	ProvisioningStateCreating  = string(redis.Creating)
	ProvisioningStateDeleting  = string(redis.Deleting)
	ProvisioningStateFailed    = string(redis.Failed)
	ProvisioningStateSucceeded = string(redis.Succeeded)
)

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
func NewCreateParameters(cr *v1beta1.Redis) redis.CreateParameters {
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

// NewUpdateParameters returns a redis.UpdateParameters object only with changed
// fields.
// TODO(muvaf): Removal of an entry from the maps such as RedisConfiguration and
// TenantSettings is not properly supported. The user has to give empty string
// for deletion instead of just deleting the whole entry.
// NOTE(muvaf): This is barely a comparison function with almost identical if
// statements which increase the cyclomatic complexity even though it's actually
// easier to maintain all this in one function.
// nolint:gocyclo
func NewUpdateParameters(spec v1beta1.RedisParameters, state redis.ResourceType) redis.UpdateParameters {
	patch := redis.UpdateParameters{
		Tags: azure.ToStringPtrMap(spec.Tags),
		UpdateProperties: &redis.UpdateProperties{
			Sku:                NewSKU(spec.SKU),
			RedisConfiguration: azure.ToStringPtrMap(spec.RedisConfiguration),
			EnableNonSslPort:   spec.EnableNonSSLPort,
			ShardCount:         azure.ToInt32(spec.ShardCount),
			TenantSettings:     azure.ToStringPtrMap(spec.TenantSettings),
			MinimumTLSVersion:  redis.TLSVersion(azure.ToString(spec.MinimumTLSVersion)),
		},
	}
	// NOTE(muvaf): One could possibly generate UpdateParameters object from
	// ResourceType and extract a JSON patch. But since the number of fields
	// are not that many, I wanted to go with if statements. Hopefully, we'll
	// generate this code in the future.
	for k, v := range state.Tags {
		if patch.Tags[k] == v {
			delete(patch.Tags, k)
		}
	}
	if len(patch.Tags) == 0 {
		patch.Tags = nil
	}
	if state.Properties == nil {
		return patch
	}
	if reflect.DeepEqual(patch.Sku, state.Properties.Sku) {
		patch.Sku = nil
	}
	for k, v := range state.RedisConfiguration {
		if reflect.DeepEqual(patch.RedisConfiguration[k], v) {
			delete(patch.RedisConfiguration, k)
		}
	}
	if len(patch.RedisConfiguration) == 0 {
		patch.RedisConfiguration = nil
	}
	if reflect.DeepEqual(patch.EnableNonSslPort, state.EnableNonSslPort) {
		patch.EnableNonSslPort = nil
	}
	if reflect.DeepEqual(patch.ShardCount, state.ShardCount) {
		patch.ShardCount = nil
	}
	for k, v := range state.TenantSettings {
		if reflect.DeepEqual(patch.TenantSettings[k], v) {
			delete(patch.TenantSettings, k)
		}
	}
	if len(patch.TenantSettings) == 0 {
		patch.TenantSettings = nil
	}
	if reflect.DeepEqual(patch.MinimumTLSVersion, state.MinimumTLSVersion) {
		patch.MinimumTLSVersion = ""
	}
	return patch
}

// NewSKU returns a Redis resource SKU suitable for use with the Azure API.
func NewSKU(s v1beta1.SKU) *redis.Sku {
	return &redis.Sku{
		Name:     redis.SkuName(s.Name),
		Family:   redis.SkuFamily(s.Family),
		Capacity: azure.ToInt32Ptr(s.Capacity, azure.FieldRequired),
	}
}

// NeedsUpdate returns true if the supplied spec object differs from the
// supplied Azure resource. It considers only fields that can be modified in
// place without deleting and recreating the instance.
func NeedsUpdate(spec v1beta1.RedisParameters, az redis.ResourceType) bool {
	if az.Properties == nil {
		return true
	}
	patch := NewUpdateParameters(spec, az)
	empty := redis.UpdateParameters{UpdateProperties: &redis.UpdateProperties{}}
	return !reflect.DeepEqual(empty, patch)
}

// GenerateObservation produces a RedisObservation object from the redis.ResourceType
// received from Azure.
func GenerateObservation(az redis.ResourceType) v1beta1.RedisObservation {
	o := v1beta1.RedisObservation{
		ID:   azure.ToString(az.ID),
		Name: azure.ToString(az.Name),
	}
	if az.Properties == nil {
		return o
	}
	o.RedisVersion = azure.ToString(az.Properties.RedisVersion)
	o.ProvisioningState = string(az.Properties.ProvisioningState)
	o.HostName = azure.ToString(az.Properties.HostName)
	o.Port = azure.ToInt(az.Properties.Port)
	o.SSLPort = azure.ToInt(az.Properties.SslPort)
	if az.LinkedServers != nil {
		o.LinkedServers = make([]string, len(*az.Properties.LinkedServers))
		for i, val := range *az.Properties.LinkedServers {
			o.LinkedServers[i] = azure.ToString(val.ID)
		}
	}
	return o
}

// LateInitialize fills the spec values that user did not fill with their
// corresponding value in the Azure, if there is any.
func LateInitialize(spec *v1beta1.RedisParameters, az redis.ResourceType) {
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
