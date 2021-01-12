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
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	networkmgmt "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

var (
	newTags = map[string]string{"one": "test1", "two": "test2"}
)

const (
	name      = "coolNSG"
	ruleName1 = "coolSecurityRule1"
	ruleName2 = "coolSecurityRule2"

	etagRule1         = "definitely-a-etag1"
	etagRule2         = "definitely-a-etag2"
	resourceGroupName = "coolRG"
)

func setRules() *[]v1alpha3.SecurityRule {
	var securityRules = new([]v1alpha3.SecurityRule)
	var rule1 = v1alpha3.SecurityRule{
		Properties: v1alpha3.SecurityRulePropertiesFormat{
			Description:              azure.ToStringPtr("Test Description"),
			Protocol:                 setProtocol("TCP"),
			SourcePortRange:          azure.ToStringPtr("8080"),
			DestinationPortRange:     azure.ToStringPtr("80"),
			SourceAddressPrefix:      azure.ToStringPtr("Internet"),
			DestinationAddressPrefix: azure.ToStringPtr("*"),
			Access:                   setAccess("Allow"),
			Priority:                 azure.ToInt32Ptr(120),
			Direction:                setDirection("Inbound"),
		},
		Name: ruleName1,
		Etag: "new " + etagRule1,
	}
	var rule2 = v1alpha3.SecurityRule{
		Properties: v1alpha3.SecurityRulePropertiesFormat{
			Description:              azure.ToStringPtr("Test Description"),
			Protocol:                 setProtocol("TCP"),
			SourcePortRange:          azure.ToStringPtr("8080"),
			DestinationPortRange:     azure.ToStringPtr("80"),
			SourceAddressPrefix:      azure.ToStringPtr("Internet"),
			DestinationAddressPrefix: azure.ToStringPtr("*"),
			Access:                   setAccess("Deny"),
			Priority:                 azure.ToInt32Ptr(130),
			Direction:                setDirection("Outbound"),
		},
		Name: ruleName2,
		Etag: "new " + etagRule2,
	}
	*securityRules = append(*securityRules, rule1)
	*securityRules = append(*securityRules, rule2)
	return securityRules
}

func setDirection(s string) *v1alpha3.SecurityRuleDirection {
	var direction = v1alpha3.SecurityRuleDirection(s)
	return &direction
}

func setAccess(s string) *v1alpha3.SecurityRuleAccess {
	var access = v1alpha3.SecurityRuleAccess(s)
	return &access
}

func setProtocol(s string) *v1alpha3.SecurityRuleProtocol {
	var protocol = v1alpha3.SecurityRuleProtocol(s)
	return &protocol
}

func setUpdatedRules() *[]v1alpha3.SecurityRule {
	var securityRules = new([]v1alpha3.SecurityRule)
	var rule1 = v1alpha3.SecurityRule{
		Properties: v1alpha3.SecurityRulePropertiesFormat{
			Description:              azure.ToStringPtr("Test Description"),
			Protocol:                 setProtocol("TCP"),
			SourcePortRange:          azure.ToStringPtr("8080"),
			DestinationPortRange:     azure.ToStringPtr("80"),
			SourceAddressPrefix:      azure.ToStringPtr("Internet"),
			DestinationAddressPrefix: azure.ToStringPtr("*"),
			Access:                   setAccess("Allow"),
			Priority:                 azure.ToInt32Ptr(120),
			Direction:                setDirection("Inbound"),
		},
		Name: ruleName1,
		Etag: "new " + etagRule1,
	}
	*securityRules = append(*securityRules, rule1)
	return securityRules
}

