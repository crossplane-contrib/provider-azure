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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"

	databasev1alpha3 "github.com/crossplaneio/stack-azure/apis/database/v1alpha3"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
)

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
		r    *databasev1alpha3.PostgreSQLServerVirtualNetworkRule
		want postgresql.VirtualNetworkRule
	}{
		{
			name: "Successful",
			r: &databasev1alpha3.PostgreSQLServerVirtualNetworkRule{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: databasev1alpha3.PostgreSQLVirtualNetworkRuleSpec{
					Name:              vnetRuleName,
					ServerName:        serverName,
					ResourceGroupName: rgName,
					VirtualNetworkRuleProperties: databasev1alpha3.VirtualNetworkRuleProperties{
						VirtualNetworkSubnetID:           vnetSubnetID,
						IgnoreMissingVnetServiceEndpoint: ignoreMissing,
					},
				},
			},
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
			r: &databasev1alpha3.PostgreSQLServerVirtualNetworkRule{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: databasev1alpha3.PostgreSQLVirtualNetworkRuleSpec{
					Name:              vnetRuleName,
					ServerName:        serverName,
					ResourceGroupName: rgName,
					VirtualNetworkRuleProperties: databasev1alpha3.VirtualNetworkRuleProperties{
						VirtualNetworkSubnetID: vnetSubnetID,
					},
				},
			},
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
				t.Errorf("MySQLVirtualNetworkRuleStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestPostgreSQLServerVirtualNetworkRuleNeedsUpdate(t *testing.T) {
	cases := []struct {
		name string
		kube *databasev1alpha3.PostgreSQLServerVirtualNetworkRule
		az   postgresql.VirtualNetworkRule
		want bool
	}{
		{
			name: "NoUpdateNeeded",
			kube: &databasev1alpha3.PostgreSQLServerVirtualNetworkRule{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: databasev1alpha3.PostgreSQLVirtualNetworkRuleSpec{
					Name:              vnetRuleName,
					ServerName:        serverName,
					ResourceGroupName: rgName,
					VirtualNetworkRuleProperties: databasev1alpha3.VirtualNetworkRuleProperties{
						VirtualNetworkSubnetID:           vnetSubnetID,
						IgnoreMissingVnetServiceEndpoint: ignoreMissing,
					},
				},
			},
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
			kube: &databasev1alpha3.PostgreSQLServerVirtualNetworkRule{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: databasev1alpha3.PostgreSQLVirtualNetworkRuleSpec{
					Name:              vnetRuleName,
					ServerName:        serverName,
					ResourceGroupName: rgName,
					VirtualNetworkRuleProperties: databasev1alpha3.VirtualNetworkRuleProperties{
						VirtualNetworkSubnetID:           vnetSubnetID,
						IgnoreMissingVnetServiceEndpoint: ignoreMissing,
					},
				},
			},
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
			kube: &databasev1alpha3.PostgreSQLServerVirtualNetworkRule{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: databasev1alpha3.PostgreSQLVirtualNetworkRuleSpec{
					Name:              vnetRuleName,
					ServerName:        serverName,
					ResourceGroupName: rgName,
					VirtualNetworkRuleProperties: databasev1alpha3.VirtualNetworkRuleProperties{
						VirtualNetworkSubnetID:           vnetSubnetID,
						IgnoreMissingVnetServiceEndpoint: ignoreMissing,
					},
				},
			},
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
				t.Errorf("MySQLServerVirtualNetworkRuleNeedsUpdate(...): -want, +got\n%s", diff)
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
		want databasev1alpha3.VirtualNetworkRuleStatus
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
					State:                            postgresql.Ready,
				},
			},
			want: databasev1alpha3.VirtualNetworkRuleStatus{
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
					State:                            postgresql.Ready,
				},
			},
			want: databasev1alpha3.VirtualNetworkRuleStatus{
				State: "Ready",
				ID:    id,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := &databasev1alpha3.PostgreSQLServerVirtualNetworkRule{
				Status: databasev1alpha3.VirtualNetworkRuleStatus{
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
