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

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	"github.com/crossplane/provider-azure/apis/database/v1beta1"
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

func postgresqlServerParameters(createMode *v1beta1.CreateMode) v1beta1.SQLServerParameters {
	fp := v1beta1.SQLServerParameters{
		CreateMode: createMode,
	}
	return fp
}

func postgresqlServerPropertiesForDefaultCreate() postgresql.BasicServerPropertiesForCreate {
	adminPassword := "admin"
	return &postgresql.ServerPropertiesForDefaultCreate{
		AdministratorLoginPassword: &adminPassword,
		CreateMode:                 postgresql.CreateModeDefault,
		StorageProfile:             &postgresql.StorageProfile{},
	}
}

func postgresqlServerPropertiesForRestore() postgresql.BasicServerPropertiesForCreate {
	return &postgresql.ServerPropertiesForRestore{
		CreateMode:     postgresql.CreateModePointInTimeRestore,
		StorageProfile: &postgresql.StorageProfile{},
	}
}

func postgresqlServerPropertiesForGeoRestore() postgresql.BasicServerPropertiesForCreate {
	return &postgresql.ServerPropertiesForGeoRestore{
		CreateMode:     postgresql.CreateModeGeoRestore,
		StorageProfile: &postgresql.StorageProfile{},
	}
}

