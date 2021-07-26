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

package agentpool

import (
	"reflect"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-03-01/containerservice"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

func agentPoolFromParams(p v1alpha3.AgentPoolParameters) *v1alpha3.AgentPool {
	return &v1alpha3.AgentPool{Spec: v1alpha3.AgentPoolSpec{AgentPoolParameters: p}}
}

func agentPoolFromStatus(s v1alpha3.AgentPoolStatus) *v1alpha3.AgentPool {
	return &v1alpha3.AgentPool{Status: s}
}

func aksClusterFromParams(p v1alpha3.AKSClusterParameters) *v1alpha3.AKSCluster {
	return &v1alpha3.AKSCluster{Spec: v1alpha3.AKSClusterSpec{AKSClusterParameters: p}}
}

func azAgentPoolFromParams(p v1alpha3.AgentPoolParameters) *containerservice.AgentPool {
	az := New(agentPoolFromParams(p))
	return &az
}

func TestNeedUpdate(t *testing.T) {
	tests := []struct {
		name       string
		c          *v1alpha3.AgentPool
		az         *containerservice.AgentPool
		needUpdate bool
	}{{
		name:       "not changed",
		c:          agentPoolFromParams(v1alpha3.AgentPoolParameters{}),
		az:         azAgentPoolFromParams(v1alpha3.AgentPoolParameters{}),
		needUpdate: false,
	}, {
		name: "VMSize changed",
		c: agentPoolFromParams(v1alpha3.AgentPoolParameters{
			NodeVMSize: "Standard_B2s",
		}),
		az: azAgentPoolFromParams(v1alpha3.AgentPoolParameters{
			NodeVMSize: "Standard_D2s",
		}),
		needUpdate: true,
	}, {
		name: "NodeCount changed",
		c: agentPoolFromParams(v1alpha3.AgentPoolParameters{
			NodeCount: azure.ToInt32Ptr(1),
		}),
		az: azAgentPoolFromParams(v1alpha3.AgentPoolParameters{
			NodeCount: azure.ToInt32Ptr(2),
		}),
		needUpdate: true,
	}, {
		name: "NodeCount changed, but MinNodeCount and MaxNodeCount not changed with autoscaling enabled",
		c: agentPoolFromParams(v1alpha3.AgentPoolParameters{
			EnableAutoScaling: azure.ToBoolPtr(true),
			NodeCount:         azure.ToInt32Ptr(1),
			MaxNodeCount:      azure.ToInt32Ptr(3),
			MinNodeCount:      azure.ToInt32Ptr(1),
		}),
		az: azAgentPoolFromParams(v1alpha3.AgentPoolParameters{
			EnableAutoScaling: azure.ToBoolPtr(true),
			NodeCount:         azure.ToInt32Ptr(2),
			MaxNodeCount:      azure.ToInt32Ptr(3),
			MinNodeCount:      azure.ToInt32Ptr(1),
		}),
		needUpdate: false,
	}, {
		name: "MaxNodeCount changed",
		c: agentPoolFromParams(v1alpha3.AgentPoolParameters{
			NodeCount:    azure.ToInt32Ptr(1),
			MaxNodeCount: azure.ToInt32Ptr(4),
			MinNodeCount: azure.ToInt32Ptr(1),
		}),
		az: azAgentPoolFromParams(v1alpha3.AgentPoolParameters{
			NodeCount:    azure.ToInt32Ptr(2),
			MaxNodeCount: azure.ToInt32Ptr(3),
			MinNodeCount: azure.ToInt32Ptr(1),
		}),
		needUpdate: true,
	}, {
		name: "MaxNodeCount changed",
		c: agentPoolFromParams(v1alpha3.AgentPoolParameters{
			NodeCount:    azure.ToInt32Ptr(1),
			MaxNodeCount: azure.ToInt32Ptr(3),
			MinNodeCount: azure.ToInt32Ptr(1),
		}),
		az: azAgentPoolFromParams(v1alpha3.AgentPoolParameters{
			NodeCount:    azure.ToInt32Ptr(2),
			MaxNodeCount: azure.ToInt32Ptr(3),
			MinNodeCount: azure.ToInt32Ptr(2),
		}),
		needUpdate: true,
	}, {
		name: "VnetSubnetID changed",
		c: agentPoolFromParams(v1alpha3.AgentPoolParameters{
			VnetSubnetID: "a",
		}),
		az: azAgentPoolFromParams(v1alpha3.AgentPoolParameters{
			VnetSubnetID: "b",
		}),
		needUpdate: true,
	}, {
		name: "EnableAutoScaling changed",
		c: agentPoolFromParams(v1alpha3.AgentPoolParameters{
			EnableAutoScaling: azure.ToBoolPtr(false),
		}),
		az: azAgentPoolFromParams(v1alpha3.AgentPoolParameters{
			EnableAutoScaling: azure.ToBoolPtr(true),
		}),
		needUpdate: true,
	}, {
		name: "NodeTaints changed",
		c: agentPoolFromParams(v1alpha3.AgentPoolParameters{
			NodeTaints: []string{"a", "b"},
		}),
		az: azAgentPoolFromParams(v1alpha3.AgentPoolParameters{
			NodeTaints: []string{"b", "c"},
		}),
		needUpdate: true,
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			needUpdate := NeedUpdate(tt.c, tt.az)
			if needUpdate != tt.needUpdate {
				t.Errorf("NeedUpdate() got = %t, want %t", needUpdate, tt.needUpdate)
			}
		})
	}
}

