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
	"testing"

	redismgmt "github.com/Azure/azure-sdk-for-go/services/redis/mgmt/2018-03-01/redis"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane-contrib/provider-azure/apis/cache/v1beta1"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
)

const (
	skuName     = "basic"
	skuFamily   = "C"
	skuCapacity = 1
)

var (
	location           = "us-east1"
	zones              = []string{"us-east1a", "us-east1b"}
	tags               = map[string]string{"key1": "val1"}
	tags2              = map[string]string{"key1": "val1", "key2": "val2"}
	enableNonSSLPort   = true
	subnetID           = "coolsubnet"
	staticIP           = "172.16.0.1"
	shardCount         = 3
	redisConfiguration = map[string]string{"cool": "socool"}
	tenantSettings     = map[string]string{"tenant1": "is-crazy"}
	minTLSVersion      = "1.1"

	redisVersion  = "3.2"
	hostName      = "108.8.8.1"
	port          = 6374
	sslPort       = 453
	linkedServers = []string{"server1", "server2"}
	resourceName  = "some-name"
	resourceID    = "23123"
)

func TestNewCreateParameters(t *testing.T) {
	cases := []struct {
		name string
		r    *v1beta1.Redis
		want redismgmt.CreateParameters
	}{
		{
			name: "Successful",
			r: &v1beta1.Redis{
				Spec: v1beta1.RedisSpec{
					ForProvider: v1beta1.RedisParameters{
						Location: location,
						Zones:    zones,
						Tags:     tags,
						SKU: v1beta1.SKU{
							Name:     skuName,
							Family:   skuFamily,
							Capacity: skuCapacity,
						},
						SubnetID:           &subnetID,
						StaticIP:           &staticIP,
						EnableNonSSLPort:   &enableNonSSLPort,
						RedisConfiguration: redisConfiguration,
						TenantSettings:     tenantSettings,
						ShardCount:         &shardCount,
						MinimumTLSVersion:  &minTLSVersion,
					},
				},
			},
			want: redismgmt.CreateParameters{
				Location: azure.ToStringPtr(location),
				Zones:    azure.ToStringArrayPtr(zones),
				Tags:     azure.ToStringPtrMap(tags),
				CreateProperties: &redismgmt.CreateProperties{
					Sku: &redismgmt.Sku{
						Name:     redismgmt.SkuName(skuName),
						Family:   redismgmt.SkuFamily(skuFamily),
						Capacity: azure.ToInt32Ptr(skuCapacity),
					},
					SubnetID:           azure.ToStringPtr(subnetID),
					StaticIP:           azure.ToStringPtr(staticIP),
					EnableNonSslPort:   azure.ToBoolPtr(enableNonSSLPort),
					RedisConfiguration: azure.ToStringPtrMap(redisConfiguration),
					TenantSettings:     azure.ToStringPtrMap(tenantSettings),
					ShardCount:         azure.ToInt32Ptr(shardCount),
					MinimumTLSVersion:  redismgmt.TLSVersion(minTLSVersion),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewCreateParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewCreateParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestNewUpdateParameters(t *testing.T) {
	redisConfiguration2 := map[string]string{
		"another": "val",
	}
	cases := []struct {
		name    string
		spec    v1beta1.RedisParameters
		current redismgmt.ResourceType
		want    redismgmt.UpdateParameters
	}{
		{
			name: "FullConversion",
			spec: v1beta1.RedisParameters{
				Tags: tags,
				SKU: v1beta1.SKU{
					Name:     skuName,
					Family:   skuFamily,
					Capacity: skuCapacity,
				},
				EnableNonSSLPort:   &enableNonSSLPort,
				RedisConfiguration: redisConfiguration,
				ShardCount:         &shardCount,
				TenantSettings:     tenantSettings,
				MinimumTLSVersion:  &minTLSVersion,
			},
			want: redismgmt.UpdateParameters{
				Tags: azure.ToStringPtrMap(tags),
				UpdateProperties: &redismgmt.UpdateProperties{
					Sku: &redismgmt.Sku{
						Name:     redismgmt.SkuName(skuName),
						Family:   redismgmt.SkuFamily(skuFamily),
						Capacity: azure.ToInt32Ptr(skuCapacity),
					},
					EnableNonSslPort:   azure.ToBoolPtr(enableNonSSLPort),
					RedisConfiguration: azure.ToStringPtrMap(redisConfiguration),
					ShardCount:         azure.ToInt32Ptr(shardCount),
					TenantSettings:     azure.ToStringPtrMap(tenantSettings),
					MinimumTLSVersion:  redismgmt.TLSVersion(minTLSVersion),
				},
			},
		},
		{
			name: "PatchTags",
			spec: v1beta1.RedisParameters{
				Tags: tags,
				SKU: v1beta1.SKU{
					Name:     skuName,
					Family:   skuFamily,
					Capacity: skuCapacity,
				},
				EnableNonSSLPort:   &enableNonSSLPort,
				RedisConfiguration: redisConfiguration,
				ShardCount:         &shardCount,
				TenantSettings:     tenantSettings,
				MinimumTLSVersion:  &minTLSVersion,
			},
			current: redismgmt.ResourceType{
				Properties: &redismgmt.Properties{
					Sku: &redismgmt.Sku{
						Name:     redismgmt.SkuName(skuName),
						Family:   redismgmt.SkuFamily(skuFamily),
						Capacity: azure.ToInt32Ptr(skuCapacity),
					},
					EnableNonSslPort:   azure.ToBoolPtr(enableNonSSLPort),
					RedisConfiguration: azure.ToStringPtrMap(redisConfiguration),
					ShardCount:         azure.ToInt32Ptr(shardCount),
					TenantSettings:     azure.ToStringPtrMap(tenantSettings),
					MinimumTLSVersion:  redismgmt.TLSVersion(minTLSVersion),
				},
			},
			want: redismgmt.UpdateParameters{
				UpdateProperties: &redismgmt.UpdateProperties{},
				Tags:             azure.ToStringPtrMap(tags),
			},
		},
		{
			name: "PatchRedisConfig",
			spec: v1beta1.RedisParameters{
				SKU: v1beta1.SKU{
					Name:     skuName,
					Family:   skuFamily,
					Capacity: skuCapacity,
				},
				EnableNonSSLPort:   &enableNonSSLPort,
				RedisConfiguration: redisConfiguration2,
				ShardCount:         &shardCount,
				TenantSettings:     tenantSettings,
				MinimumTLSVersion:  &minTLSVersion,
			},
			current: redismgmt.ResourceType{
				Properties: &redismgmt.Properties{
					Sku: &redismgmt.Sku{
						Name:     redismgmt.SkuName(skuName),
						Family:   redismgmt.SkuFamily(skuFamily),
						Capacity: azure.ToInt32Ptr(skuCapacity),
					},
					EnableNonSslPort:   azure.ToBoolPtr(enableNonSSLPort),
					RedisConfiguration: azure.ToStringPtrMap(redisConfiguration),
					ShardCount:         azure.ToInt32Ptr(shardCount),
					TenantSettings:     azure.ToStringPtrMap(tenantSettings),
					MinimumTLSVersion:  redismgmt.TLSVersion(minTLSVersion),
				},
			},
			want: redismgmt.UpdateParameters{
				UpdateProperties: &redismgmt.UpdateProperties{
					RedisConfiguration: azure.ToStringPtrMap(redisConfiguration2),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewUpdateParameters(tc.spec, tc.current)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewUpdateParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestNeedsUpdate(t *testing.T) {
	cases := []struct {
		name string
		spec v1beta1.RedisParameters
		az   redismgmt.ResourceType
		want bool
	}{
		{
			name: "DifferentField",
			spec: v1beta1.RedisParameters{
				SKU: v1beta1.SKU{
					Name:     skuName,
					Family:   skuFamily,
					Capacity: skuCapacity,
				},
				EnableNonSSLPort:   &enableNonSSLPort,
				RedisConfiguration: redisConfiguration,
				ShardCount:         &shardCount,
				Tags:               tags2,
			},
			az: redismgmt.ResourceType{
				Tags: azure.ToStringPtrMap(tags),
				Properties: &redismgmt.Properties{
					Sku: &redismgmt.Sku{
						Name:     redismgmt.SkuName(skuName),
						Family:   redismgmt.SkuFamily(skuFamily),
						Capacity: azure.ToInt32Ptr(skuCapacity),
					},
					EnableNonSslPort:   azure.ToBoolPtr(enableNonSSLPort),
					RedisConfiguration: azure.ToStringPtrMap(redisConfiguration),
					ShardCount:         azure.ToInt32Ptr(shardCount + 1),
				},
			},
			want: true,
		},
		{
			name: "NoProperties",
			spec: v1beta1.RedisParameters{
				SKU: v1beta1.SKU{
					Name:     skuName,
					Family:   skuFamily,
					Capacity: skuCapacity,
				},
				EnableNonSSLPort:   &enableNonSSLPort,
				RedisConfiguration: redisConfiguration,
				ShardCount:         &shardCount,
				Tags:               tags,
			},
			az: redismgmt.ResourceType{
				Tags: azure.ToStringPtrMap(tags),
			},
			want: true,
		},
		{
			name: "NeedsNoUpdate",
			spec: v1beta1.RedisParameters{
				SKU: v1beta1.SKU{
					Name:     skuName,
					Family:   skuFamily,
					Capacity: skuCapacity,
				},
				EnableNonSSLPort:   &enableNonSSLPort,
				RedisConfiguration: redisConfiguration,
				ShardCount:         &shardCount,
			},
			az: redismgmt.ResourceType{
				Properties: &redismgmt.Properties{
					Sku: &redismgmt.Sku{
						Name:     redismgmt.SkuName(skuName),
						Family:   redismgmt.SkuFamily(skuFamily),
						Capacity: azure.ToInt32Ptr(skuCapacity),
					},
					EnableNonSslPort:   azure.ToBoolPtr(enableNonSSLPort),
					RedisConfiguration: azure.ToStringPtrMap(redisConfiguration),
					ShardCount:         azure.ToInt32Ptr(shardCount),
				},
			},
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NeedsUpdate(tc.spec, tc.az)
			if got != tc.want {
				t.Errorf("NeedsUpdate(...): want %t, got %t", tc.want, got)
			}
		})
	}
}

func TestGenerateObservation(t *testing.T) {
	cases := map[string]struct {
		arg  redismgmt.ResourceType
		want v1beta1.RedisObservation
	}{
		"FullConversion": {
			arg: redismgmt.ResourceType{
				Name: azure.ToStringPtr(resourceName),
				ID:   azure.ToStringPtr(resourceID),
				Properties: &redismgmt.Properties{
					RedisVersion:      azure.ToStringPtr(redisVersion),
					ProvisioningState: redismgmt.ProvisioningState(ProvisioningStateCreating),
					HostName:          azure.ToStringPtr(hostName),
					Port:              azure.ToInt32(&port),
					LinkedServers: &[]redismgmt.LinkedServer{
						{ID: azure.ToStringPtr(linkedServers[0])},
						{ID: azure.ToStringPtr(linkedServers[1])},
					},
					Sku: &redismgmt.Sku{
						Name:     redismgmt.SkuName(skuName),
						Family:   redismgmt.SkuFamily(skuFamily),
						Capacity: azure.ToInt32Ptr(skuCapacity),
					},
					SubnetID:           azure.ToStringPtr(subnetID),
					StaticIP:           azure.ToStringPtr(staticIP),
					TenantSettings:     azure.ToStringPtrMap(tenantSettings),
					MinimumTLSVersion:  redismgmt.TLSVersion(minTLSVersion),
					EnableNonSslPort:   azure.ToBoolPtr(enableNonSSLPort),
					RedisConfiguration: azure.ToStringPtrMap(redisConfiguration),
					ShardCount:         azure.ToInt32Ptr(shardCount),
					SslPort:            azure.ToInt32(&sslPort),
				},
			},
			want: v1beta1.RedisObservation{
				RedisVersion:      redisVersion,
				ProvisioningState: ProvisioningStateCreating,
				HostName:          hostName,
				Port:              port,
				SSLPort:           sslPort,
				LinkedServers:     linkedServers,
				Name:              resourceName,
				ID:                resourceID,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateObservation(tc.arg)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		az   redismgmt.ResourceType
		spec *v1beta1.RedisParameters
	}
	type want struct {
		spec *v1beta1.RedisParameters
	}
	cases := map[string]struct {
		args
		want
	}{
		"LateInitializeEmptyObject": {
			args: args{
				az: redismgmt.ResourceType{
					Zones: azure.ToStringArrayPtr(zones),
					Tags:  azure.ToStringPtrMap(tags),
					Properties: &redismgmt.Properties{
						RedisVersion:      azure.ToStringPtr(redisVersion),
						ProvisioningState: redismgmt.ProvisioningState(ProvisioningStateCreating),
						HostName:          azure.ToStringPtr(hostName),
						Port:              azure.ToInt32(&port),
						LinkedServers: &[]redismgmt.LinkedServer{
							{ID: azure.ToStringPtr(linkedServers[0])},
							{ID: azure.ToStringPtr(linkedServers[1])},
						},
						Sku: &redismgmt.Sku{
							Name:     redismgmt.SkuName(skuName),
							Family:   redismgmt.SkuFamily(skuFamily),
							Capacity: azure.ToInt32Ptr(skuCapacity),
						},
						SubnetID:           azure.ToStringPtr(subnetID),
						StaticIP:           azure.ToStringPtr(staticIP),
						TenantSettings:     azure.ToStringPtrMap(tenantSettings),
						MinimumTLSVersion:  redismgmt.TLSVersion(minTLSVersion),
						EnableNonSslPort:   azure.ToBoolPtr(enableNonSSLPort),
						RedisConfiguration: azure.ToStringPtrMap(redisConfiguration),
						ShardCount:         azure.ToInt32Ptr(shardCount),
						SslPort:            azure.ToInt32(&sslPort),
					},
				},
				spec: &v1beta1.RedisParameters{},
			},
			want: want{
				spec: &v1beta1.RedisParameters{
					Zones:              zones,
					Tags:               tags,
					SubnetID:           &subnetID,
					StaticIP:           &staticIP,
					EnableNonSSLPort:   &enableNonSSLPort,
					RedisConfiguration: redisConfiguration,
					TenantSettings:     tenantSettings,
					ShardCount:         &shardCount,
					MinimumTLSVersion:  &minTLSVersion,
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(tc.args.spec, tc.args.az)
			if diff := cmp.Diff(tc.want.spec, tc.args.spec); diff != "" {
				t.Errorf("LateInitialize(...): -want, +got\n%s", diff)
			}
		})
	}
}
