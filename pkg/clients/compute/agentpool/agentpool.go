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
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-03-01/containerservice"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

// New returns an Azure AgentPool object from a AgentPool spec
func New(c *v1alpha3.AgentPool) containerservice.AgentPool {
	p := containerservice.AgentPool{
		ManagedClusterAgentPoolProfileProperties: &containerservice.ManagedClusterAgentPoolProfileProperties{
			VMSize:            containerservice.VMSizeTypes(c.Spec.NodeVMSize),
			EnableAutoScaling: c.Spec.EnableAutoScaling,
			// AgentPool APIs supported only for VMSS agentpools.
			// For more information, please check https://aka.ms/multiagentpoollimitations
			Type: containerservice.VirtualMachineScaleSets,
			// Defaults override below if needed
			Mode: containerservice.User,
		},
	}
	if c.Spec.Mode != "" {
		p.ManagedClusterAgentPoolProfileProperties.Mode = containerservice.AgentPoolMode(c.Spec.Mode)
	}
	if len(c.Spec.AvailabilityZones) > 0 {
		p.AvailabilityZones = &c.Spec.AvailabilityZones
	}
	if c.Spec.NodeCount != nil {
		p.ManagedClusterAgentPoolProfileProperties.Count = c.Spec.NodeCount
	}
	if c.Spec.MinNodeCount != nil {
		p.ManagedClusterAgentPoolProfileProperties.MinCount = c.Spec.MinNodeCount
	}
	if c.Spec.MaxNodeCount != nil {
		p.ManagedClusterAgentPoolProfileProperties.MaxCount = c.Spec.MaxNodeCount
	}
	if c.Spec.VnetSubnetID != "" {
		p.VnetSubnetID = to.StringPtr(c.Spec.VnetSubnetID)
	}
	if len(c.Spec.NodeTaints) > 0 {
		p.NodeTaints = &c.Spec.NodeTaints
	}
	return p
}

// NewProfile returns an Azure ManagedClusterAgentPoolProfile object from a AKSCluster spec
func NewProfile(c *v1alpha3.AKSCluster) containerservice.ManagedClusterAgentPoolProfile {
	nodeCount := int32(v1alpha3.DefaultNodeCount)
	if c.Spec.NodeCount != nil {
		nodeCount = int32(*c.Spec.NodeCount)
	}
	p := containerservice.ManagedClusterAgentPoolProfile{
		Name:   to.StringPtr(v1alpha3.AgentPoolProfileName),
		Count:  &nodeCount,
		VMSize: containerservice.VMSizeTypes(c.Spec.NodeVMSize),
		Mode:   containerservice.System,
		Type:   containerservice.VirtualMachineScaleSets,
	}
	if c.Spec.VnetSubnetID != "" {
		p.VnetSubnetID = to.StringPtr(c.Spec.VnetSubnetID)
	}
	return p
}

// NeedUpdate determines if a AgentPool need to be updated
func NeedUpdate(c *v1alpha3.AgentPool, az *containerservice.AgentPool) bool {
	if c.Spec.NodeVMSize != string(az.VMSize) {
		return true
	}
	if c.Spec.VnetSubnetID != "" {
		azureValue := azure.ToString(az.VnetSubnetID)
		specValue := c.Spec.VnetSubnetID
		if azureValue != specValue {
			return true
		}
	}
	if azure.ToBool(c.Spec.EnableAutoScaling) != azure.ToBool(az.EnableAutoScaling) {
		return true
	}
	if nodeCountNeedUpdate(c, az) {
		return true
	}
	if az.NodeTaints != nil {
		azTaints := *az.NodeTaints
		if !cmp.Equal(azTaints, c.Spec.NodeTaints) {
			return true
		}
	}
	if az.NodeTaints == nil && len(c.Spec.NodeTaints) > 0 {
		return true
	}
	return false
}

func nodeCountNeedUpdate(c *v1alpha3.AgentPool, az *containerservice.AgentPool) bool {
	// For enabled autoscaling NodeCount is read-only field.
	// And can differ from min to max on different requests.
	// When autoscaling disabled NodeCount is read-write value.
	// And represent static node count of agent pool.
	if azure.ToBool(c.Spec.EnableAutoScaling) {
		azureCount := azure.ToInt(az.MinCount)
		specCount := azure.ToInt(c.Spec.MinNodeCount)
		if azureCount != specCount {
			return true
		}
		azureCount = azure.ToInt(az.MaxCount)
		specCount = azure.ToInt(c.Spec.MaxNodeCount)
		if azureCount != specCount {
			return true
		}
	} else {
		azureCount := azure.ToInt(az.Count)
		specCount := azure.ToInt(c.Spec.NodeCount)
		if azureCount != specCount {
			return true
		}
	}
	return false
}

// UpdateStatus updates the status related to the external
// Azure virtual network in the AgentPoolStatus
func UpdateStatus(c *v1alpha3.AgentPool, az *containerservice.AgentPool) {
	c.Status.ProviderID = to.String(az.ID)
	c.Status.State = to.String(az.ProvisioningState)
	c.Status.NodesCount = int(to.Int32(az.Count))
	if az.AvailabilityZones != nil {
		c.Status.AvailabilityZones = *az.AvailabilityZones
	} else {
		c.Status.AvailabilityZones = nil
	}
}
