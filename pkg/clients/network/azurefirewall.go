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
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"reflect"
)

// UpdateAzureFirewallStatusFromAzure updates the status related to the external
// Azure Firewall in the AzureFirewallStatus
func UpdateAzureFirewallStatusFromAzure(v *v1alpha3.AzureFirewall, az networkmgmt.AzureFirewall) {

	v.Status.State = toStringProvisioningState(az.ProvisioningState)
	v.Status.ID = azure.ToString(az.ID)
	v.Status.Etag = azure.ToString(az.Etag)
	v.Status.Type = azure.ToString(az.Type)
}

func toStringProvisioningState(provisioningState networkmgmt.ProvisioningState) string {
	return string(provisioningState)
}

func setHubIpAddresses(addresses *v1alpha3.HubIPAddresses) *networkmgmt.HubIPAddresses {
	var hubIpAddresses = new(networkmgmt.HubIPAddresses)
	if nil != addresses {
		hubIpAddresses.PrivateIPAddress = addresses.PrivateIPAddress
		for _, publicIpAddress := range *addresses.PublicIPAddresses {
			var azureFirewallPublicIPAddress = networkmgmt.AzureFirewallPublicIPAddress{}
			azureFirewallPublicIPAddress.Address = publicIpAddress.Address
			*hubIpAddresses.PublicIPAddresses = append(*hubIpAddresses.PublicIPAddresses, azureFirewallPublicIPAddress)
		}
	}
	return hubIpAddresses
}

func setIPConfigurations(configurations *[]v1alpha3.AzureFirewallIPConfiguration) *[]networkmgmt.AzureFirewallIPConfiguration {
	var azipc = new([]networkmgmt.AzureFirewallIPConfiguration)
	for _, c := range *configurations {
		var config = networkmgmt.AzureFirewallIPConfiguration{}
		config.Etag = c.Etag
		config.ID = c.ID
		config.Name = c.Name
		config.AzureFirewallIPConfigurationPropertiesFormat = new(networkmgmt.AzureFirewallIPConfigurationPropertiesFormat)
		if c.AzureFirewallIPConfigurationPropertiesFormat.PrivateIPAddress != nil {
			config.PrivateIPAddress = c.AzureFirewallIPConfigurationPropertiesFormat.PrivateIPAddress
		}
		if c.AzureFirewallIPConfigurationPropertiesFormat.ProvisioningState != nil {
			config.ProvisioningState = networkmgmt.ProvisioningState(*c.AzureFirewallIPConfigurationPropertiesFormat.ProvisioningState)
		}

		if nil != setSubResource(c.AzureFirewallIPConfigurationPropertiesFormat.PublicIPAddress) {
			config.AzureFirewallIPConfigurationPropertiesFormat.PublicIPAddress = setSubResource(c.AzureFirewallIPConfigurationPropertiesFormat.PublicIPAddress)
		}
		if nil != setSubResource(c.AzureFirewallIPConfigurationPropertiesFormat.Subnet) {
			config.AzureFirewallIPConfigurationPropertiesFormat.Subnet = setSubResource(c.AzureFirewallIPConfigurationPropertiesFormat.Subnet)
		}
		*azipc = append(*azipc, config)
	}
	return azipc
}

func setSubResource(sr *v1alpha3.SubResource) *networkmgmt.SubResource {
	if nil != sr {
		if nil != sr.ID {
			var subResource = new(networkmgmt.SubResource)
			subResource.ID = sr.ID
			return subResource
		}
	}
	return nil
}

func AzureFirewallNeedsUpdate(firewall *v1alpha3.AzureFirewall, az networkmgmt.AzureFirewall) bool {

	if !reflect.DeepEqual(firewall.Name, az.Name) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.Location, az.Location) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.Zones, az.Zones) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.Etag, az.Etag) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.FirewallPolicy, az.FirewallPolicy) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.HubIPAddresses, az.HubIPAddresses) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.VirtualHub, az.VirtualHub) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.Type, az.Type) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.ThreatIntelMode, az.ThreatIntelMode) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.Tags, az.Tags) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.NatRuleCollections, az.AzureFirewallPropertiesFormat.NatRuleCollections) {
		return true
	}
	if !reflect.DeepEqual(firewall.Spec.NetworkRuleCollections, az.AzureFirewallPropertiesFormat.NetworkRuleCollections) {
		return true
	}
	//TODO: Azure firewall rules needed to added here once completed with structs

	return false
}

// NewSecurityGroupParameters returns an Azure SecurityGroup object from a Security Group Spec
func NewAzureFirewallParameters(v *v1alpha3.AzureFirewall) networkmgmt.AzureFirewall {
	return networkmgmt.AzureFirewall{
		Zones:    azure.ToStringArrayPtr(v.Spec.Zones),
		Etag:     azure.ToStringPtr(v.Spec.Etag),
		ID:       azure.ToStringPtr(v.Spec.ID),
		Name:     azure.ToStringPtr(v.Name),
		Type:     azure.ToStringPtr(v.Spec.Type),
		Location: azure.ToStringPtr(v.Spec.Location),
		Tags:     azure.ToStringPtrMap(v.Spec.Tags),
		AzureFirewallPropertiesFormat: &networkmgmt.AzureFirewallPropertiesFormat{
			//ApplicationRuleCollections: nil,
			NatRuleCollections:     setNatRulesCollections(v.Spec.NatRuleCollections),
			NetworkRuleCollections: setNetworkRulesCollections(v.Spec.NetworkRuleCollections),
			IPConfigurations:       setIPConfigurations(v.Spec.IPConfigurations),
			ProvisioningState:      networkmgmt.ProvisioningState(v.Spec.ProvisioningState),
			ThreatIntelMode:        networkmgmt.AzureFirewallThreatIntelMode(v.Spec.ThreatIntelMode),
			VirtualHub:             setSubResource(v.Spec.VirtualHub),
			FirewallPolicy:         setSubResource(v.Spec.FirewallPolicy),
			HubIPAddresses:         setHubIpAddresses(v.Spec.HubIPAddresses),
		},
	}
}

