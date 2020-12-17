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
package network

import (
	networkmgmt "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network/networkapi"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

// A GroupsClient handles CRUD operations for Azure Security Group resources.
type GroupsClient networkapi.SecurityGroupsClientAPI

// UpdateSecurityGroupStatusFromAzure updates the status related to the external
// Azure Security Group in the SecurityGroupStatus
func UpdateSecurityGroupStatusFromAzure(v *v1alpha3.SecurityGroup, az networkmgmt.SecurityGroup) {
	v.Status.State = azure.ToString(az.ProvisioningState)
	v.Status.ID = azure.ToString(az.ID)
	v.Status.Etag = azure.ToString(az.Etag)
	v.Status.ResourceGUID = azure.ToString(az.ResourceGUID)
	v.Status.Type = azure.ToString(az.Type)
}

// NewSecurityGroupParameters returns an Azure SecurityGroup object from a Security Group Spec
func NewSecurityGroupParameters(v *v1alpha3.SecurityGroup) networkmgmt.SecurityGroup {
	return networkmgmt.SecurityGroup{
		Location: azure.ToStringPtr(v.Spec.Location),
		Tags:     azure.ToStringPtrMap(v.Spec.Tags),
		SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{
			// Default spec changes will be added if needed here
			SecurityRules: SetSecurityRulesToSecurityGroup(v.Spec.SecurityGroupPropertiesFormat.SecurityRules),
		},
	}
}

func SetSecurityRulesToSecurityGroup(vList *[]v1alpha3.SecurityRule) *[]networkmgmt.SecurityRule {
	var vSecList = new([]networkmgmt.SecurityRule)
	if vList != nil {
		for _, v := range *vList {
			var sRule = networkmgmt.SecurityRule{}
			sRule.ID = azure.ToStringPtr(v.ID)
			sRule.Name = azure.ToStringPtr(v.Name)
			sRule.Etag = azure.ToStringPtr(v.Etag)
			var ruleProperties = new(networkmgmt.SecurityRulePropertiesFormat)
			ruleProperties.Description = azure.ToStringPtr(v.Properties.Description)
			ruleProperties.Protocol = setSecurityRuleProtocol(v.Properties.Protocol)
			ruleProperties.Access = setSecurityRuleAccess(v.Properties.Access)
			ruleProperties.ProvisioningState = azure.ToStringPtr(v.Properties.ProvisioningState)
			ruleProperties.SourcePortRange = azure.ToStringPtr(v.Properties.SourcePortRange)
			ruleProperties.DestinationPortRange = azure.ToStringPtr(v.Properties.DestinationPortRange)
			ruleProperties.SourcePortRanges = azure.ToStringArrayPtr(v.Properties.SourcePortRanges)
			ruleProperties.DestinationPortRanges = azure.ToStringArrayPtr(v.Properties.DestinationPortRanges)
			ruleProperties.SourceAddressPrefix = azure.ToStringPtr(v.Properties.SourceAddressPrefix)
			ruleProperties.DestinationAddressPrefix = azure.ToStringPtr(v.Properties.DestinationAddressPrefix)
			ruleProperties.SourceAddressPrefixes = azure.ToStringArrayPtr(v.Properties.SourceAddressPrefixes)
			ruleProperties.DestinationAddressPrefixes = azure.ToStringArrayPtr(v.Properties.DestinationAddressPrefixes)
			ruleProperties.Direction = setSecurityRuleDirection(v.Properties.Direction)
			ruleProperties.Priority = azure.ToInt32Ptr(int(v.Properties.Priority))
			ruleProperties.SourceApplicationSecurityGroups = setApplicationSecurityGroups(&v.Properties.SourceApplicationSecurityGroups)
			ruleProperties.DestinationApplicationSecurityGroups = setApplicationSecurityGroups(&v.Properties.DestinationApplicationSecurityGroups)
			sRule.SecurityRulePropertiesFormat = ruleProperties
			*vSecList = append(*vSecList, sRule)
		}
		return vSecList
	}
	return nil
}

func setApplicationSecurityGroups(groups *[]v1alpha3.ApplicationSecurityGroup) *[]networkmgmt.ApplicationSecurityGroup {
	var appSecurityGroups = new([]networkmgmt.ApplicationSecurityGroup)
	for _, v := range *groups {
		var applicationSecurityGroup = networkmgmt.ApplicationSecurityGroup{}
		applicationSecurityGroup.ID = azure.ToStringPtr(v.ID)
		applicationSecurityGroup.Name = azure.ToStringPtr(v.Name)
		applicationSecurityGroup.Location = azure.ToStringPtr(v.Location)
		applicationSecurityGroup.Type = azure.ToStringPtr(v.Type)
		applicationSecurityGroup.Etag = azure.ToStringPtr(v.Etag)
		var applicationSecurityGroupPropertiesFormat = new(networkmgmt.ApplicationSecurityGroupPropertiesFormat)
		applicationSecurityGroupPropertiesFormat.ResourceGUID = azure.ToStringPtr(v.Properties.ResourceGUID)
		applicationSecurityGroupPropertiesFormat.ProvisioningState = azure.ToStringPtr(v.Properties.ProvisioningState)
		applicationSecurityGroup.ApplicationSecurityGroupPropertiesFormat = applicationSecurityGroupPropertiesFormat
		*appSecurityGroups = append(*appSecurityGroups, applicationSecurityGroup)
	}
	return appSecurityGroups
}

func setSecurityRuleDirection(direction v1alpha3.SecurityRuleDirection) networkmgmt.SecurityRuleDirection {
	return networkmgmt.SecurityRuleDirection(direction)
}

func setSecurityRuleAccess(access v1alpha3.SecurityRuleAccess) networkmgmt.SecurityRuleAccess {
	return networkmgmt.SecurityRuleAccess(access)
}

func setSecurityRuleProtocol(protocol v1alpha3.SecurityRuleProtocol) networkmgmt.SecurityRuleProtocol {
	return networkmgmt.SecurityRuleProtocol(protocol)
}

// SecurityGroupNeedsUpdate determines if a Security Group need to be updated
func SecurityGroupNeedsUpdate(sg *v1alpha3.SecurityGroup, az networkmgmt.SecurityGroup) bool {
	up := NewSecurityGroupParameters(sg)
	if sg.Spec.SecurityRules != nil && az.SecurityRules != nil {
		sgRules := SetSecurityRulesToSecurityGroup(sg.Spec.SecurityRules)
		azSgRules := az.SecurityRules
		if !reflect.DeepEqual(len(*sgRules), len(*azSgRules)) {
			return true
		}
		for _, rule := range *sgRules {
			for _, azRule := range *azSgRules {
				if reflect.DeepEqual(rule.Name, azRule.Name) {
					if !reflect.DeepEqual(rule.Etag, azRule.Etag) {
						return true
					}
				}
			}
		}
	}
	if nil == sg.Spec.SecurityRules && nil != az.SecurityRules {
		return true
	}
	if nil != sg.Spec.SecurityRules && nil == az.SecurityRules {
		return true
	}
	if !reflect.DeepEqual(up.Tags, az.Tags) {
		return true
	}
	if !reflect.DeepEqual(azure.ToStringPtr(sg.Spec.Location), az.Location) {
		return true
	}
	if !reflect.DeepEqual(azure.ToStringPtr(sg.Name), az.Name) {
		return true
	}
	return false
}
