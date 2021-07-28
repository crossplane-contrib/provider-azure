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

package registry

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2019-05-01/containerregistry"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		r    v1alpha3.RegistrySpec
		want containerregistry.Registry
	}{{
		name: "simple",
		r: v1alpha3.RegistrySpec{
			AdminUserEnabled: true,
			Sku:              "sku",
			Location:         "location",
		},
		want: containerregistry.Registry{
			Sku: &containerregistry.Sku{
				Name: "sku",
				Tier: "sku",
			},
			RegistryProperties: &containerregistry.RegistryProperties{AdminUserEnabled: azure.ToBoolPtr(true)},
			Location:           azure.ToStringPtr("location"),
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpToDate(t *testing.T) {
	tests := []struct {
		name string
		r    *v1alpha3.RegistrySpec
		az   *containerregistry.Registry
		want bool
	}{{
		name: "up to date",
		r: &v1alpha3.RegistrySpec{
			AdminUserEnabled: true,
			Sku:              "sku",
			Location:         "location",
		},
		az: &containerregistry.Registry{
			Sku: &containerregistry.Sku{
				Name: "sku",
				Tier: "sku",
			},
			RegistryProperties: &containerregistry.RegistryProperties{AdminUserEnabled: azure.ToBoolPtr(true)},
			Location:           azure.ToStringPtr("location"),
		},
		want: true,
	}, {
		name: "need update sku",
		r: &v1alpha3.RegistrySpec{
			AdminUserEnabled: true,
			Sku:              "new sku",
			Location:         "location",
		},
		az: &containerregistry.Registry{
			Sku: &containerregistry.Sku{
				Name: "sku",
				Tier: "sku",
			},
			RegistryProperties: &containerregistry.RegistryProperties{AdminUserEnabled: azure.ToBoolPtr(true)},
			Location:           azure.ToStringPtr("location"),
		},
		want: false,
	}, {
		name: "need update adminUserEnabled",
		r: &v1alpha3.RegistrySpec{
			AdminUserEnabled: false,
			Sku:              "sku",
			Location:         "location",
		},
		az: &containerregistry.Registry{
			Sku: &containerregistry.Sku{
				Name: "sku",
				Tier: "sku",
			},
			RegistryProperties: &containerregistry.RegistryProperties{AdminUserEnabled: azure.ToBoolPtr(true)},
			Location:           azure.ToStringPtr("location"),
		},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UpToDate(tt.r, tt.az); got != tt.want {
				t.Errorf("UpToDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInitialized(t *testing.T) {
	tests := []struct {
		name string
		r    *v1alpha3.Registry
		want bool
	}{{
		name: "Succeeded",
		r: &v1alpha3.Registry{Status: v1alpha3.RegistryStatus{
			State: "Succeeded",
		}},
		want: true,
	}, {
		name: "Failed",
		r: &v1alpha3.Registry{Status: v1alpha3.RegistryStatus{
			State: "Failed",
		}},
		want: true,
	}, {
		name: "Canceled",
		r: &v1alpha3.Registry{Status: v1alpha3.RegistryStatus{
			State: "Canceled",
		}},
		want: true,
	}, {
		name: "Creating",
		r: &v1alpha3.Registry{Status: v1alpha3.RegistryStatus{
			State: "Creating",
		}},
		want: false,
	}, {
		name: "Updating",
		r: &v1alpha3.Registry{Status: v1alpha3.RegistryStatus{
			State: "Updating",
		}},
		want: false,
	}, {
		name: "Deleting",
		r: &v1alpha3.Registry{Status: v1alpha3.RegistryStatus{
			State: "Deleting",
		}},
		want: false,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Initialized(tt.r); got != tt.want {
				t.Errorf("Initialized() = %v, want %v", got, tt.want)
			}
		})
	}
}
