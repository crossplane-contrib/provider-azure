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
	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
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

// PostgreSQLServerAPI represents the API interface for a PostgreSQL Server client
type PostgreSQLServerAPI interface {
	GetServer(ctx context.Context, s *azuredbv1beta1.PostgreSQLServer) (postgresql.Server, error)
	CreateServer(ctx context.Context, s *azuredbv1beta1.PostgreSQLServer, adminPassword string) error
	DeleteServer(ctx context.Context, s *azuredbv1beta1.PostgreSQLServer) error
	UpdateServer(ctx context.Context, s *azuredbv1beta1.PostgreSQLServer) error
	GetRESTClient() autorest.Sender
}

// PostgreSQLServerClient is the concreate implementation of the SQLServerAPI interface for PostgreSQL that calls Azure API.
type PostgreSQLServerClient struct {
	postgresql.ServersClient
}

// NewPostgreSQLServerClient creates and initializes a PostgreSQLServerClient instance.
func NewPostgreSQLServerClient(cl postgresql.ServersClient) *PostgreSQLServerClient {
	return &PostgreSQLServerClient{
		ServersClient: cl,
	}
}

// GetRESTClient returns the underlying REST client that the client object uses.
func (c *PostgreSQLServerClient) GetRESTClient() autorest.Sender {
	return c.ServersClient.Client
}

