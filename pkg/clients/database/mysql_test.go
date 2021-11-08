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

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	"github.com/crossplane/provider-azure/apis/database/v1beta1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

const (
	vnetRuleName  = "myvnetrule"
	serverName    = "myserver"
	rgName        = "myrg"
	vnetSubnetID  = "a/very/important/subnet"
	ignoreMissing = true

	id           = "very-cool-id"
	resourceType = "very-cool-type"
)

type mySQLVirtualNetworkRuleModifier func(*v1alpha3.MySQLServerVirtualNetworkRule)

func mySQLWithSubnetID(id string) mySQLVirtualNetworkRuleModifier {
	return func(r *v1alpha3.MySQLServerVirtualNetworkRule) {
		r.Spec.VirtualNetworkSubnetID = id
	}
}

func mySQLWithIgnoreMissing(ignore bool) mySQLVirtualNetworkRuleModifier {
	return func(r *v1alpha3.MySQLServerVirtualNetworkRule) {
		r.Spec.IgnoreMissingVnetServiceEndpoint = ignore
	}
}

func mySQLServerParameters(createMode *v1beta1.CreateMode) v1beta1.SQLServerParameters {
	fp := v1beta1.SQLServerParameters{
		CreateMode: createMode,
	}
	return fp
}

func mySQLServerPropertiesForDefaultCreate() mysql.BasicServerPropertiesForCreate {
	adminPassword := "admin"
	return &mysql.ServerPropertiesForDefaultCreate{
		AdministratorLoginPassword: &adminPassword,
		CreateMode:                 mysql.CreateModeDefault,
		StorageProfile:             &mysql.StorageProfile{},
	}
}

func mySQLServerPropertiesForRestore() mysql.BasicServerPropertiesForCreate {
	return &mysql.ServerPropertiesForRestore{
		CreateMode:     mysql.CreateModePointInTimeRestore,
		StorageProfile: &mysql.StorageProfile{},
	}
}

func mySQLServerPropertiesForGeoRestore() mysql.BasicServerPropertiesForCreate {
	return &mysql.ServerPropertiesForGeoRestore{
		CreateMode:     mysql.CreateModeGeoRestore,
		StorageProfile: &mysql.StorageProfile{},
	}
}

func mySQLServerPropertiesForReplica() mysql.BasicServerPropertiesForCreate {
	return &mysql.ServerPropertiesForReplica{
		CreateMode:     mysql.CreateModeReplica,
		StorageProfile: &mysql.StorageProfile{},
	}
}