func TestNewProfile(t *testing.T) {
	tests := []struct {
		name string
		c    *v1alpha3.AKSCluster
		want containerservice.ManagedClusterAgentPoolProfile
	}{{
		name: "simple",
		c: aksClusterFromParams(v1alpha3.AKSClusterParameters{
			ResourceGroupName: "rg",
			VnetSubnetID:      "sub",
			NodeCount:         to.IntPtr(1),
			NodeVMSize:        "Standard_B2s",
		}),
		want: containerservice.ManagedClusterAgentPoolProfile{
			Name:         azure.ToStringPtr(v1alpha3.AgentPoolProfileName),
			Count:        azure.ToInt32Ptr(1),
			VMSize:       "Standard_B2s",
			VnetSubnetID: azure.ToStringPtr("sub"),
			Type:         "VirtualMachineScaleSets",
			Mode:         "System",
		},
	}, {
		name: "empty subnet",
		c: aksClusterFromParams(v1alpha3.AKSClusterParameters{
			ResourceGroupName: "rg",
			NodeCount:         to.IntPtr(1),
			NodeVMSize:        "Standard_B2s",
		}),
		want: containerservice.ManagedClusterAgentPoolProfile{
			Name:   azure.ToStringPtr(v1alpha3.AgentPoolProfileName),
			Count:  azure.ToInt32Ptr(1),
			VMSize: "Standard_B2s",
			Type:   "VirtualMachineScaleSets",
			Mode:   "System",
		},
	}, {
		name: "default node count",
		c: aksClusterFromParams(v1alpha3.AKSClusterParameters{
			ResourceGroupName: "rg",
			VnetSubnetID:      "sub",
			NodeVMSize:        "Standard_B2s",
		}),
		want: containerservice.ManagedClusterAgentPoolProfile{
			Name:         azure.ToStringPtr(v1alpha3.AgentPoolProfileName),
			Count:        azure.ToInt32Ptr(v1alpha3.DefaultNodeCount),
			VMSize:       "Standard_B2s",
			VnetSubnetID: azure.ToStringPtr("sub"),
			Type:         "VirtualMachineScaleSets",
			Mode:         "System",
		},
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProfile(tt.c); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProfile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	tests := []struct {
		name string
		c    *v1alpha3.AgentPool
		az   *containerservice.AgentPool
		want *v1alpha3.AgentPool
	}{{
		name: "init",
		c:    agentPoolFromStatus(v1alpha3.AgentPoolStatus{}),
		az: &containerservice.AgentPool{
			ID: azure.ToStringPtr("id"),
			ManagedClusterAgentPoolProfileProperties: &containerservice.ManagedClusterAgentPoolProfileProperties{
				Count:             azure.ToInt32Ptr(1),
				ProvisioningState: azure.ToStringPtr("state"),
			},
		},
		want: agentPoolFromStatus(v1alpha3.AgentPoolStatus{
			State:      "state",
			NodesCount: 1,
			ProviderID: "id",
		}),
	}, {
		name: "update",
		c: agentPoolFromStatus(v1alpha3.AgentPoolStatus{
			State:      "before",
			NodesCount: 10,
			ProviderID: "before",
		}),
		az: &containerservice.AgentPool{
			ID: azure.ToStringPtr("id"),
			ManagedClusterAgentPoolProfileProperties: &containerservice.ManagedClusterAgentPoolProfileProperties{
				Count:             azure.ToInt32Ptr(1),
				ProvisioningState: azure.ToStringPtr("state"),
			},
		},
		want: agentPoolFromStatus(v1alpha3.AgentPoolStatus{
			State:      "state",
			NodesCount: 1,
			ProviderID: "id",
		}),
	}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			UpdateStatus(tt.c, tt.az)
		})
		if !reflect.DeepEqual(tt.c, tt.want) {
			t.Errorf("After UpdateStatus() AgentPool = %v, want %v", tt.c, tt.want)
		}
	}
}