// GetServer retrieves the requested PostgreSQL Server
func (c *PostgreSQLServerClient) GetServer(ctx context.Context, cr *azuredbv1beta1.PostgreSQLServer) (postgresql.Server, error) {
	return c.ServersClient.Get(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
}

// toPGSQLProperties converts the CrossPlane ForProvider object to a PostgreSQL Azure properties object
func toPGSQLProperties(s v1beta1.SQLServerParameters, adminPassword string) postgresql.BasicServerPropertiesForCreate {
	createMode := pointerToCreateMode(s.CreateMode)
	switch createMode {
	case azuredbv1beta1.CreateModePointInTimeRestore:
		return &postgresql.ServerPropertiesForRestore{
			MinimalTLSVersion:   postgresql.MinimalTLSVersionEnum(s.MinimalTLSVersion),
			Version:             postgresql.ServerVersion(s.Version),
			SslEnforcement:      postgresql.SslEnforcementEnum(s.SSLEnforcement),
			PublicNetworkAccess: postgresql.PublicNetworkAccessEnum(azure.ToString(s.PublicNetworkAccess)),
			CreateMode:          postgresql.CreateModePointInTimeRestore,
			RestorePointInTime:  safeDate(s.RestorePointInTime),
			SourceServerID:      s.SourceServerID,
			StorageProfile: &postgresql.StorageProfile{
				BackupRetentionDays: azure.ToInt32PtrFromIntPtr(s.StorageProfile.BackupRetentionDays),
				GeoRedundantBackup:  postgresql.GeoRedundantBackup(azure.ToString(s.StorageProfile.GeoRedundantBackup)),
				StorageMB:           azure.ToInt32Ptr(s.StorageProfile.StorageMB),
				StorageAutogrow:     postgresql.StorageAutogrow(azure.ToString(s.StorageProfile.StorageAutogrow)),
			},
		}
	case azuredbv1beta1.CreateModeGeoRestore:
		return &postgresql.ServerPropertiesForGeoRestore{
			MinimalTLSVersion:   postgresql.MinimalTLSVersionEnum(s.MinimalTLSVersion),
			Version:             postgresql.ServerVersion(s.Version),
			SslEnforcement:      postgresql.SslEnforcementEnum(s.SSLEnforcement),
			SourceServerID:      s.SourceServerID,
			PublicNetworkAccess: postgresql.PublicNetworkAccessEnum(azure.ToString(s.PublicNetworkAccess)),
			CreateMode:          postgresql.CreateModeGeoRestore,
			StorageProfile: &postgresql.StorageProfile{
				BackupRetentionDays: azure.ToInt32PtrFromIntPtr(s.StorageProfile.BackupRetentionDays),
				GeoRedundantBackup:  postgresql.GeoRedundantBackup(azure.ToString(s.StorageProfile.GeoRedundantBackup)),
				StorageMB:           azure.ToInt32Ptr(s.StorageProfile.StorageMB),
				StorageAutogrow:     postgresql.StorageAutogrow(azure.ToString(s.StorageProfile.StorageAutogrow)),
			},
		}
	case azuredbv1beta1.CreateModeReplica:
		return &postgresql.ServerPropertiesForReplica{
			MinimalTLSVersion:   postgresql.MinimalTLSVersionEnum(s.MinimalTLSVersion),
			Version:             postgresql.ServerVersion(s.Version),
			SslEnforcement:      postgresql.SslEnforcementEnum(s.SSLEnforcement),
			PublicNetworkAccess: postgresql.PublicNetworkAccessEnum(azure.ToString(s.PublicNetworkAccess)),
			CreateMode:          postgresql.CreateModeReplica,
			SourceServerID:      s.SourceServerID,
			StorageProfile: &postgresql.StorageProfile{
				BackupRetentionDays: azure.ToInt32PtrFromIntPtr(s.StorageProfile.BackupRetentionDays),
				GeoRedundantBackup:  postgresql.GeoRedundantBackup(azure.ToString(s.StorageProfile.GeoRedundantBackup)),
				StorageMB:           azure.ToInt32Ptr(s.StorageProfile.StorageMB),
				StorageAutogrow:     postgresql.StorageAutogrow(azure.ToString(s.StorageProfile.StorageAutogrow)),
			},
		}
	case azuredbv1beta1.CreateModeDefault:
		fallthrough
	default:
		return &postgresql.ServerPropertiesForDefaultCreate{
			MinimalTLSVersion:          postgresql.MinimalTLSVersionEnum(s.MinimalTLSVersion),
			AdministratorLogin:         azure.ToStringPtr(s.AdministratorLogin),
			AdministratorLoginPassword: &adminPassword,
			Version:                    postgresql.ServerVersion(s.Version),
			SslEnforcement:             postgresql.SslEnforcementEnum(s.SSLEnforcement),
			PublicNetworkAccess:        postgresql.PublicNetworkAccessEnum(azure.ToString(s.PublicNetworkAccess)),
			CreateMode:                 postgresql.CreateModeDefault,
			StorageProfile: &postgresql.StorageProfile{
				BackupRetentionDays: azure.ToInt32PtrFromIntPtr(s.StorageProfile.BackupRetentionDays),
				GeoRedundantBackup:  postgresql.GeoRedundantBackup(azure.ToString(s.StorageProfile.GeoRedundantBackup)),
				StorageMB:           azure.ToInt32Ptr(s.StorageProfile.StorageMB),
				StorageAutogrow:     postgresql.StorageAutogrow(azure.ToString(s.StorageProfile.StorageAutogrow)),
			},
		}
	}
}

// CreateServer creates a PostgreSQL Server
func (c *PostgreSQLServerClient) CreateServer(ctx context.Context, cr *azuredbv1beta1.PostgreSQLServer, adminPassword string) error {
	s := cr.Spec.ForProvider
	sku, err := ToPostgreSQLSKU(s.SKU)
	if err != nil {
		return err
	}
	createParams := postgresql.ServerForCreate{
		Sku:        sku,
		Properties: toPGSQLProperties(s, adminPassword),
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

// UpdateServer updates a PostgreSQL Server.
func (c *PostgreSQLServerClient) UpdateServer(ctx context.Context, cr *azuredbv1beta1.PostgreSQLServer) error {
	// TODO(muvaf): password update via Update call is supported by Azure but
	// we don't support that.
	s := cr.Spec.ForProvider
	properties := &postgresql.ServerUpdateParametersProperties{
		Version:             postgresql.ServerVersion(s.Version),
		MinimalTLSVersion:   postgresql.MinimalTLSVersionEnum(s.MinimalTLSVersion),
		SslEnforcement:      postgresql.SslEnforcementEnum(s.SSLEnforcement),
		PublicNetworkAccess: postgresql.PublicNetworkAccessEnum(azure.ToString(s.PublicNetworkAccess)),
		StorageProfile: &postgresql.StorageProfile{
			BackupRetentionDays: azure.ToInt32PtrFromIntPtr(s.StorageProfile.BackupRetentionDays),
			GeoRedundantBackup:  postgresql.GeoRedundantBackup(azure.ToString(s.StorageProfile.GeoRedundantBackup)),
			StorageMB:           azure.ToInt32Ptr(s.StorageProfile.StorageMB),
			StorageAutogrow:     postgresql.StorageAutogrow(azure.ToString(s.StorageProfile.StorageAutogrow)),
		},
	}
	sku, err := ToPostgreSQLSKU(s.SKU)
	if err != nil {
		return err
	}
	updateParams := postgresql.ServerUpdateParameters{
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

// DeleteServer deletes the given PostgreSQL resource
func (c *PostgreSQLServerClient) DeleteServer(ctx context.Context, cr *azuredbv1beta1.PostgreSQLServer) error {
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

// NewPostgreSQLVirtualNetworkRuleParameters returns an Azure VirtualNetworkRule object from a virtual network spec
func NewPostgreSQLVirtualNetworkRuleParameters(v *azuredbv1alpha3.PostgreSQLServerVirtualNetworkRule) postgresql.VirtualNetworkRule {
	return postgresql.VirtualNetworkRule{
		Name: azure.ToStringPtr(meta.GetExternalName(v)),
		VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
			VirtualNetworkSubnetID:           azure.ToStringPtr(v.Spec.VirtualNetworkRuleProperties.VirtualNetworkSubnetID),
			IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(v.Spec.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint, azure.FieldRequired),
		},
	}
}

// PostgreSQLServerVirtualNetworkRuleNeedsUpdate determines if a virtual network rule needs to be updated
func PostgreSQLServerVirtualNetworkRuleNeedsUpdate(kube *azuredbv1alpha3.PostgreSQLServerVirtualNetworkRule, az postgresql.VirtualNetworkRule) bool {
	up := NewPostgreSQLVirtualNetworkRuleParameters(kube)

	switch {
	case !reflect.DeepEqual(up.VirtualNetworkRuleProperties.VirtualNetworkSubnetID, az.VirtualNetworkRuleProperties.VirtualNetworkSubnetID):
		return true
	case !reflect.DeepEqual(up.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint, az.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint):
		return true
	}

	return false
}

// UpdatePostgreSQLVirtualNetworkRuleStatusFromAzure updates the status related to the external
// Azure PostgreSQLVirtualNetworkRule in the VirtualNetworkStatus
func UpdatePostgreSQLVirtualNetworkRuleStatusFromAzure(v *azuredbv1alpha3.PostgreSQLServerVirtualNetworkRule, az postgresql.VirtualNetworkRule) {
	v.Status.State = string(az.VirtualNetworkRuleProperties.State)
	v.Status.ID = azure.ToString(az.ID)
	v.Status.Type = azure.ToString(az.Type)
}

// NewPostgreSQLFirewallRuleParameters returns an Azure FirewallRule object from a
// firewall spec.
func NewPostgreSQLFirewallRuleParameters(r *azuredbv1alpha3.PostgreSQLServerFirewallRule) postgresql.FirewallRule {
	return postgresql.FirewallRule{
		Name: azure.ToStringPtr(meta.GetExternalName(r)),
		FirewallRuleProperties: &postgresql.FirewallRuleProperties{
			StartIPAddress: azure.ToStringPtr(r.Spec.ForProvider.StartIPAddress),
			EndIPAddress:   azure.ToStringPtr(r.Spec.ForProvider.EndIPAddress),
		},
	}
}

// PostgreSQLServerFirewallRuleIsUpToDate returns true if the supplied FirewallRule
// appears to be up to date with the supplied PostgreSQLServerFirewallRule.
func PostgreSQLServerFirewallRuleIsUpToDate(kube *azuredbv1alpha3.PostgreSQLServerFirewallRule, az postgresql.FirewallRule) bool {
	up := NewPostgreSQLFirewallRuleParameters(kube)
	return cmp.Equal(up.FirewallRuleProperties, az.FirewallRuleProperties)
}

// The name must match the specification of the SKU, so, we don't allow user
// to specify an arbitrary name. The format is tier + family + cores, e.g. B_Gen4_1, GP_Gen5_8.

// ToPostgreSQLSKU returns a *postgresql.Sku object that can be used in Azure API calls.
func ToPostgreSQLSKU(skuSpec azuredbv1beta1.SKU) (*postgresql.Sku, error) {
	t, ok := skuShortTiers[mysql.SkuTier(skuSpec.Tier)]
	if !ok {
		return nil, fmt.Errorf("tier '%s' is not one of the supported values: %+v", skuSpec.Tier, mysql.PossibleSkuTierValues())
	}
	return &postgresql.Sku{
		Name:     azure.ToStringPtr(fmt.Sprintf("%s_%s_%s", t, skuSpec.Family, strconv.Itoa(skuSpec.Capacity))),
		Tier:     postgresql.SkuTier(skuSpec.Tier),
		Capacity: azure.ToInt32Ptr(skuSpec.Capacity),
		Family:   azure.ToStringPtr(skuSpec.Family),
		Size:     skuSpec.Size,
	}, nil
}

// UpdatePostgreSQLObservation produces SQLServerObservation from postgresql.Server.
func UpdatePostgreSQLObservation(o *azuredbv1beta1.SQLServerObservation, in postgresql.Server) {
	o.ID = azure.ToString(in.ID)
	o.Name = azure.ToString(in.Name)
	o.Type = azure.ToString(in.Type)
	o.UserVisibleState = string(in.UserVisibleState)
	o.FullyQualifiedDomainName = azure.ToString(in.FullyQualifiedDomainName)
	o.MasterServerID = azure.ToString(in.MasterServerID)
}

// LateInitializePostgreSQL fills the empty values of SQLServerParameters with
// the ones that are retrieved from the Azure API. Returns true if the params
// were late initialized.
func LateInitializePostgreSQL(p *azuredbv1beta1.SQLServerParameters, in postgresql.Server) bool {
	before := p.DeepCopy()
	if in.Sku != nil {
		p.SKU.Size = azure.LateInitializeStringPtrFromPtr(p.SKU.Size, in.Sku.Size)
	}
	p.Tags = azure.LateInitializeStringMap(p.Tags, in.Tags)
	if in.StorageProfile != nil {
		p.StorageProfile.BackupRetentionDays = azure.LateInitializeIntPtrFromInt32Ptr(p.StorageProfile.BackupRetentionDays, in.StorageProfile.BackupRetentionDays)
		p.StorageProfile.GeoRedundantBackup = azure.LateInitializeStringPtrFromVal(p.StorageProfile.GeoRedundantBackup, string(in.StorageProfile.GeoRedundantBackup))
		p.StorageProfile.StorageAutogrow = azure.LateInitializeStringPtrFromVal(p.StorageProfile.StorageAutogrow, string(in.StorageProfile.StorageAutogrow))
	}
	if p.MinimalTLSVersion == "" {
		p.MinimalTLSVersion = string(in.MinimalTLSVersion)
	}
	if p.SSLEnforcement == "" {
		p.SSLEnforcement = string(in.SslEnforcement)
	}

	if p.PublicNetworkAccess == nil {
		p.PublicNetworkAccess = azure.ToStringPtr(string(in.PublicNetworkAccess))
	}

	return !cmp.Equal(before, p)
}

// IsPostgreSQLUpToDate is used to report whether given postgresql.Server is in
// sync with the SQLServerParameters that user desires.
func IsPostgreSQLUpToDate(p azuredbv1beta1.SQLServerParameters, in postgresql.Server) bool { // nolint:gocyclo
	if in.StorageProfile == nil || in.Sku == nil {
		return false
	}
	switch {
	case p.MinimalTLSVersion != string(in.MinimalTLSVersion) && p.SSLEnforcement != string(postgresql.SslEnforcementEnumDisabled):
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
	case azure.ToString(p.PublicNetworkAccess) != string(in.PublicNetworkAccess):
		return false
	}
	return true
}
