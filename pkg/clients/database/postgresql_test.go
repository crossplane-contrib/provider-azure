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

package database

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

type postgreSQLVirtualNetworkRuleModifier func(*v1alpha3.PostgreSQLServerVirtualNetworkRule)

func postgreSQLWithSubnetID(id string) postgreSQLVirtualNetworkRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerVirtualNetworkRule) {
		r.Spec.VirtualNetworkSubnetID = id
	}
}

func postgreSQLWithIgnoreMissing(ignore bool) postgreSQLVirtualNetworkRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerVirtualNetworkRule) {
		r.Spec.IgnoreMissingVnetServiceEndpoint = ignore
	}
}

func postgreSQLVirtualNetworkRule(sm ...postgreSQLVirtualNetworkRuleModifier) *v1alpha3.PostgreSQLServerVirtualNetworkRule {
	r := &v1alpha3.PostgreSQLServerVirtualNetworkRule{
		Spec: v1alpha3.PostgreSQLVirtualNetworkRuleSpec{
			ServerName:        serverName,
			ResourceGroupName: rgName,
		},
	}

	meta.SetExternalName(r, vnetRuleName)

	for _, m := range sm {
		m(r)
	}

	return r
}

func TestNewPostgreSQLVirtualNetworkRulesClient(t *testing.T) {
	cases := []struct {
		name       string
		r          []byte
		returnsErr bool
	}{
		{
			name: "Successful",
			r:    []byte(credentials),
		},
		{
			name:       "Unsuccessful",
			r:          []byte("invalid"),
			returnsErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewPostgreSQLVirtualNetworkRulesClient(ctx, tc.r)

			if tc.returnsErr != (err != nil) {
				t.Errorf("NewPostgreSQLVirtualNetworkRulesClient(...) error: want: %t got: %t", tc.returnsErr, err != nil)
			}

			if _, ok := got.(PostgreSQLVirtualNetworkRulesClient); !ok && !tc.returnsErr {
				t.Error("NewPostgreSQLVirtualNetworkRulesClient(...): got does not satisfy PostgreSQLVirtualNetworkRulesClient interface")
			}
		})
	}
}