func mySQLVirtualNetworkRule(sm ...mySQLVirtualNetworkRuleModifier) *v1alpha3.MySQLServerVirtualNetworkRule {
	r := &v1alpha3.MySQLServerVirtualNetworkRule{
		Spec: v1alpha3.MySQLVirtualNetworkRuleSpec{
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

func TestToMySQLProperties(t *testing.T) {
	cases := []struct {
		name string
		fp   v1beta1.SQLServerParameters
		want mysql.BasicServerPropertiesForCreate
	}{
		{
			name: "CreateModeDefault",
			fp:   mySQLServerParameters(pointerFromCreateMode(v1beta1.CreateModeDefault)),
			want: mySQLServerPropertiesForDefaultCreate(),
		},
		{
			name: "CreateModePointInTimeRestore",
			fp:   mySQLServerParameters(pointerFromCreateMode(v1beta1.CreateModePointInTimeRestore)),
			want: mySQLServerPropertiesForRestore(),
		},
		{
			name: "CreateModeGeoRestore",
			fp:   mySQLServerParameters(pointerFromCreateMode(v1beta1.CreateModeGeoRestore)),
			want: mySQLServerPropertiesForGeoRestore(),
		},
		{
			name: "CreateModeReplica",
			fp:   mySQLServerParameters(pointerFromCreateMode(v1beta1.CreateModeReplica)),
			want: mySQLServerPropertiesForReplica(),
		},
		{
			name: "ServerPropertiesForInvalidString",
			fp:   mySQLServerParameters(pointerFromCreateMode("")),
			want: mySQLServerPropertiesForDefaultCreate(),
		},
		{
			name: "ServerPropertiesForDefaultCreate",
			fp:   mySQLServerParameters(nil),
			want: mySQLServerPropertiesForDefaultCreate(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := toMySQLProperties(tc.fp, "admin")
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("TestToMySQLProperties(%s): -want, +got\n%s", tc.name, diff)
			}
		})
	}
}

func TestNewMySQLVirtualNetworkRuleParameters(t *testing.T) {
	cases := []struct {
		name string
		r    *v1alpha3.MySQLServerVirtualNetworkRule
		want mysql.VirtualNetworkRule
	}{
		{
			name: "Successful",
			r: mySQLVirtualNetworkRule(
				mySQLWithSubnetID(vnetSubnetID),
				mySQLWithIgnoreMissing(ignoreMissing),
			),
			want: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           to.StringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: to.BoolPtr(ignoreMissing),
				},
			},
		},
		{
			name: "SuccessfulPartial",
			r: mySQLVirtualNetworkRule(
				mySQLWithSubnetID(vnetSubnetID),
			),
			want: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           to.StringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: to.BoolPtr(false),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewMySQLVirtualNetworkRuleParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewMySQLVirtualNetworkRuleParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestMySQLServerVirtualNetworkRuleNeedsUpdate(t *testing.T) {
	cases := []struct {
		name string
		kube *v1alpha3.MySQLServerVirtualNetworkRule
		az   mysql.VirtualNetworkRule
		want bool
	}{
		{
			name: "NoUpdateNeeded",
			kube: mySQLVirtualNetworkRule(
				mySQLWithSubnetID(vnetSubnetID),
				mySQLWithIgnoreMissing(ignoreMissing),
			),
			az: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
				},
			},
			want: false,
		},
		{
			name: "UpdateNeededVirtualNetworkSubnetID",
			kube: mySQLVirtualNetworkRule(
				mySQLWithSubnetID(vnetSubnetID),
				mySQLWithIgnoreMissing(ignoreMissing),
			),
			az: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr("some/other/subnet"),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
				},
			},
			want: true,
		},
		{
			name: "UpdateNeededIgnoreMissingVnetServiceEndpoint",
			kube: mySQLVirtualNetworkRule(
				mySQLWithSubnetID(vnetSubnetID),
				mySQLWithIgnoreMissing(ignoreMissing),
			),
			az: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(!ignoreMissing),
				},
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MySQLServerVirtualNetworkRuleNeedsUpdate(tc.kube, tc.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("MySQLServerVirtualNetworkRuleNeedsUpdate(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestUpdateMySQLVirtualNetworkRuleStatusFromAzure(t *testing.T) {

	mockCondition := xpv1.Condition{Message: "mockMessage"}
	resourceStatus := xpv1.ResourceStatus{
		ConditionedStatus: xpv1.ConditionedStatus{
			Conditions: []xpv1.Condition{mockCondition},
		},
	}

	cases := []struct {
		name string
		r    mysql.VirtualNetworkRule
		want v1alpha3.VirtualNetworkRuleStatus
	}{
		{
			name: "SuccessfulFull",
			r: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				ID:   azure.ToStringPtr(id),
				Type: azure.ToStringPtr(resourceType),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
					State:                            mysql.VirtualNetworkRuleStateReady,
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
			r: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				ID:   azure.ToStringPtr(id),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
					State:                            mysql.VirtualNetworkRuleStateReady,
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
			v := &v1alpha3.MySQLServerVirtualNetworkRule{
				Status: v1alpha3.VirtualNetworkRuleStatus{
					ResourceStatus: resourceStatus,
				},
			}

			UpdateMySQLVirtualNetworkRuleStatusFromAzure(v, tc.r)

			// make sure that internal resource status hasn't changed
			if diff := cmp.Diff(mockCondition, v.Status.ResourceStatus.Conditions[0]); diff != "" {
				t.Errorf("UpdateMySQLVirtualNetworkRuleStatusFromAzure(...): -want, +got\n%s", diff)
			}

			// make sure that other resource parameters are updated
			tc.want.ResourceStatus = resourceStatus
			if diff := cmp.Diff(tc.want, v.Status); diff != "" {
				t.Errorf("UpdateMySQLVirtualNetworkRuleStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestNewMySQLFirewallRuleParameters(t *testing.T) {
	name := "coolrule"
	start := "127.0.0.1."
	end := "It was just a dream Bender - there's no such thing as two."

	cases := map[string]struct {
		r    *v1alpha3.MySQLServerFirewallRule
		want mysql.FirewallRule
	}{
		"Successful": {
			r: func() *v1alpha3.MySQLServerFirewallRule {
				r := &v1alpha3.MySQLServerFirewallRule{
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
			want: mysql.FirewallRule{
				Name: azure.ToStringPtr(name),
				FirewallRuleProperties: &mysql.FirewallRuleProperties{
					StartIPAddress: azure.ToStringPtr(start),
					EndIPAddress:   azure.ToStringPtr(end),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := NewMySQLFirewallRuleParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewMySQLFirewallRuleParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestMySQLServerFirewallRuleIsUpToDate(t *testing.T) {
	start := "127.0.0.1."
	end := "256"

	cases := map[string]struct {
		kube *v1alpha3.MySQLServerFirewallRule
		az   mysql.FirewallRule
		want bool
	}{
		"UpToDate": {
			kube: &v1alpha3.MySQLServerFirewallRule{},
			az: mysql.FirewallRule{
				Name:                   azure.ToStringPtr(vnetRuleName),
				FirewallRuleProperties: &mysql.FirewallRuleProperties{},
			},
			want: true,
		},
		"StartNeedsUpdate": {
			kube: &v1alpha3.MySQLServerFirewallRule{
				Spec: v1alpha3.FirewallRuleSpec{ForProvider: v1alpha3.FirewallRuleParameters{FirewallRuleProperties: v1alpha3.FirewallRuleProperties{
					StartIPAddress: start,
					EndIPAddress:   end,
				}}},
			},
			az: mysql.FirewallRule{
				FirewallRuleProperties: &mysql.FirewallRuleProperties{
					StartIPAddress: azure.ToStringPtr("255.255.255.254"),
					EndIPAddress:   azure.ToStringPtr(end),
				},
			},
			want: false,
		},
		"EndNeedsUpdate": {
			kube: &v1alpha3.MySQLServerFirewallRule{
				Spec: v1alpha3.FirewallRuleSpec{ForProvider: v1alpha3.FirewallRuleParameters{FirewallRuleProperties: v1alpha3.FirewallRuleProperties{
					StartIPAddress: start,
					EndIPAddress:   end,
				}}},
			},
			az: mysql.FirewallRule{
				FirewallRuleProperties: &mysql.FirewallRuleProperties{
					StartIPAddress: azure.ToStringPtr(start),
					EndIPAddress:   azure.ToStringPtr("192.168.0.1"),
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := MySQLServerFirewallRuleIsUpToDate(tc.kube, tc.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("MySQLServerFirewallRuleIsUpToDate(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestIsMysqlUpToDate(t *testing.T) {
	type args struct {
		p  v1beta1.SQLServerParameters
		in mysql.Server
	}
	cases := map[string]struct {
		args
		want bool
	}{
		"IsUpToDateWithAllDefault": {
			args: args{
				p: v1beta1.SQLServerParameters{},
				in: mysql.Server{
					Sku: &mysql.Sku{},
					ServerProperties: &mysql.ServerProperties{
						StorageProfile: &mysql.StorageProfile{},
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
					Version:           "8.0.15",
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
				in: mysql.Server{
					Tags: map[string]*string{
						"created_by": azure.ToStringPtr("crossplane"),
					},
					Sku: &mysql.Sku{
						Tier:     mysql.GeneralPurpose,
						Capacity: azure.ToInt32Ptr(2),
						Family:   azure.ToStringPtr("Gen5"),
					},
					ServerProperties: &mysql.ServerProperties{
						Version: "8.0.15",
						StorageProfile: &mysql.StorageProfile{
							StorageMB:           azure.ToInt32Ptr(20480),
							StorageAutogrow:     mysql.StorageAutogrowEnabled,
							BackupRetentionDays: azure.ToInt32Ptr(5),
							GeoRedundantBackup:  mysql.Disabled,
						},
						SslEnforcement:      mysql.SslEnforcementEnumEnabled,
						MinimalTLSVersion:   mysql.TLS12,
						PublicNetworkAccess: mysql.PublicNetworkAccessEnumEnabled,
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
				in: mysql.Server{
					Sku: &mysql.Sku{},
					ServerProperties: &mysql.ServerProperties{
						StorageProfile:      &mysql.StorageProfile{},
						PublicNetworkAccess: mysql.PublicNetworkAccessEnumEnabled,
					},
				},
			},
			want: false,
		},
		"IsNotUpToDateWithServerWithoutSku": {
			args: args{
				p: v1beta1.SQLServerParameters{},
				in: mysql.Server{
					ServerProperties: &mysql.ServerProperties{
						StorageProfile: &mysql.StorageProfile{},
					},
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsMySQLUpToDate(tc.args.p, tc.args.in)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("IsMySQLUpToDate(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestLateInitializeMySQL(t *testing.T) {
	type args struct {
		p  *v1beta1.SQLServerParameters
		in mysql.Server
	}
	cases := map[string]struct {
		args
		want *v1beta1.SQLServerParameters
	}{
		"PublicNetworkAccessLateInitialize": {
			args: args{
				p: &v1beta1.SQLServerParameters{},
				in: mysql.Server{
					Sku: &mysql.Sku{},
					ServerProperties: &mysql.ServerProperties{
						PublicNetworkAccess: mysql.PublicNetworkAccessEnumEnabled,
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
			LateInitializeMySQL(tc.args.p, tc.args.in)
			if diff := cmp.Diff(tc.want, tc.args.p); diff != "" {
				t.Errorf("TestLateInitializeMySQL(...): -want, +got\n%s", diff)
			}
		})
	}
}