func setNewRules() *[]v1alpha3.SecurityRule {
	var securityRules = new([]v1alpha3.SecurityRule)
	var rule1 = v1alpha3.SecurityRule{
		Properties: v1alpha3.SecurityRulePropertiesFormat{
			Description:                          azure.ToStringPtr("Test Description"),
			Protocol:                             setProtocol("TCP"),
			SourcePortRange:                      azure.ToStringPtr("8080"),
			DestinationPortRange:                 azure.ToStringPtr("80"),
			SourceAddressPrefix:                  azure.ToStringPtr("Internet"),
			SourceAddressPrefixes:                nil,
			SourceApplicationSecurityGroups:      setASGs(),
			DestinationAddressPrefix:             azure.ToStringPtr("*"),
			DestinationAddressPrefixes:           nil,
			DestinationApplicationSecurityGroups: setASGs(),
			SourcePortRanges:                     nil,
			DestinationPortRanges:                nil,
			Access:                               setAccess("Allow"),
			Priority:                             azure.ToInt32Ptr(120),
			Direction:                            setDirection("Inbound"),
			ProvisioningState:                    azure.ToStringPtr(""),
		},
		Name: ruleName1,
		Etag: etagRule1,
	}
	var rule2 = v1alpha3.SecurityRule{
		Properties: v1alpha3.SecurityRulePropertiesFormat{
			Description:                          azure.ToStringPtr("Test Description"),
			Protocol:                             setProtocol("TCP"),
			SourcePortRange:                      azure.ToStringPtr("8080"),
			DestinationPortRange:                 azure.ToStringPtr("80"),
			SourceAddressPrefix:                  azure.ToStringPtr("Internet"),
			SourceAddressPrefixes:                nil,
			SourceApplicationSecurityGroups:      setASGs(),
			DestinationAddressPrefix:             azure.ToStringPtr("*"),
			DestinationAddressPrefixes:           nil,
			DestinationApplicationSecurityGroups: setASGs(),
			SourcePortRanges:                     nil,
			DestinationPortRanges:                nil,
			Access:                               setAccess("Deny"),
			Priority:                             azure.ToInt32Ptr(130),
			Direction:                            setDirection("Outbound"),
			ProvisioningState:                    azure.ToStringPtr(""),
		},
		Name: ruleName2,
		Etag: etagRule2,
	}
	*securityRules = append(*securityRules, rule1)
	*securityRules = append(*securityRules, rule2)
	return securityRules
}

func setSecurityRules() *[]network.SecurityRule {
	var securityRules = new([]network.SecurityRule)
	var rule1 = network.SecurityRule{
		SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
			Description:                          azure.ToStringPtr("Test Description"),
			Protocol:                             network.SecurityRuleProtocol("TCP"),
			SourcePortRange:                      azure.ToStringPtr("8080"),
			DestinationPortRange:                 azure.ToStringPtr("80"),
			SourceAddressPrefix:                  azure.ToStringPtr("Internet"),
			SourceAddressPrefixes:                azure.ToStringArrayPtr(nil),
			SourceApplicationSecurityGroups:      setNetworkASGs(),
			DestinationAddressPrefix:             azure.ToStringPtr("*"),
			DestinationAddressPrefixes:           azure.ToStringArrayPtr(nil),
			DestinationApplicationSecurityGroups: setNetworkASGs(),
			SourcePortRanges:                     azure.ToStringArrayPtr(nil),
			DestinationPortRanges:                azure.ToStringArrayPtr(nil),
			Access:                               network.SecurityRuleAccess("Allow"),
			Priority:                             azure.ToInt32Ptr(120),
			Direction:                            network.SecurityRuleDirection("Inbound"),
			ProvisioningState:                    azure.ToStringPtr(""),
		},
		Name: azure.ToStringPtr(ruleName1),
		Etag: azure.ToStringPtr(etagRule1),
	}
	var rule2 = network.SecurityRule{
		SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
			Description:                          azure.ToStringPtr("Test Description"),
			Protocol:                             network.SecurityRuleProtocol("TCP"),
			SourcePortRange:                      azure.ToStringPtr("8080"),
			DestinationPortRange:                 azure.ToStringPtr("80"),
			SourceAddressPrefix:                  azure.ToStringPtr("Internet"),
			SourceAddressPrefixes:                azure.ToStringArrayPtr(nil),
			SourceApplicationSecurityGroups:      setNetworkASGs(),
			DestinationAddressPrefix:             azure.ToStringPtr("*"),
			DestinationAddressPrefixes:           azure.ToStringArrayPtr(nil),
			DestinationApplicationSecurityGroups: setNetworkASGs(),
			SourcePortRanges:                     azure.ToStringArrayPtr(nil),
			DestinationPortRanges:                azure.ToStringArrayPtr(nil),
			Access:                               network.SecurityRuleAccess("Deny"),
			Priority:                             azure.ToInt32Ptr(130),
			Direction:                            network.SecurityRuleDirection("Outbound"),
			ProvisioningState:                    azure.ToStringPtr(""),
		},
		Name: azure.ToStringPtr(ruleName2),
		Etag: azure.ToStringPtr(etagRule2),
	}
	*securityRules = append(*securityRules, rule1)
	*securityRules = append(*securityRules, rule2)
	return securityRules
}
func setASGs() *[]v1alpha3.ApplicationSecurityGroup {
	asgs := new([]v1alpha3.ApplicationSecurityGroup)
	applicationSecurityGroup := v1alpha3.ApplicationSecurityGroup{
		Properties: v1alpha3.ApplicationSecurityGroupPropertiesFormat{},
		Etag:       "",
		ID:         "",
		Name:       "cool-asg",
		Type:       "",
		Location:   "",
	}
	*asgs = append(*asgs, applicationSecurityGroup)
	return asgs
}