func postgresqlServerPropertiesForReplica() postgresql.BasicServerPropertiesForCreate {
	return &postgresql.ServerPropertiesForReplica{
		CreateMode:     postgresql.CreateModeReplica,
		StorageProfile: &postgresql.StorageProfile{},
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

func TestTopostgresqlProperties(t *testing.T) {
	cases := []struct {
		name string
		fp   v1beta1.SQLServerParameters
		want postgresql.BasicServerPropertiesForCreate
	}{
		{
			name: "CreateModeDefault",
			fp:   postgresqlServerParameters(pointerFromCreateMode(v1beta1.CreateModeDefault)),
			want: postgresqlServerPropertiesForDefaultCreate(),
		},
		{
			name: "CreateModePointInTimeRestore",
			fp:   postgresqlServerParameters(pointerFromCreateMode(v1beta1.CreateModePointInTimeRestore)),
			want: postgresqlServerPropertiesForRestore(),
		},
		{
			name: "CreateModeGeoRestore",
			fp:   postgresqlServerParameters(pointerFromCreateMode(v1beta1.CreateModeGeoRestore)),
			want: postgresqlServerPropertiesForGeoRestore(),
		},
		{
			name: "CreateModeReplica",
			fp:   postgresqlServerParameters(pointerFromCreateMode(v1beta1.CreateModeReplica)),
			want: postgresqlServerPropertiesForReplica(),
		},
		{
			name: "ServerPropertiesForInvalidString",
			fp:   postgresqlServerParameters(pointerFromCreateMode("")),
			want: postgresqlServerPropertiesForDefaultCreate(),
		},
		{
			name: "ServerPropertiesForDefaultCreate",
			fp:   postgresqlServerParameters(nil),
			want: postgresqlServerPropertiesForDefaultCreate(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := toPGSQLProperties(tc.fp, "admin")
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("TestTopostgresqlProperties(%s): -want, +got\n%s", tc.name, diff)
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
	mockCondition := xpv1.Condition{Message: "mockMessage"}
	resourceStatus := xpv1.ResourceStatus{
		ConditionedStatus: xpv1.ConditionedStatus{
			Conditions: []xpv1.Condition{mockCondition},
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

func TestIsPostgreSQLUpToDate(t *testing.T) {
	type args struct {
		p  v1beta1.SQLServerParameters
		in postgresql.Server
	}
	cases := map[string]struct {
		args
		want bool
	}{
		"IsUpToDateWithAllDefault": {
			args: args{
				p: v1beta1.SQLServerParameters{},
				in: postgresql.Server{
					Sku: &postgresql.Sku{},
					ServerProperties: &postgresql.ServerProperties{
						StorageProfile: &postgresql.StorageProfile{},
					},
				},
			},
			want: true,
		},
		"IsUpToDate": {
			args: args{
				p: v1beta1.SQLServerParameters{
					MinimalTLSVersion: "TLS1_2",
					SSLEnforcement:    "Enabled",
					Version:           "9.6",
					Tags: map[string]string{
						"created_by": "crossplane",
					},
					SKU: v1beta1.SKU{
						Tier:     "GeneralPurpose",
						Capacity: 2,
						Family:   "Gen5",
					},
					PublicNetworkAccess: azure.ToStringPtr("Enabled"),
					StorageProfile: v1beta1.StorageProfile{
						StorageMB:           20480,
						StorageAutogrow:     azure.ToStringPtr("Enabled"),
						BackupRetentionDays: to.IntPtr(5),
						GeoRedundantBackup:  azure.ToStringPtr("Disabled"),
					},
				},
				in: postgresql.Server{
					Tags: map[string]*string{
						"created_by": azure.ToStringPtr("crossplane"),
					},
					Sku: &postgresql.Sku{
						Tier:     postgresql.GeneralPurpose,
						Capacity: azure.ToInt32Ptr(2),
						Family:   azure.ToStringPtr("Gen5"),
					},
					ServerProperties: &postgresql.ServerProperties{
						Version: "9.6",
						StorageProfile: &postgresql.StorageProfile{
							StorageMB:           azure.ToInt32Ptr(20480),
							StorageAutogrow:     postgresql.StorageAutogrowEnabled,
							BackupRetentionDays: azure.ToInt32Ptr(5),
							GeoRedundantBackup:  postgresql.Disabled,
						},
						SslEnforcement:      postgresql.SslEnforcementEnumEnabled,
						MinimalTLSVersion:   postgresql.TLS12,
						PublicNetworkAccess: postgresql.PublicNetworkAccessEnumEnabled,
					},
				},
			},
			want: true,
		},
		"IsNotUpToDate": {
			args: args{
				p: v1beta1.SQLServerParameters{
					PublicNetworkAccess: azure.ToStringPtr("Disabled"),
				},
				in: postgresql.Server{
					Sku: &postgresql.Sku{},
					ServerProperties: &postgresql.ServerProperties{
						StorageProfile:      &postgresql.StorageProfile{},
						PublicNetworkAccess: postgresql.PublicNetworkAccessEnumEnabled,
					},
				},
			},
			want: false,
		},
		"IsNotUpToDateWithServerWithoutSku": {
			args: args{
				p: v1beta1.SQLServerParameters{},
				in: postgresql.Server{
					ServerProperties: &postgresql.ServerProperties{
						StorageProfile: &postgresql.StorageProfile{},
					},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsPostgreSQLUpToDate(tc.args.p, tc.args.in)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("IsPostgreSQLUpToDate(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestLateInitializePostgreSQL(t *testing.T) {
	type args struct {
		p  *v1beta1.SQLServerParameters
		in postgresql.Server
	}
	cases := map[string]struct {
		args
		want *v1beta1.SQLServerParameters
	}{
		"PublicNetworkAccessLateInitialize": {
			args: args{
				p: &v1beta1.SQLServerParameters{},
				in: postgresql.Server{
					Sku: &postgresql.Sku{},
					ServerProperties: &postgresql.ServerProperties{
						PublicNetworkAccess: postgresql.PublicNetworkAccessEnumEnabled,
					},
				},
			},
			want: &v1beta1.SQLServerParameters{
				PublicNetworkAccess: azure.ToStringPtr("Enabled"),
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializePostgreSQL(tc.args.p, tc.args.in)
			if diff := cmp.Diff(tc.want, tc.args.p); diff != "" {
				t.Errorf("LateInitializePostgreSQL(...): -want, +got\n%s", diff)
			}
		})
	}
}
