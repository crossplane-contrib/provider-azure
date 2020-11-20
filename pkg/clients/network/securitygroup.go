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
	"encoding/json"
	networkmgmt "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	"reflect"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network/networkapi"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/pkg/errors"

	//"github.com/crossplane/crossplane-runtime/pkg/meta"

	//"github.com/crossplane/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

// A GroupsClient handles CRUD operations for Azure Security Group resources.
type GroupsClient networkapi.SecurityGroupsClientAPI

func NewClient(credentials []byte) (GroupsClient, error) {
	c := azure.Credentials{}
	if err := json.Unmarshal(credentials, &c); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal Azure client secret data")
	}
	client := network.NewSecurityGroupsClient(c.SubscriptionID)

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
			//sRule.Protocol = setSecurityRuleProtocol(v.Properties.Protocol)
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

	if !reflect.DeepEqual(sg.Name, az.Name) {
		return true
	}
	if !reflect.DeepEqual(sg.Spec.Location, az.Location) {
		return true
	}
	if !reflect.DeepEqual(sg.Spec.SecurityRules, az.SecurityRules) {
		return true
	}

	return false
}
