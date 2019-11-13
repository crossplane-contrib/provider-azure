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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql/mysqlapi"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/pkg/errors"

	"github.com/crossplaneio/crossplane-runtime/pkg/meta"

	azuredbv1alpha3 "github.com/crossplaneio/stack-azure/apis/database/v1alpha3"
	azuredbv1beta1 "github.com/crossplaneio/stack-azure/apis/database/v1beta1"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
)

// NOTE: postgresql and mysql structs and functions live in their respective
// packages even though they are exactly the same. However, Crossplane does not
// make that assumption and use the respective package for each type, although,
// they both share the same SQLServerParameters and SQLServerObservation objects.
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/mysql/mgmt/2017-12-01/mysql/models.go
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/postgresql/mgmt/2017-12-01/postgresql/models.go

// State strings for MySQL and PostgreSQL.
const (
	StateDisabled = string(mysql.ServerStateDisabled)
	StateDropping = string(mysql.ServerStateDropping)
	StateReady    = string(mysql.ServerStateReady)
)

var (
	skuShortTiers = map[mysql.SkuTier]string{
		mysql.Basic:           "B",
		mysql.GeneralPurpose:  "GP",
		mysql.MemoryOptimized: "MO",
	}
)

// MySQLServerAPI represents the API interface for a MySQL Server client
type MySQLServerAPI interface {
	ServerNameTaken(ctx context.Context, s *azuredbv1beta1.MySQLServer) (bool, error)
	GetServer(ctx context.Context, s *azuredbv1beta1.MySQLServer) (mysql.Server, error)
	CreateServer(ctx context.Context, s *azuredbv1beta1.MySQLServer, adminPassword string) error
	UpdateServer(ctx context.Context, s *azuredbv1beta1.MySQLServer) error
	DeleteServer(ctx context.Context, s *azuredbv1beta1.MySQLServer) error
}

// MySQLServerClient is the concrete implementation of the MySQLServerAPI
// interface for MySQL that calls Azure API.
type MySQLServerClient struct {
	mysql.ServersClient
	mysql.CheckNameAvailabilityClient
}

// NewMySQLServerClient creates and initializes a MySQLServerClient instance.
func NewMySQLServerClient(c *azure.Client) (*MySQLServerClient, error) {
	mysqlServersClient := mysql.NewServersClient(c.SubscriptionID)
	mysqlServersClient.Authorizer = c.Authorizer
	mysqlServersClient.AddToUserAgent(azure.UserAgent)

	nameClient := mysql.NewCheckNameAvailabilityClient(c.SubscriptionID)
	nameClient.Authorizer = c.Authorizer
	nameClient.AddToUserAgent(azure.UserAgent)

	return &MySQLServerClient{
		ServersClient:               mysqlServersClient,
		CheckNameAvailabilityClient: nameClient,
	}, nil
}

// ServerNameTaken returns true if the supplied server's name has been taken.
func (c *MySQLServerClient) ServerNameTaken(ctx context.Context, s *azuredbv1beta1.MySQLServer) (bool, error) {
	r, err := c.Execute(ctx, mysql.NameAvailabilityRequest{Name: azure.ToStringPtr(meta.GetExternalName(s))})
	if err != nil {
		return false, err
	}
	return !azure.ToBool(r.NameAvailable), nil
}

// GetServer retrieves the requested MySQL Server
func (c *MySQLServerClient) GetServer(ctx context.Context, s *azuredbv1beta1.MySQLServer) (mysql.Server, error) {
	return c.ServersClient.Get(ctx, s.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(s))
}

