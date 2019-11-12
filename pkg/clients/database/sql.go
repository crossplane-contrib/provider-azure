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
	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql/postgresqlapi"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/pkg/errors"

	"github.com/crossplaneio/crossplane-runtime/pkg/meta"

	azuredbv1alpha3 "github.com/crossplaneio/stack-azure/apis/database/v1alpha3"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
)

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
	ServerNameTaken(ctx context.Context, s *azuredbv1alpha3.MySQLServer) (bool, error)
	GetServer(ctx context.Context, s *azuredbv1alpha3.MySQLServer) (mysql.Server, error)
	CreateServer(ctx context.Context, s *azuredbv1alpha3.MySQLServer, adminPassword string) error
	UpdateServer(ctx context.Context, s *azuredbv1alpha3.MySQLServer) error
	DeleteServer(ctx context.Context, s *azuredbv1alpha3.MySQLServer) error
}

//---------------------------------------------------------------------------------------------------------------------
// MySQLServerClient

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
func (c *MySQLServerClient) ServerNameTaken(ctx context.Context, s *azuredbv1alpha3.MySQLServer) (bool, error) {
	r, err := c.Execute(ctx, mysql.NameAvailabilityRequest{Name: azure.ToStringPtr(meta.GetExternalName(s))})
	if err != nil {
		return false, err
	}
	return !azure.ToBool(r.NameAvailable), nil
}

// GetServer retrieves the requested MySQL Server
func (c *MySQLServerClient) GetServer(ctx context.Context, s *azuredbv1alpha3.MySQLServer) (mysql.Server, error) {
	return c.ServersClient.Get(ctx, s.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(s))
}

// CreateServer creates a MySQL Server.
func (c *MySQLServerClient) CreateServer(ctx context.Context, cr *azuredbv1alpha3.MySQLServer, adminPassword string) error {
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
func (c *MySQLServerClient) UpdateServer(ctx context.Context, cr *azuredbv1alpha3.MySQLServer) error {
	// TODO(muvaf): password update via Update call is supported by Azure but
	// we don't support that.
	s := cr.Spec.ForProvider
	properties := &mysql.ServerUpdateParametersProperties{
		Version:        mysql.ServerVersion(s.Version),
		SslEnforcement: mysql.SslEnforcementEnum(s.SSLEnforcement),
		//ReplicationRole: s.Spec.ForProvider.ReplicationRole,
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
func (c *MySQLServerClient) DeleteServer(ctx context.Context, cr *azuredbv1alpha3.MySQLServer) error {
	op, err := c.ServersClient.Delete(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
	if err != nil {
		return err
	}
	_, err = op.DoneWithContext(ctx, c.ServersClient.Client)
	return err
}

//---------------------------------------------------------------------------------------------------------------------
// MySQLVirtualNetworkRulesClient

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

//---------------------------------------------------------------------------------------------------------------------
// PostgreSQLServerClient

// PostgreSQLServerAPI represents the API interface for a MySQL Server client
type PostgreSQLServerAPI interface {
	ServerNameTaken(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) (bool, error)
	GetServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) (postgresql.Server, error)
	CreateServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer, adminPassword string) error
	DeleteServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) error
	UpdateServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) error
}

// PostgreSQLServerClient is the concreate implementation of the SQLServerAPI interface for PostgreSQL that calls Azure API.
type PostgreSQLServerClient struct {
	postgresql.ServersClient
	postgresql.CheckNameAvailabilityClient
}

// NewPostgreSQLServerClient creates and initializes a PostgreSQLServerClient instance.
func NewPostgreSQLServerClient(c *azure.Client) (*PostgreSQLServerClient, error) {
	postgreSQLServerClient := postgresql.NewServersClient(c.SubscriptionID)
	postgreSQLServerClient.Authorizer = c.Authorizer
	postgreSQLServerClient.AddToUserAgent(azure.UserAgent)

	nameClient := postgresql.NewCheckNameAvailabilityClient(c.SubscriptionID)
	nameClient.Authorizer = c.Authorizer
	nameClient.AddToUserAgent(azure.UserAgent)

	return &PostgreSQLServerClient{
		ServersClient:               postgreSQLServerClient,
		CheckNameAvailabilityClient: nameClient,
	}, nil
}

// ServerNameTaken returns true if the supplied server's name has been taken.
func (c *PostgreSQLServerClient) ServerNameTaken(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) (bool, error) {
	r, err := c.Execute(ctx, postgresql.NameAvailabilityRequest{Name: azure.ToStringPtr(meta.GetExternalName(s))})
	if err != nil {
		return false, err
	}
	return !azure.ToBool(r.NameAvailable), nil
}

// GetServer retrieves the requested PostgreSQL Server
func (c *PostgreSQLServerClient) GetServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) (postgresql.Server, error) {
	return c.ServersClient.Get(ctx, s.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(s))
}