func TestNewPostgreSQLVirtualNetworkRuleParameters(t *testing.T) {
	cases := []struct {
		name string
		r    *v1alpha3.PostgreSQLServerVirtualNetworkRule
		want postgresql.VirtualNetworkRule
	}{
		{
			name: "Successful",
			r: postgreSQLVirtualNetworkRule(
				postgreSQLWithSubnetID(vnetSubnetID),
				postgreSQLWithIgnoreMissing(ignoreMissing),
			),
			want: postgresql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           to.StringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: to.BoolPtr(ignoreMissing),
				},
			},
		},
		{
			name: "SuccessfulPartial",
			r: postgreSQLVirtualNetworkRule(
				postgreSQLWithSubnetID(vnetSubnetID),
			),
			want: postgresql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           to.StringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: to.BoolPtr(false),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewPostgreSQLVirtualNetworkRuleParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("PostgreSQLVirtualNetworkRuleStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestPostgreSQLServerVirtualNetworkRuleNeedsUpdate(t *testing.T) {
	cases := []struct {
		name string
		kube *v1alpha3.PostgreSQLServerVirtualNetworkRule
		az   postgresql.VirtualNetworkRule
		want bool
	}{
		{
			name: "NoUpdateNeeded",
			kube: postgreSQLVirtualNetworkRule(
				postgreSQLWithSubnetID(vnetSubnetID),
				postgreSQLWithIgnoreMissing(ignoreMissing),
			),
			az: postgresql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
				},
			},
			want: false,
		},
		{
			name: "UpdateNeededVirtualNetworkSubnetID",
			kube: postgreSQLVirtualNetworkRule(
				postgreSQLWithSubnetID(vnetSubnetID),
				postgreSQLWithIgnoreMissing(ignoreMissing),
			),
			az: postgresql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr("some/other/subnet"),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
				},
			},
			want: true,
		},
		{
			name: "UpdateNeededIgnoreMissingVnetServiceEndpoint",
			kube: postgreSQLVirtualNetworkRule(
				postgreSQLWithSubnetID(vnetSubnetID),
				postgreSQLWithIgnoreMissing(ignoreMissing),
			),
			az: postgresql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(!ignoreMissing),
				},
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := PostgreSQLServerVirtualNetworkRuleNeedsUpdate(tc.kube, tc.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("PostgreSQLServerVirtualNetworkRuleNeedsUpdate(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestUpdatePostgreSQLVirtualNetworkRuleStatusFromAzure(t *testing.T) {
	mockCondition := runtimev1alpha1.Condition{Message: "mockMessage"}
	resourceStatus := runtimev1alpha1.ResourceStatus{
		ConditionedStatus: runtimev1alpha1.ConditionedStatus{
			Conditions: []runtimev1alpha1.Condition{mockCondition},
		},
	}

	cases := []struct {
		name string
		r    postgresql.VirtualNetworkRule
		want v1alpha3.VirtualNetworkRuleStatus
	}{
		{
			name: "SuccessfulFull",
			r: postgresql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				ID:   azure.ToStringPtr(id),
				Type: azure.ToStringPtr(resourceType),
				VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
					State:                            postgresql.VirtualNetworkRuleStateReady,
				},
			},
			want: v1alpha3.VirtualNetworkRuleStatus{
				State: "Ready",
				ID:    id,
				Type:  resourceType,
			},
		},
		{
			name: "SuccessfulPartial",
			r: postgresql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				ID:   azure.ToStringPtr(id),
				VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
					State:                            postgresql.VirtualNetworkRuleStateReady,
				},
			},
			want: v1alpha3.VirtualNetworkRuleStatus{
				State: "Ready",
				ID:    id,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := &v1alpha3.PostgreSQLServerVirtualNetworkRule{
				Status: v1alpha3.VirtualNetworkRuleStatus{
					ResourceStatus: resourceStatus,
				},
			}

			UpdatePostgreSQLVirtualNetworkRuleStatusFromAzure(v, tc.r)

			// make sure that internal resource status hasn't changed
			if diff := cmp.Diff(mockCondition, v.Status.ResourceStatus.Conditions[0]); diff != "" {
				t.Errorf("UpdatePostgreSQLVirtualNetworkRuleStatusFromAzure(...): -want, +got\n%s", diff)
			}

			// make sure that other resource parameters are updated
			tc.want.ResourceStatus = resourceStatus
			if diff := cmp.Diff(tc.want, v.Status); diff != "" {
				t.Errorf("UpdatePostgreSQLVirtualNetworkRuleStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestNewPostgreSQLFirewallRuleParameters(t *testing.T) {
	name := "coolrule"
	start := "127.0.0.1."
	end := "It was just a dream Bender - there's no such thing as two."

	cases := map[string]struct {
		r    *v1alpha3.PostgreSQLServerFirewallRule
		want postgresql.FirewallRule
	}{
		"Successful": {
			r: func() *v1alpha3.PostgreSQLServerFirewallRule {
				r := &v1alpha3.PostgreSQLServerFirewallRule{
					Spec: v1alpha3.FirewallRuleSpec{
						ForProvider: v1alpha3.FirewallRuleParameters{
							FirewallRuleProperties: v1alpha3.FirewallRuleProperties{
								StartIPAddress: start,
								EndIPAddress:   end,
							},
						},
					},
				}
				meta.SetExternalName(r, name)
				return r
			}(),
			want: postgresql.FirewallRule{
				Name: azure.ToStringPtr(name),
				FirewallRuleProperties: &postgresql.FirewallRuleProperties{
					StartIPAddress: azure.ToStringPtr(start),
					EndIPAddress:   azure.ToStringPtr(end),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := NewPostgreSQLFirewallRuleParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewPostgreSQLFirewallRuleParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestPostgreSQLServerFirewallRuleIsUpToDate(t *testing.T) {
	start := "127.0.0.1."
	end := "256"

	cases := map[string]struct {
		kube *v1alpha3.PostgreSQLServerFirewallRule
		az   postgresql.FirewallRule
		want bool
	}{
		"UpToDate": {
			kube: &v1alpha3.PostgreSQLServerFirewallRule{},
			az: postgresql.FirewallRule{
				Name:                   azure.ToStringPtr(vnetRuleName),
				FirewallRuleProperties: &postgresql.FirewallRuleProperties{},
			},
			want: true,
		},
		"StartNeedsUpdate": {
			kube: &v1alpha3.PostgreSQLServerFirewallRule{
				Spec: v1alpha3.FirewallRuleSpec{ForProvider: v1alpha3.FirewallRuleParameters{FirewallRuleProperties: v1alpha3.FirewallRuleProperties{
					StartIPAddress: start,
					EndIPAddress:   end,
				}}},
			},
			az: postgresql.FirewallRule{
				FirewallRuleProperties: &postgresql.FirewallRuleProperties{
					StartIPAddress: azure.ToStringPtr("255.255.255.254"),
					EndIPAddress:   azure.ToStringPtr(end),
				},
			},
			want: false,
		},
		"EndNeedsUpdate": {
			kube: &v1alpha3.PostgreSQLServerFirewallRule{
				Spec: v1alpha3.FirewallRuleSpec{ForProvider: v1alpha3.FirewallRuleParameters{FirewallRuleProperties: v1alpha3.FirewallRuleProperties{
					StartIPAddress: start,
					EndIPAddress:   end,
				}}},
			},
			az: postgresql.FirewallRule{
				FirewallRuleProperties: &postgresql.FirewallRuleProperties{
					StartIPAddress: azure.ToStringPtr(start),
					EndIPAddress:   azure.ToStringPtr("192.168.0.1"),
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := PostgreSQLServerFirewallRuleIsUpToDate(tc.kube, tc.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("PostgreSQLServerFirewallRuleIsUpToDate(...): -want, +got\n%s", diff)
			}
		})
	}
}