// CreateServer creates a MySQL Server.
func (c *MySQLServerClient) CreateServer(ctx context.Context, cr *azuredbv1beta1.MySQLServer, adminPassword string) error {
	s := cr.Spec.ForProvider
	properties := &mysql.ServerPropertiesForDefaultCreate{
		AdministratorLogin:         azure.ToStringPtr(s.AdministratorLogin),
		AdministratorLoginPassword: &adminPassword,
		Version:                    mysql.ServerVersion(s.Version),
		SslEnforcement:             mysql.SslEnforcementEnum(s.SSLEnforcement),
		CreateMode:                 mysql.CreateModeDefault,
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
	createParams := mysql.ServerForCreate{
		Sku:        sku,
		Properties: properties,
		Location:   &s.Location,
		Tags:       azure.ToStringPtrMap(s.Tags),
	}
	op, err := c.Create(ctx, s.ResourceGroupName, meta.GetExternalName(cr), createParams)
	if err != nil {
		return err
	}
	// NOTE(muvaf): There are cases where Create call does not return error
	// and you don't see any SQL instance in the console UI. In those cases,
	// since name is not taken, we always call Create with a new password each time.
	// The problem is, error that blocks creation happens after the Create call
	// is initiated.
	// DoneWithContext checks the operation once and errors usually surface
	// themselves in the first check, if there is one we are able to warn the user
	// for a fix.
	// However, there could be cases where the error appears after a long while.
	// For that reason, some libraries make use of WithCompletionRef function. However,
	// that function blocks until success or error with the context deadline.
	// Azure does not provide any guarantees regarding provisioning time and it
	// can get really long. In most of the cases, our context deadline gets reached.
	//
	// For the time being, we check only once to cover the greater possibility of
	// capturing the error. If there is none, we'll never call Create again anyway.
	// If the problem happens after a while, we'd fallback to old behavior where
	// we continuously try to create and fail.
	_, err = op.DoneWithContext(ctx, c.ServersClient.Client)
	return err
}

// UpdateServer updates a MySQL Server.
func (c *MySQLServerClient) UpdateServer(ctx context.Context, cr *azuredbv1beta1.MySQLServer) error {
	// TODO(muvaf): password update via Update call is supported by Azure but
	// we don't support that.
	s := cr.Spec.ForProvider
	properties := &mysql.ServerUpdateParametersProperties{
		Version:        mysql.ServerVersion(s.Version),
		SslEnforcement: mysql.SslEnforcementEnum(s.SSLEnforcement),
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
	_, err = op.DoneWithContext(ctx, c.ServersClient.Client)
	return err
}

// DeleteServer deletes the given MySQLServer resource.
func (c *MySQLServerClient) DeleteServer(ctx context.Context, cr *azuredbv1beta1.MySQLServer) error {
	op, err := c.ServersClient.Delete(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
	if err != nil {
		return err
	}
	_, err = op.DoneWithContext(ctx, c.ServersClient.Client)
	return err
}

// A MySQLVirtualNetworkRulesClient handles CRUD operations for Azure Virtual Network Rules.
type MySQLVirtualNetworkRulesClient mysqlapi.VirtualNetworkRulesClientAPI

// NewMySQLVirtualNetworkRulesClient returns a new Azure Virtual Network Rules client. Credentials must be
// passed as JSON encoded data.
func NewMySQLVirtualNetworkRulesClient(ctx context.Context, credentials []byte) (MySQLVirtualNetworkRulesClient, error) {
	c := azure.Credentials{}
	if err := json.Unmarshal(credentials, &c); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal Azure client secret data")
	}

	client := mysql.NewVirtualNetworkRulesClient(c.SubscriptionID)

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

// NewMySQLVirtualNetworkRuleParameters returns an Azure VirtualNetworkRule object from a virtual network spec
func NewMySQLVirtualNetworkRuleParameters(v *azuredbv1alpha3.MySQLServerVirtualNetworkRule) mysql.VirtualNetworkRule {
	return mysql.VirtualNetworkRule{
		Name: azure.ToStringPtr(v.Spec.Name),
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

// GenerateMySQLObservation produces SQLServerObservation from mysql.Server
func GenerateMySQLObservation(in mysql.Server) azuredbv1beta1.SQLServerObservation {
	return azuredbv1beta1.SQLServerObservation{
		ID:                       azure.ToString(in.ID),
		Name:                     azure.ToString(in.Name),
		Type:                     azure.ToString(in.Type),
		UserVisibleState:         string(in.UserVisibleState),
		FullyQualifiedDomainName: azure.ToString(in.FullyQualifiedDomainName),
		MasterServerID:           azure.ToString(in.MasterServerID),
	}
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