func setNetworkRulesCollections(networkRulesCollections *[]v1alpha3.AzureFirewallNetworkRuleCollection) *[]networkmgmt.AzureFirewallNetworkRuleCollection {
	if nil != networkRulesCollections {
		var afnrc = new([]networkmgmt.AzureFirewallNetworkRuleCollection)
		for _, nrc := range *networkRulesCollections {
			var networkRuleCollection = networkmgmt.AzureFirewallNetworkRuleCollection{}
			networkRuleCollection.ID = azure.ToStringPtr(nrc.ID)
			networkRuleCollection.Name = azure.ToStringPtr(nrc.Name)
			networkRuleCollection.Etag = azure.ToStringPtr(nrc.Etag)
			networkRuleCollection.AzureFirewallNetworkRuleCollectionPropertiesFormat = &networkmgmt.AzureFirewallNetworkRuleCollectionPropertiesFormat{
				Priority:          azure.ToInt32Ptr(int(nrc.Properties.Priority)),
				Action:            &networkmgmt.AzureFirewallRCAction{Type: networkmgmt.AzureFirewallRCActionType(nrc.Properties.Action)},
				Rules:             setNetworkRules(nrc.Properties.Rules),
				ProvisioningState: networkmgmt.ProvisioningState(nrc.Properties.ProvisioningState),
			}
			*afnrc = append(*afnrc, networkRuleCollection)
		}
		return afnrc
	}
	return nil
}

func setNatRulesCollections(natRuleCollections *[]v1alpha3.AzureFirewallNatRuleCollection) *[]networkmgmt.AzureFirewallNatRuleCollection {
	if nil != natRuleCollections {
		var afnrc = new([]networkmgmt.AzureFirewallNatRuleCollection)
		for _, nrc := range *natRuleCollections {
			var natRuleCollection = networkmgmt.AzureFirewallNatRuleCollection{}
			natRuleCollection.Name = azure.ToStringPtr(nrc.Name)
			natRuleCollection.ID = azure.ToStringPtr(nrc.ID)
			natRuleCollection.Etag = azure.ToStringPtr(nrc.Etag)
			natRuleCollection.AzureFirewallNatRuleCollectionProperties = &networkmgmt.AzureFirewallNatRuleCollectionProperties{
				Priority: azure.ToInt32Ptr(int(nrc.Properties.Priority)),
				Action: &networkmgmt.AzureFirewallNatRCAction{
					Type: networkmgmt.AzureFirewallNatRCActionType(nrc.Properties.Action),
				},
				Rules:             setNATRules(nrc.Properties.Rules),
				ProvisioningState: networkmgmt.ProvisioningState(nrc.Properties.ProvisioningState),
			}
			*afnrc = append(*afnrc, natRuleCollection)
		}
		return afnrc
	}
	return nil
}

func setNetworkRules(rules []v1alpha3.AzureFirewallNetworkRule) *[]networkmgmt.AzureFirewallNetworkRule {
	if nil != rules {
		var afnr = new([]networkmgmt.AzureFirewallNetworkRule)
		for _, rule := range rules {
			var r = networkmgmt.AzureFirewallNetworkRule{}
			r.Name = azure.ToStringPtr(rule.Name)
			r.Description = azure.ToStringPtr(rule.Description)
			r.Protocols = setProtocols(rule.Protocols)
			r.SourceAddresses = azure.ToStringArrayPtr(rule.SourceAddresses)
			r.DestinationAddresses = azure.ToStringArrayPtr(rule.DestinationAddresses)
			r.DestinationPorts = azure.ToStringArrayPtr(rule.DestinationPorts)
			*afnr = append(*afnr, r)
		}
		return afnr
	}
	return nil
}

func setNATRules(rules []v1alpha3.AzureFirewallNatRule) *[]networkmgmt.AzureFirewallNatRule {
	if nil != rules {
		var afnr = new([]networkmgmt.AzureFirewallNatRule)
		for _, rule := range rules {
			var r = networkmgmt.AzureFirewallNatRule{}
			r.Name = azure.ToStringPtr(rule.Name)
			r.Description = azure.ToStringPtr(rule.Description)
			r.SourceAddresses = azure.ToStringArrayPtr(rule.SourceAddresses)
			r.DestinationAddresses = azure.ToStringArrayPtr(rule.DestinationAddresses)
			r.DestinationPorts = azure.ToStringArrayPtr(rule.DestinationPorts)
			r.TranslatedAddress = azure.ToStringPtr(rule.TranslatedAddress)
			r.TranslatedPort = azure.ToStringPtr(rule.TranslatedPort)
			r.Protocols = setProtocols(rule.Protocols)
			*afnr = append(*afnr, r)
		}
		return afnr
	}
	return nil
}

func setProtocols(protocols []string) *[]networkmgmt.AzureFirewallNetworkRuleProtocol {
	if nil != protocols {
		var afnrp = new([]networkmgmt.AzureFirewallNetworkRuleProtocol)
		for _, protocol := range protocols {
			*afnrp = append(*afnrp, networkmgmt.AzureFirewallNetworkRuleProtocol(protocol))
		}
		return afnrp
	}
	return nil
}