func setNetworkASGs() *[]networkmgmt.ApplicationSecurityGroup {
	asgs := new([]networkmgmt.ApplicationSecurityGroup)
	applicationSecurityGroup := networkmgmt.ApplicationSecurityGroup{
		ApplicationSecurityGroupPropertiesFormat: &networkmgmt.ApplicationSecurityGroupPropertiesFormat{},
		Etag:                                     azure.ToStringPtr(""),
		ID:                                       azure.ToStringPtr(""),
		Name:                                     azure.ToStringPtr("cool-asg"),
		Type:                                     azure.ToStringPtr(""),
		Location:                                 azure.ToStringPtr(""),
		Tags:                                     azure.ToStringPtrMap(nil),
	}
	*asgs = append(*asgs, applicationSecurityGroup)
	return asgs
}

func TestUpdateSecurityGroupStatusFromAzure(t *testing.T) {
	mockCondition := runtimev1alpha1.Condition{Message: "mockMessage"}
	resourceStatus := runtimev1alpha1.ResourceStatus{
		ConditionedStatus: runtimev1alpha1.ConditionedStatus{
			Conditions: []runtimev1alpha1.Condition{mockCondition},
		},
	}

	cases := []struct {
		name string
		sg   networkmgmt.SecurityGroup
		want v1alpha3.SecurityGroupStatus
	}{
		{
			name: "SuccessfulFull",
			sg: networkmgmt.SecurityGroup{
				Location: azure.ToStringPtr(location),
				Etag:     azure.ToStringPtr(etag),
				ID:       azure.ToStringPtr(id),
				Type:     azure.ToStringPtr(resourceType),
				Tags:     azure.ToStringPtrMap(nil),
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{
					SecurityRules:     setSecurityRules(),
					ResourceGUID:      azure.ToStringPtr(string(uid)),
					ProvisioningState: azure.ToStringPtr("Succeeded"),
				},
			},
			want: v1alpha3.SecurityGroupStatus{
				State:        string(networkmgmt.Succeeded),
				ID:           id,
				Etag:         etag,
				Type:         resourceType,
				ResourceGUID: string(uid),
			},
		},
		{
			name: "SuccessfulPartial",
			sg: networkmgmt.SecurityGroup{
				Location: azure.ToStringPtr(location),
				Type:     azure.ToStringPtr(resourceType),
				Tags:     azure.ToStringPtrMap(nil),
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{
					SecurityRules:     setSecurityRules(),
					ResourceGUID:      azure.ToStringPtr(string(uid)),
					ProvisioningState: azure.ToStringPtr("Succeeded"),
				},
			},
			want: v1alpha3.SecurityGroupStatus{
				State:        string(networkmgmt.Succeeded),
				ResourceGUID: string(uid),
				Type:         resourceType,
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			sg := &v1alpha3.SecurityGroup{
				Status: v1alpha3.SecurityGroupStatus{
					ResourceStatus: resourceStatus,
				},
			}

			UpdateSecurityGroupStatusFromAzure(sg, tc.sg)

			// make sure that internal resource status hasn't changed
			if diff := cmp.Diff(mockCondition, sg.Status.ResourceStatus.Conditions[0]); diff != "" {
				t.Errorf("UpdateSecurityGroupStatusFromAzure(...): -want, +got\n%s", diff)
			}

			// make sure that other resource parameters are updated
			tc.want.ResourceStatus = resourceStatus
			if diff := cmp.Diff(tc.want, sg.Status); diff != "" {
				t.Errorf("UpdateSecurityGroupStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestSecurityGroupNeedsUpdate(t *testing.T) {
	cases := []struct {
		name string
		kube *v1alpha3.SecurityGroup
		az   networkmgmt.SecurityGroup
		want bool
	}{
		{
			name: "NeedsUpdateName",
			kube: &v1alpha3.SecurityGroup{
				Spec: v1alpha3.SecurityGroupSpec{
					ForProvider: v1alpha3.SecurityGroupParameters{
						ResourceGroupName:             resourceGroupName,
						Location:                      location,
						SecurityGroupPropertiesFormat: v1alpha3.SecurityGroupPropertiesFormat{},
						Tags:                          tags,
					},
				},
			},
			az: networkmgmt.SecurityGroup{
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{},
				Name:                          azure.ToStringPtr("new name"),
				Location:                      azure.ToStringPtr(location),
				Tags:                          azure.ToStringPtrMap(tags),
			},
			want: true,
		},
		{
			name: "NeedsUpdateLocation",
			kube: &v1alpha3.SecurityGroup{
				Spec: v1alpha3.SecurityGroupSpec{
					ForProvider: v1alpha3.SecurityGroupParameters{
						ResourceGroupName:             resourceGroupName,
						Location:                      location,
						SecurityGroupPropertiesFormat: v1alpha3.SecurityGroupPropertiesFormat{},
						Tags:                          tags,
					},
				},
			},
			az: networkmgmt.SecurityGroup{
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{},
				Location:                      azure.ToStringPtr("new location"),
				Name:                          azure.ToStringPtr(name),
				Tags:                          azure.ToStringPtrMap(tags),
			},
			want: true,
		},
		{
			name: "NeedsUpdateTags",
			kube: &v1alpha3.SecurityGroup{
				Spec: v1alpha3.SecurityGroupSpec{
					ForProvider: v1alpha3.SecurityGroupParameters{
						ResourceGroupName:             resourceGroupName,
						Location:                      location,
						SecurityGroupPropertiesFormat: v1alpha3.SecurityGroupPropertiesFormat{},
						Tags:                          tags,
					},
				},
			},
			az: networkmgmt.SecurityGroup{
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{},
				Location:                      azure.ToStringPtr(location),
				Name:                          azure.ToStringPtr(name),
				Tags:                          azure.ToStringPtrMap(newTags),
			},
			want: true,
		},
		{
			name: "NeedsUpdateSecurityRules",
			kube: &v1alpha3.SecurityGroup{
				Spec: v1alpha3.SecurityGroupSpec{
					ForProvider: v1alpha3.SecurityGroupParameters{
						ResourceGroupName: resourceGroupName,
						Location:          location,
						SecurityGroupPropertiesFormat: v1alpha3.SecurityGroupPropertiesFormat{
							SecurityRules: setRules(),
						},
						Tags: tags,
					},
				},
			},
			az: networkmgmt.SecurityGroup{
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{
					SecurityRules: setSecurityRules(),
				},
				Location: azure.ToStringPtr(location),
				Name:     azure.ToStringPtr(name),
				Tags:     azure.ToStringPtrMap(tags),
			},
			want: true,
		},
		{
			name: "NeedsUpdateSecurityRulesRuleDeletedOnAzure",
			kube: &v1alpha3.SecurityGroup{
				Spec: v1alpha3.SecurityGroupSpec{
					ForProvider: v1alpha3.SecurityGroupParameters{
						ResourceGroupName: resourceGroupName,
						Location:          location,
						SecurityGroupPropertiesFormat: v1alpha3.SecurityGroupPropertiesFormat{
							SecurityRules: setRules(),
						},
						Tags: tags,
					},
				},
			},
			az: networkmgmt.SecurityGroup{
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{
					SecurityRules: nil,
				},
				Location: azure.ToStringPtr(location),
				Name:     azure.ToStringPtr(name),
				Tags:     azure.ToStringPtrMap(tags),
			},
			want: true,
		}, {
			name: "NeedsUpdateSecurityRulesRuleDeletedOnCluster",
			kube: &v1alpha3.SecurityGroup{
				Spec: v1alpha3.SecurityGroupSpec{
					ForProvider: v1alpha3.SecurityGroupParameters{
						ResourceGroupName: resourceGroupName,
						Location:          location,
						SecurityGroupPropertiesFormat: v1alpha3.SecurityGroupPropertiesFormat{
							SecurityRules: nil,
						},
						Tags: tags,
					},
				},
			},
			az: networkmgmt.SecurityGroup{
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{
					SecurityRules: setSecurityRules(),
				},
				Location: azure.ToStringPtr(location),
				Name:     azure.ToStringPtr(name),
				Tags:     azure.ToStringPtrMap(tags),
			},
			want: true,
		}, {
			name: "NeedsUpdateSecurityRulesRuleCountDontMatch",
			kube: &v1alpha3.SecurityGroup{
				Spec: v1alpha3.SecurityGroupSpec{
					ForProvider: v1alpha3.SecurityGroupParameters{
						ResourceGroupName: resourceGroupName,
						Location:          location,
						SecurityGroupPropertiesFormat: v1alpha3.SecurityGroupPropertiesFormat{
							SecurityRules: setUpdatedRules(),
						},
						Tags: tags,
					},
				},
			},
			az: networkmgmt.SecurityGroup{
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{
					SecurityRules: setSecurityRules(),
				},
				Location: azure.ToStringPtr(location),
				Name:     azure.ToStringPtr(name),
				Tags:     azure.ToStringPtrMap(tags),
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.kube.ObjectMeta.Name = name
			got := SecurityGroupNeedsUpdate(tc.kube, tc.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("SecurityGroupNeedsUpdate(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestNewSecurityGroupParameters(t *testing.T) {
	cases := []struct {
		name string
		sg   *v1alpha3.SecurityGroup
		want networkmgmt.SecurityGroup
	}{
		{
			name: "SuccessfulFull",
			sg: &v1alpha3.SecurityGroup{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.SecurityGroupSpec{
					ForProvider: v1alpha3.SecurityGroupParameters{
						Location: location,
						SecurityGroupPropertiesFormat: v1alpha3.SecurityGroupPropertiesFormat{
							SecurityRules:        setNewRules(),
							DefaultSecurityRules: nil,
							ResourceGUID:         nil,
							ProvisioningState:    nil,
						},
						Tags: tags,
					},
				},
			},
			want: networkmgmt.SecurityGroup{
				Location: azure.ToStringPtr(location),
				Tags:     azure.ToStringPtrMap(tags),
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{
					SecurityRules:        setSecurityRules(),
					DefaultSecurityRules: nil,
					NetworkInterfaces:    nil,
					Subnets:              nil,
					ResourceGUID:         nil,
					ProvisioningState:    nil,
				},
			},
		},
		{
			name: "SuccessfulPartial",
			sg: &v1alpha3.SecurityGroup{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.SecurityGroupSpec{
					ForProvider: v1alpha3.SecurityGroupParameters{
						Location:                      location,
						SecurityGroupPropertiesFormat: v1alpha3.SecurityGroupPropertiesFormat{},
						Tags:                          tags,
					},
				},
			},
			want: networkmgmt.SecurityGroup{
				Location:                      azure.ToStringPtr(location),
				Tags:                          azure.ToStringPtrMap(tags),
				SecurityGroupPropertiesFormat: &networkmgmt.SecurityGroupPropertiesFormat{},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewSecurityGroupParameters(tc.sg)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewSecurityGroupParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestSetSecurityRulesToSecurityGroup(t *testing.T) {
	cases := []struct {
		name string
		sg   *[]v1alpha3.SecurityRule
		want *[]networkmgmt.SecurityRule
	}{
		{
			name: "SuccessfulFull",
			sg:   setNewRules(),
			want: setSecurityRules(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := SetSecurityRulesToSecurityGroup(tc.sg)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("SetSecurityRulesToSecurityGroup(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestSetApplicationSecurityGroups(t *testing.T) {
	cases := []struct {
		name string
		sg   *[]v1alpha3.ApplicationSecurityGroup
		want *[]networkmgmt.ApplicationSecurityGroup
	}{
		{
			name: "SuccessfulFull",
			sg:   setTestASGs(),
			want: getTestASGs(),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := setApplicationSecurityGroups(tc.sg)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("SetApplicationSecurityGroups(...): -want, +got\n%s", diff)
			}
		})
	}

}

func getTestASGs() *[]networkmgmt.ApplicationSecurityGroup {
	asgs := new([]networkmgmt.ApplicationSecurityGroup)
	applicationSecurityGroup := networkmgmt.ApplicationSecurityGroup{
		ApplicationSecurityGroupPropertiesFormat: &networkmgmt.ApplicationSecurityGroupPropertiesFormat{},
		Etag:                                     azure.ToStringPtr(etag),
		ID:                                       azure.ToStringPtr("new-id"),
		Name:                                     azure.ToStringPtr("cool-asg"),
		Type:                                     azure.ToStringPtr("new-type"),
		Location:                                 azure.ToStringPtr(location),
		Tags:                                     nil,
	}
	*asgs = append(*asgs, applicationSecurityGroup)
	return asgs
}

func setTestASGs() *[]v1alpha3.ApplicationSecurityGroup {
	asgs := new([]v1alpha3.ApplicationSecurityGroup)
	applicationSecurityGroup := v1alpha3.ApplicationSecurityGroup{
		Properties: v1alpha3.ApplicationSecurityGroupPropertiesFormat{},
		Etag:       etag,
		ID:         "new-id",
		Name:       "cool-asg",
		Type:       "new-type",
		Location:   location,
	}
	*asgs = append(*asgs, applicationSecurityGroup)
	return asgs
}
