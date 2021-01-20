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
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	azuredbv1alpha3 "github.com/crossplane/provider-azure/apis/database/v1alpha3"
	"github.com/crossplane/provider-azure/apis/database/v1beta1"
	azuredbv1beta1 "github.com/crossplane/provider-azure/apis/database/v1beta1"
	"github.com/crossplane/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

// NOTE: postgresql and mysql structs and functions live in their respective
// packages even though they are exactly the same. However, Crossplane does not
// make that assumption and use the respective package for each type, although,
// they both share the same SQLServerParameters and SQLServerObservation objects.
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/mysql/mgmt/2017-12-01/mysql/models.go
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/postgresql/mgmt/2017-12-01/postgresql/models.go

var (
	skuShortTiers = map[mysql.SkuTier]string{
		mysql.Basic:           "B",
		mysql.GeneralPurpose:  "GP",
		mysql.MemoryOptimized: "MO",
	}
)

// MySQLServerAPI represents the API interface for a MySQL Server client
type MySQLServerAPI interface {
	GetServer(ctx context.Context, s *azuredbv1beta1.MySQLServer) (mysql.Server, error)
	CreateServer(ctx context.Context, s *azuredbv1beta1.MySQLServer, adminPassword string) error
	UpdateServer(ctx context.Context, s *azuredbv1beta1.MySQLServer) error
	DeleteServer(ctx context.Context, s *azuredbv1beta1.MySQLServer) error
	GetRESTClient() autorest.Sender
}

// MySQLServerClient is the concrete implementation of the MySQLServerAPI
// interface for MySQL that calls Azure API.
type MySQLServerClient struct {
	mysql.ServersClient
}

// NewMySQLServerClient creates and initializes a MySQLServerClient instance.
func NewMySQLServerClient(cl mysql.ServersClient) *MySQLServerClient {
	return &MySQLServerClient{
		ServersClient: cl,
	}
}

// GetRESTClient returns the underlying REST client that the client object uses.
func (c *MySQLServerClient) GetRESTClient() autorest.Sender {
	return c.ServersClient.Client
}