// CreateServer creates a PostgreSQL Server
func (c *PostgreSQLServerClient) CreateServer(ctx context.Context, cr *azuredbv1alpha3.PostgreSQLServer, adminPassword string) error {
	s := cr.Spec.ForProvider
	properties := &postgresql.ServerPropertiesForDefaultCreate{
		AdministratorLogin:         azure.ToStringPtr(s.AdministratorLogin),
		AdministratorLoginPassword: &adminPassword,
		Version:                    postgresql.ServerVersion(s.Version),
		SslEnforcement:             postgresql.SslEnforcementEnum(s.SSLEnforcement),
		CreateMode:                 postgresql.CreateModeDefault,
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
	createParams := postgresql.ServerForCreate{
		Sku:        sku,
		Properties: properties,
		Location:   &s.Location,
		Tags:       azure.ToStringPtrMap(s.Tags),
	}
	op, err := c.Create(ctx, s.ResourceGroupName, meta.GetExternalName(cr), createParams)
	if err != nil {
		return err
	}
	_, err = op.DoneWithContext(ctx, c.ServersClient.Client)
	return err
}

// UpdateServer updates a PostgreSQL Server.
func (c *PostgreSQLServerClient) UpdateServer(ctx context.Context, cr *azuredbv1alpha3.PostgreSQLServer) error {
	// TODO(muvaf): password update via Update call is supported by Azure but
	// we don't support that.
	s := cr.Spec.ForProvider
	properties := &postgresql.ServerUpdateParametersProperties{
		Version:        postgresql.ServerVersion(s.Version),
		SslEnforcement: postgresql.SslEnforcementEnum(s.SSLEnforcement),
		//ReplicationRole: s.Spec.ForProvider.ReplicationRole,
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
	_, err = op.DoneWithContext(ctx, c.ServersClient.Client)
	return err
}

// DeleteServer deletes the given PostgreSQL resource
func (c *PostgreSQLServerClient) DeleteServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) error {
	op, err := c.ServersClient.Delete(ctx, s.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(s))
	if err != nil {
		return err
	}
	_, err = op.DoneWithContext(ctx, c.ServersClient.Client)
	return err
}

//---------------------------------------------------------------------------------------------------------------------
// PostgreSQLVirtualNetworkRulesClient

// A PostgreSQLVirtualNetworkRulesClient handles CRUD operations for Azure Virtual Network Rules.
type PostgreSQLVirtualNetworkRulesClient postgresqlapi.VirtualNetworkRulesClientAPI

// NewPostgreSQLVirtualNetworkRulesClient returns a new Azure Virtual Network Rules client. Credentials must be
// passed as JSON encoded data.
func NewPostgreSQLVirtualNetworkRulesClient(ctx context.Context, credentials []byte) (PostgreSQLVirtualNetworkRulesClient, error) {
	c := azure.Credentials{}
	if err := json.Unmarshal(credentials, &c); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal Azure client secret data")
	}

	client := postgresql.NewVirtualNetworkRulesClient(c.SubscriptionID)

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

// NewPostgreSQLVirtualNetworkRuleParameters returns an Azure VirtualNetworkRule object from a virtual network spec
func NewPostgreSQLVirtualNetworkRuleParameters(v *azuredbv1alpha3.PostgreSQLServerVirtualNetworkRule) postgresql.VirtualNetworkRule {
	return postgresql.VirtualNetworkRule{
		Name: azure.ToStringPtr(v.Spec.Name),
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

// Helper functions
// NOTE: These helper functions work for both MySQL and PostreSQL, but we cast everything to the MySQL types because
// the generated Azure clients for MySQL and PostgreSQL are exactly the same content, just a different package. See:
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/mysql/mgmt/2017-12-01/mysql/models.go
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/postgresql/mgmt/2017-12-01/postgresql/models.go

// The name must match the specification of the SKU, so, we don't allow user
// to specify an arbitrary name.

// ToMySQLSKU returns the name of the MySQL Server SKU, which is tier + family + cores, e.g. B_Gen4_1, GP_Gen5_8.
func ToMySQLSKU(skuSpec azuredbv1alpha3.SKU) (*mysql.Sku, error) {
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

// ToPostgreSQLSKU returns the name of the PostgreSQL SKU, which is tier + family + cores, e.g. B_Gen4_1, GP_Gen5_8.
func ToPostgreSQLSKU(skuSpec azuredbv1alpha3.SKU) (*postgresql.Sku, error) {
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

// GeneratePostgreSQLObservation produces SQLServerObservation from postgresql.Server.
func GeneratePostgreSQLObservation(in postgresql.Server) azuredbv1alpha3.SQLServerObservation {
	return azuredbv1alpha3.SQLServerObservation{
		ID:                       azure.ToString(in.ID),
		Name:                     azure.ToString(in.Name),
		Type:                     azure.ToString(in.Type),
		UserVisibleState:         string(in.UserVisibleState),
		FullyQualifiedDomainName: azure.ToString(in.FullyQualifiedDomainName),
		MasterServerID:           azure.ToString(in.MasterServerID),
	}
}

// GenerateMySQLObservation produces SQLServerObservation from mysql.Server
func GenerateMySQLObservation(in mysql.Server) azuredbv1alpha3.SQLServerObservation {
	return azuredbv1alpha3.SQLServerObservation{
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
func LateInitializeMySQL(p *azuredbv1alpha3.SQLServerParameters, in mysql.Server) {
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
// nolint:gocyclo
func IsMySQLUpToDate(p azuredbv1alpha3.SQLServerParameters, in mysql.Server) bool {
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

// LateInitializePostgreSQL fills the empty values of SQLServerParameters with the
// ones that are retrieved from the Azure API.
func LateInitializePostgreSQL(p *azuredbv1alpha3.SQLServerParameters, in postgresql.Server) {
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

// IsPostgreSQLUpToDate is used to report whether given postgresql.Server is in
// sync with the SQLServerParameters that user desires.
// nolint:gocyclo
func IsPostgreSQLUpToDate(p azuredbv1alpha3.SQLServerParameters, in postgresql.Server) bool {
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