// GetServer retrieves the requested MySQL Server
func (c *MySQLServerClient) GetServer(ctx context.Context, cr *azuredbv1beta1.MySQLServer) (mysql.Server, error) {
	return c.ServersClient.Get(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
}

// toMySQLProperties converts the CrossPlane ForProvider object to a MySQL Azure properties object
func toMySQLProperties(s v1beta1.SQLServerParameters, adminPassword string) mysql.BasicServerPropertiesForCreate {
	createMode := pointerToCreateMode(s.CreateMode)
	switch createMode {
	case azuredbv1beta1.CreateModePointInTimeRestore:
		return &mysql.ServerPropertiesForRestore{
			MinimalTLSVersion:  mysql.MinimalTLSVersionEnum(s.MinimalTLSVersion),
			Version:            mysql.ServerVersion(s.Version),
			SslEnforcement:     mysql.SslEnforcementEnum(s.SSLEnforcement),
			CreateMode:         mysql.CreateMode(createMode),
			RestorePointInTime: safeDate(s.RestorePointInTime),
			SourceServerID:     s.SourceServerID,
			StorageProfile: &mysql.StorageProfile{
				BackupRetentionDays: azure.ToInt32PtrFromIntPtr(s.StorageProfile.BackupRetentionDays),
				GeoRedundantBackup:  mysql.GeoRedundantBackup(azure.ToString(s.StorageProfile.GeoRedundantBackup)),
				StorageMB:           azure.ToInt32Ptr(s.StorageProfile.StorageMB),
				StorageAutogrow:     mysql.StorageAutogrow(azure.ToString(s.StorageProfile.StorageAutogrow)),
			},
		}
	case azuredbv1beta1.CreateModeGeoRestore:
		return &mysql.ServerPropertiesForGeoRestore{
			MinimalTLSVersion: mysql.MinimalTLSVersionEnum(s.MinimalTLSVersion),
			Version:           mysql.ServerVersion(s.Version),
			SslEnforcement:    mysql.SslEnforcementEnum(s.SSLEnforcement),
			CreateMode:        mysql.CreateMode(createMode),
			SourceServerID:    s.SourceServerID,
			StorageProfile: &mysql.StorageProfile{
				BackupRetentionDays: azure.ToInt32PtrFromIntPtr(s.StorageProfile.BackupRetentionDays),
				GeoRedundantBackup:  mysql.GeoRedundantBackup(azure.ToString(s.StorageProfile.GeoRedundantBackup)),
				StorageMB:           azure.ToInt32Ptr(s.StorageProfile.StorageMB),
				StorageAutogrow:     mysql.StorageAutogrow(azure.ToString(s.StorageProfile.StorageAutogrow)),
			},
		}
	case azuredbv1beta1.CreateModeReplica:
		return &mysql.ServerPropertiesForReplica{
			MinimalTLSVersion: mysql.MinimalTLSVersionEnum(s.MinimalTLSVersion),
			Version:           mysql.ServerVersion(s.Version),
			SslEnforcement:    mysql.SslEnforcementEnum(s.SSLEnforcement),
			CreateMode:        mysql.CreateMode(createMode),
			SourceServerID:    s.SourceServerID,
			StorageProfile: &mysql.StorageProfile{
				BackupRetentionDays: azure.ToInt32PtrFromIntPtr(s.StorageProfile.BackupRetentionDays),
				GeoRedundantBackup:  mysql.GeoRedundantBackup(azure.ToString(s.StorageProfile.GeoRedundantBackup)),
				StorageMB:           azure.ToInt32Ptr(s.StorageProfile.StorageMB),
				StorageAutogrow:     mysql.StorageAutogrow(azure.ToString(s.StorageProfile.StorageAutogrow)),
			},
		}
	case azuredbv1beta1.CreateModeDefault:
		fallthrough
	default:
		return &mysql.ServerPropertiesForDefaultCreate{
			MinimalTLSVersion:          mysql.MinimalTLSVersionEnum(s.MinimalTLSVersion),
			AdministratorLogin:         azure.ToStringPtr(s.AdministratorLogin),
			AdministratorLoginPassword: &adminPassword,
			Version:                    mysql.ServerVersion(s.Version),
			SslEnforcement:             mysql.SslEnforcementEnum(s.SSLEnforcement),
			CreateMode:                 mysql.CreateMode(createMode),
			StorageProfile: &mysql.StorageProfile{
				BackupRetentionDays: azure.ToInt32PtrFromIntPtr(s.StorageProfile.BackupRetentionDays),
				GeoRedundantBackup:  mysql.GeoRedundantBackup(azure.ToString(s.StorageProfile.GeoRedundantBackup)),
				StorageMB:           azure.ToInt32Ptr(s.StorageProfile.StorageMB),
				StorageAutogrow:     mysql.StorageAutogrow(azure.ToString(s.StorageProfile.StorageAutogrow)),
			},
		}
	}
}

// CreateServer creates a MySQL Server.
func (c *MySQLServerClient) CreateServer(ctx context.Context, cr *azuredbv1beta1.MySQLServer, adminPassword string) error {
	s := cr.Spec.ForProvider
	sku, err := ToMySQLSKU(s.SKU)
	if err != nil {
		return err
	}
	createParams := mysql.ServerForCreate{
		Sku:        sku,
		Properties: toMySQLProperties(s, adminPassword),
		Location:   &s.Location,
		Tags:       azure.ToStringPtrMap(s.Tags),
	}
	op, err := c.Create(ctx, s.ResourceGroupName, meta.GetExternalName(cr), createParams)
	if err != nil {
		return err
	}
	cr.Status.AtProvider.LastOperation = v1alpha3.AsyncOperation{
		PollingURL: op.PollingURL(),
		Method:     http.MethodPut,
	}
	return nil
}

// UpdateServer updates a MySQL Server.
func (c *MySQLServerClient) UpdateServer(ctx context.Context, cr *azuredbv1beta1.MySQLServer) error {
	// TODO(muvaf): password update via Update call is supported by Azure but
	// we don't support that.
	s := cr.Spec.ForProvider
	properties := &mysql.ServerUpdateParametersProperties{
		Version:           mysql.ServerVersion(s.Version),
		MinimalTLSVersion: mysql.MinimalTLSVersionEnum(s.MinimalTLSVersion),
		SslEnforcement:    mysql.SslEnforcementEnum(s.SSLEnforcement),
		StorageProfile: &mysql.StorageProfile{
			BackupRetentionDays: azure.ToInt32PtrFromIntPtr(s.StorageProfile.BackupRetentionDays),
			GeoRedundantBackup:  mysql.GeoRedundantBackup(azure.ToString(s.StorageProfile.GeoRedundantBackup)),
			StorageMB:           azure.ToInt32Ptr(s.StorageProfile.StorageMB),
			StorageAutogrow:     mysql.StorageAutogrow(azure.ToString(s.StorageProfile.StorageAutogrow)),
		},
	}
	sku, err := ToMySQLSKU(s.SKU)
	if err != nil {
		return err
	}
	updateParams := mysql.ServerUpdateParameters{
		Sku:                              sku,
		ServerUpdateParametersProperties: properties,
		Tags:                             azure.ToStringPtrMap(s.Tags),
	}
	op, err := c.Update(ctx, s.ResourceGroupName, meta.GetExternalName(cr), updateParams)
	if err != nil {
		return err
	}
	cr.Status.AtProvider.LastOperation = v1alpha3.AsyncOperation{
		PollingURL: op.PollingURL(),
		Method:     http.MethodPatch,
	}
	return nil
}

// DeleteServer deletes the given MySQLServer resource.
func (c *MySQLServerClient) DeleteServer(ctx context.Context, cr *azuredbv1beta1.MySQLServer) error {
	op, err := c.ServersClient.Delete(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
	if err != nil {
		return err
	}
	cr.Status.AtProvider.LastOperation = v1alpha3.AsyncOperation{
		PollingURL: op.PollingURL(),
		Method:     http.MethodDelete,
	}
	return nil
}

// NewMySQLVirtualNetworkRuleParameters returns an Azure VirtualNetworkRule object from a virtual network spec
func NewMySQLVirtualNetworkRuleParameters(v *azuredbv1alpha3.MySQLServerVirtualNetworkRule) mysql.VirtualNetworkRule {
	return mysql.VirtualNetworkRule{
		Name: azure.ToStringPtr(meta.GetExternalName(v)),
		VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
			VirtualNetworkSubnetID:           azure.ToStringPtr(v.Spec.VirtualNetworkRuleProperties.VirtualNetworkSubnetID),
			IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(v.Spec.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint, azure.FieldRequired),
		},
	}
}

// MySQLServerVirtualNetworkRuleNeedsUpdate determines if a virtual network rule needs to be updated
func MySQLServerVirtualNetworkRuleNeedsUpdate(kube *azuredbv1alpha3.MySQLServerVirtualNetworkRule, az mysql.VirtualNetworkRule) bool {
	up := NewMySQLVirtualNetworkRuleParameters(kube)

	switch {
	case !reflect.DeepEqual(up.VirtualNetworkRuleProperties.VirtualNetworkSubnetID, az.VirtualNetworkRuleProperties.VirtualNetworkSubnetID):
		return true
	case !reflect.DeepEqual(up.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint, az.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint):
		return true
	}

	return false
}

// UpdateMySQLVirtualNetworkRuleStatusFromAzure updates the status related to the external
// Azure MySQLVirtualNetworkRule in the VirtualNetworkStatus
func UpdateMySQLVirtualNetworkRuleStatusFromAzure(v *azuredbv1alpha3.MySQLServerVirtualNetworkRule, az mysql.VirtualNetworkRule) {
	v.Status.State = string(az.VirtualNetworkRuleProperties.State)
	v.Status.ID = azure.ToString(az.ID)
	v.Status.Type = azure.ToString(az.Type)
}

// NewMySQLFirewallRuleParameters returns an Azure FirewallRule object from a
// firewall spec.
func NewMySQLFirewallRuleParameters(r *azuredbv1alpha3.MySQLServerFirewallRule) mysql.FirewallRule {
	return mysql.FirewallRule{
		Name: azure.ToStringPtr(meta.GetExternalName(r)),
		FirewallRuleProperties: &mysql.FirewallRuleProperties{
			StartIPAddress: azure.ToStringPtr(r.Spec.ForProvider.StartIPAddress),
			EndIPAddress:   azure.ToStringPtr(r.Spec.ForProvider.EndIPAddress),
		},
	}
}

// MySQLServerFirewallRuleIsUpToDate returns true if the supplied FirewallRule
// appears to be up to date with the supplied MySQLServerFirewallRule.
func MySQLServerFirewallRuleIsUpToDate(kube *azuredbv1alpha3.MySQLServerFirewallRule, az mysql.FirewallRule) bool {
	up := NewMySQLFirewallRuleParameters(kube)
	return cmp.Equal(up.FirewallRuleProperties, az.FirewallRuleProperties)
}

// The name must match the specification of the SKU, so, we don't allow user
// to specify an arbitrary name. The format is tier + family + cores, e.g. B_Gen4_1, GP_Gen5_8.

// ToMySQLSKU returns a *mysql.Sku object that can be used in Azure API calls.
func ToMySQLSKU(skuSpec azuredbv1beta1.SKU) (*mysql.Sku, error) {
	t, ok := skuShortTiers[mysql.SkuTier(skuSpec.Tier)]
	if !ok {
		return nil, fmt.Errorf("tier '%s' is not one of the supported values: %+v", skuSpec.Tier, mysql.PossibleSkuTierValues())
	}
	return &mysql.Sku{
		Name:     azure.ToStringPtr(fmt.Sprintf("%s_%s_%s", t, skuSpec.Family, strconv.Itoa(skuSpec.Capacity))),
		Tier:     mysql.SkuTier(skuSpec.Tier),
		Capacity: azure.ToInt32Ptr(skuSpec.Capacity),
		Family:   azure.ToStringPtr(skuSpec.Family),
		Size:     skuSpec.Size,
	}, nil
}

// UpdateMySQLObservation produces SQLServerObservation from mysql.Server.
func UpdateMySQLObservation(o *azuredbv1beta1.SQLServerObservation, in mysql.Server) {
	o.ID = azure.ToString(in.ID)
	o.Name = azure.ToString(in.Name)
	o.Type = azure.ToString(in.Type)
	o.UserVisibleState = string(in.UserVisibleState)
	o.FullyQualifiedDomainName = azure.ToString(in.FullyQualifiedDomainName)
	o.MasterServerID = azure.ToString(in.MasterServerID)
}

// LateInitializeMySQL fills the empty values of SQLServerParameters with the
// ones that are retrieved from the Azure API.
func LateInitializeMySQL(p *azuredbv1beta1.SQLServerParameters, in mysql.Server) {
	if in.Sku != nil {
		p.SKU.Size = azure.LateInitializeStringPtrFromPtr(p.SKU.Size, in.Sku.Size)
	}
	p.Tags = azure.LateInitializeStringMap(p.Tags, in.Tags)
	if in.StorageProfile != nil {
		p.StorageProfile.BackupRetentionDays = azure.LateInitializeIntPtrFromInt32Ptr(p.StorageProfile.BackupRetentionDays, in.StorageProfile.BackupRetentionDays)
		p.StorageProfile.GeoRedundantBackup = azure.LateInitializeStringPtrFromVal(p.StorageProfile.GeoRedundantBackup, string(in.StorageProfile.GeoRedundantBackup))
		p.StorageProfile.StorageAutogrow = azure.LateInitializeStringPtrFromVal(p.StorageProfile.StorageAutogrow, string(in.StorageProfile.StorageAutogrow))
	}
}

// IsMySQLUpToDate is used to report whether given mysql.Server is in
// sync with the SQLServerParameters that user desires.
func IsMySQLUpToDate(p azuredbv1beta1.SQLServerParameters, in mysql.Server) bool { // nolint:gocyclo
	if in.StorageProfile == nil || in.Sku == nil {
		return false
	}
	switch {
	case p.MinimalTLSVersion != azuredbv1beta1.MinimalTLSVersionEnum(in.MinimalTLSVersion):
		return false
	case p.SSLEnforcement != string(in.SslEnforcement):
		return false
	case p.Version != string(in.Version):
		return false
	case !reflect.DeepEqual(azure.ToStringPtrMap(p.Tags), in.Tags):
		return false
	case p.SKU.Tier != string(in.Sku.Tier):
		return false
	case p.SKU.Capacity != azure.ToInt(in.Sku.Capacity):
		return false
	case p.SKU.Family != azure.ToString(in.Sku.Family):
		return false
	case !reflect.DeepEqual(azure.ToInt32PtrFromIntPtr(p.StorageProfile.BackupRetentionDays), in.StorageProfile.BackupRetentionDays):
		return false
	case azure.ToString(p.StorageProfile.GeoRedundantBackup) != string(in.StorageProfile.GeoRedundantBackup):
		return false
	case p.StorageProfile.StorageMB != azure.ToInt(in.StorageProfile.StorageMB):
		return false
	case azure.ToString(p.StorageProfile.StorageAutogrow) != string(in.StorageProfile.StorageAutogrow):
		return false
	}
	return true
}
