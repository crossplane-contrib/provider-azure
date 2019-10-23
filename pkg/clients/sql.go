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

package azure

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

	azuredbv1alpha2 "github.com/crossplaneio/stack-azure/apis/database/v1alpha2"
)

const (
	backupRetentionDaysDefault = int32(7)
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
	ServerNameTaken(ctx context.Context, s *azuredbv1alpha2.MysqlServer) (bool, error)
	GetServer(ctx context.Context, s *azuredbv1alpha2.MysqlServer) (mysql.Server, error)
	CreateServer(ctx context.Context, s *azuredbv1alpha2.MysqlServer, adminPassword string) error
	DeleteServer(ctx context.Context, s *azuredbv1alpha2.MysqlServer) error
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
func NewMySQLServerClient(c *Client) (*MySQLServerClient, error) {
	mysqlServersClient := mysql.NewServersClient(c.SubscriptionID)
	mysqlServersClient.Authorizer = c.Authorizer
	mysqlServersClient.AddToUserAgent(UserAgent)

	nameClient := mysql.NewCheckNameAvailabilityClient(c.SubscriptionID)
	nameClient.Authorizer = c.Authorizer
	nameClient.AddToUserAgent(UserAgent)

	return &MySQLServerClient{
		ServersClient:               mysqlServersClient,
		CheckNameAvailabilityClient: nameClient,
	}, nil
}

// ServerNameTaken returns true if the supplied server's name has been taken.
func (c *MySQLServerClient) ServerNameTaken(ctx context.Context, s *azuredbv1alpha2.MysqlServer) (bool, error) {
	r, err := c.Execute(ctx, mysql.NameAvailabilityRequest{Name: ToStringPtr(s.GetName())})
	if err != nil {
		return false, err
	}
	return !ToBool(r.NameAvailable), nil
}

// GetServer retrieves the requested MySQL Server
func (c *MySQLServerClient) GetServer(ctx context.Context, s *azuredbv1alpha2.MysqlServer) (mysql.Server, error) {
	return c.ServersClient.Get(ctx, s.Spec.ResourceGroupName, s.GetName())
}

// CreateServer creates a MySQL Server.
func (c *MySQLServerClient) CreateServer(ctx context.Context, s *azuredbv1alpha2.MysqlServer, adminPassword string) error {
	// initialize all the parameters that specify how to configure the server during creation
	skuName, err := SQLServerSkuName(s.Spec.PricingTier)
	if err != nil {
		return errors.Wrap(err, "failed to create server SKU name")
	}
	capacity := int32(s.Spec.PricingTier.VCores)
	storageMB := int32(s.Spec.StorageProfile.StorageGB * 1024)
	backupRetentionDays := backupRetentionDaysDefault
	if s.Spec.StorageProfile.BackupRetentionDays > 0 {
		backupRetentionDays = int32(s.Spec.StorageProfile.BackupRetentionDays)
	}
	createParams := mysql.ServerForCreate{
		Sku: &mysql.Sku{
			Name:     &skuName,
			Tier:     mysql.SkuTier(s.Spec.PricingTier.Tier),
			Capacity: &capacity,
			Family:   &s.Spec.PricingTier.Family,
		},
		Properties: &mysql.ServerPropertiesForDefaultCreate{
			AdministratorLogin:         &s.Spec.AdminLoginName,
			AdministratorLoginPassword: &adminPassword,
			Version:                    mysql.ServerVersion(s.Spec.Version),
			SslEnforcement:             ToSslEnforcement(s.Spec.SSLEnforced),
			StorageProfile: &mysql.StorageProfile{
				BackupRetentionDays: &backupRetentionDays,
				GeoRedundantBackup:  ToGeoRedundantBackup(s.Spec.StorageProfile.GeoRedundantBackup),
				StorageMB:           &storageMB,
			},
			CreateMode: mysql.CreateModeDefault,
		},
		Location: &s.Spec.Location,
	}

	_, err = c.Create(ctx, s.Spec.ResourceGroupName, s.GetName(), createParams)
	return err
}

// DeleteServer deletes the given MySQLServer resource.
func (c *MySQLServerClient) DeleteServer(ctx context.Context, s *azuredbv1alpha2.MysqlServer) error {
	_, err := c.ServersClient.Delete(ctx, s.Spec.ResourceGroupName, s.GetName())
	return err
}

//---------------------------------------------------------------------------------------------------------------------
// MySQLVirtualNetworkRulesClient

// A MySQLVirtualNetworkRulesClient handles CRUD operations for Azure Virtual Network Rules.
type MySQLVirtualNetworkRulesClient mysqlapi.VirtualNetworkRulesClientAPI

// NewMySQLVirtualNetworkRulesClient returns a new Azure Virtual Network Rules client. Credentials must be
// passed as JSON encoded data.
func NewMySQLVirtualNetworkRulesClient(ctx context.Context, credentials []byte) (MySQLVirtualNetworkRulesClient, error) {
	c := Credentials{}
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
	if err := client.AddToUserAgent(UserAgent); err != nil {
		return nil, errors.Wrap(err, "cannot add to Azure client user agent")
	}

	return client, nil
}

// NewMySQLVirtualNetworkRuleParameters returns an Azure VirtualNetworkRule object from a virtual network spec
func NewMySQLVirtualNetworkRuleParameters(v *azuredbv1alpha2.MysqlServerVirtualNetworkRule) mysql.VirtualNetworkRule {
	return mysql.VirtualNetworkRule{
		Name: ToStringPtr(v.Spec.Name),
		VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
			VirtualNetworkSubnetID:           ToStringPtr(v.Spec.VirtualNetworkRuleProperties.VirtualNetworkSubnetID),
			IgnoreMissingVnetServiceEndpoint: ToBoolPtr(v.Spec.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint, FieldRequired),
		},
	}
}

// MySQLServerVirtualNetworkRuleNeedsUpdate determines if a virtual network rule needs to be updated
func MySQLServerVirtualNetworkRuleNeedsUpdate(kube *azuredbv1alpha2.MysqlServerVirtualNetworkRule, az mysql.VirtualNetworkRule) bool {
	up := NewMySQLVirtualNetworkRuleParameters(kube)

	switch {
	case !reflect.DeepEqual(up.VirtualNetworkRuleProperties.VirtualNetworkSubnetID, az.VirtualNetworkRuleProperties.VirtualNetworkSubnetID):
		return true
	case !reflect.DeepEqual(up.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint, az.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint):
		return true
	}

	return false
}

// MySQLVirtualNetworkRuleStatusFromAzure converts an Azure subnet to a SubnetStatus
func MySQLVirtualNetworkRuleStatusFromAzure(az mysql.VirtualNetworkRule) azuredbv1alpha2.VirtualNetworkRuleStatus {
	return azuredbv1alpha2.VirtualNetworkRuleStatus{
		State: string(az.VirtualNetworkRuleProperties.State),
		ID:    ToString(az.ID),
		Type:  ToString(az.Type),
	}
}

//---------------------------------------------------------------------------------------------------------------------
// PostgreSQLServerClient

// PostgreSQLServerAPI represents the API interface for a MySQL Server client
type PostgreSQLServerAPI interface {
	ServerNameTaken(ctx context.Context, s *azuredbv1alpha2.PostgresqlServer) (bool, error)
	GetServer(ctx context.Context, s *azuredbv1alpha2.PostgresqlServer) (postgresql.Server, error)
	CreateServer(ctx context.Context, s *azuredbv1alpha2.PostgresqlServer, adminPassword string) error
	DeleteServer(ctx context.Context, s *azuredbv1alpha2.PostgresqlServer) error
}

// PostgreSQLServerClient is the concreate implementation of the SQLServerAPI interface for PostgreSQL that calls Azure API.
type PostgreSQLServerClient struct {
	postgresql.ServersClient
	postgresql.CheckNameAvailabilityClient
}

// NewPostgreSQLServerClient creates and initializes a PostgreSQLServerClient instance.
func NewPostgreSQLServerClient(c *Client) (*PostgreSQLServerClient, error) {
	postgreSQLServerClient := postgresql.NewServersClient(c.SubscriptionID)
	postgreSQLServerClient.Authorizer = c.Authorizer
	postgreSQLServerClient.AddToUserAgent(UserAgent)

	nameClient := postgresql.NewCheckNameAvailabilityClient(c.SubscriptionID)
	nameClient.Authorizer = c.Authorizer
	nameClient.AddToUserAgent(UserAgent)

	return &PostgreSQLServerClient{
		ServersClient:               postgreSQLServerClient,
		CheckNameAvailabilityClient: nameClient,
	}, nil
}

// ServerNameTaken returns true if the supplied server's name has been taken.
func (c *PostgreSQLServerClient) ServerNameTaken(ctx context.Context, s *azuredbv1alpha2.PostgresqlServer) (bool, error) {
	r, err := c.Execute(ctx, postgresql.NameAvailabilityRequest{Name: ToStringPtr(s.GetName())})
	if err != nil {
		return false, err
	}
	return !ToBool(r.NameAvailable), nil
}

// GetServer retrieves the requested PostgreSQL Server
func (c *PostgreSQLServerClient) GetServer(ctx context.Context, s *azuredbv1alpha2.PostgresqlServer) (postgresql.Server, error) {
	return c.ServersClient.Get(ctx, s.Spec.ResourceGroupName, s.GetName())
}

// CreateServer creates a PostgreSQL Server
func (c *PostgreSQLServerClient) CreateServer(ctx context.Context, s *azuredbv1alpha2.PostgresqlServer, adminPassword string) error {
	// initialize all the parameters that s.Specify how to configure the server during creation
	skuName, err := SQLServerSkuName(s.Spec.PricingTier)
	if err != nil {
		return errors.Wrap(err, "failed to create server SKU name")
	}
	capacity := int32(s.Spec.PricingTier.VCores)
	storageMB := int32(s.Spec.StorageProfile.StorageGB * 1024)
	backupRetentionDays := backupRetentionDaysDefault
	if s.Spec.StorageProfile.BackupRetentionDays > 0 {
		backupRetentionDays = int32(s.Spec.StorageProfile.BackupRetentionDays)
	}
	createParams := postgresql.ServerForCreate{
		Sku: &postgresql.Sku{
			Name:     &skuName,
			Tier:     postgresql.SkuTier(s.Spec.PricingTier.Tier),
			Capacity: &capacity,
			Family:   &s.Spec.PricingTier.Family,
		},
		Properties: &postgresql.ServerPropertiesForDefaultCreate{
			AdministratorLogin:         &s.Spec.AdminLoginName,
			AdministratorLoginPassword: &adminPassword,
			Version:                    postgresql.ServerVersion(s.Spec.Version),
			SslEnforcement:             postgresql.SslEnforcementEnum(ToSslEnforcement(s.Spec.SSLEnforced)),
			StorageProfile: &postgresql.StorageProfile{
				BackupRetentionDays: &backupRetentionDays,
				GeoRedundantBackup:  postgresql.GeoRedundantBackup(ToGeoRedundantBackup(s.Spec.StorageProfile.GeoRedundantBackup)),
				StorageMB:           &storageMB,
			},
			CreateMode: postgresql.CreateModeDefault,
		},
		Location: &s.Spec.Location,
	}

	_, err = c.Create(ctx, s.Spec.ResourceGroupName, s.GetName(), createParams)
	return err
}

// DeleteServer deletes the given PostgreSQL resource
func (c *PostgreSQLServerClient) DeleteServer(ctx context.Context, s *azuredbv1alpha2.PostgresqlServer) error {
	_, err := c.ServersClient.Delete(ctx, s.Spec.ResourceGroupName, s.GetName())
	return err
}

//---------------------------------------------------------------------------------------------------------------------
// PostgreSQLVirtualNetworkRulesClient

// A PostgreSQLVirtualNetworkRulesClient handles CRUD operations for Azure Virtual Network Rules.
type PostgreSQLVirtualNetworkRulesClient postgresqlapi.VirtualNetworkRulesClientAPI

// NewPostgreSQLVirtualNetworkRulesClient returns a new Azure Virtual Network Rules client. Credentials must be
// passed as JSON encoded data.
func NewPostgreSQLVirtualNetworkRulesClient(ctx context.Context, credentials []byte) (PostgreSQLVirtualNetworkRulesClient, error) {
	c := Credentials{}
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
	if err := client.AddToUserAgent(UserAgent); err != nil {
		return nil, errors.Wrap(err, "cannot add to Azure client user agent")
	}

	return client, nil
}

// NewPostgreSQLVirtualNetworkRuleParameters returns an Azure VirtualNetworkRule object from a virtual network spec
func NewPostgreSQLVirtualNetworkRuleParameters(v *azuredbv1alpha2.PostgresqlServerVirtualNetworkRule) postgresql.VirtualNetworkRule {
	return postgresql.VirtualNetworkRule{
		Name: ToStringPtr(v.Spec.Name),
		VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
			VirtualNetworkSubnetID:           ToStringPtr(v.Spec.VirtualNetworkRuleProperties.VirtualNetworkSubnetID),
			IgnoreMissingVnetServiceEndpoint: ToBoolPtr(v.Spec.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint, FieldRequired),
		},
	}
}

// PostgreSQLServerVirtualNetworkRuleNeedsUpdate determines if a virtual network rule needs to be updated
func PostgreSQLServerVirtualNetworkRuleNeedsUpdate(kube *azuredbv1alpha2.PostgresqlServerVirtualNetworkRule, az postgresql.VirtualNetworkRule) bool {
	up := NewPostgreSQLVirtualNetworkRuleParameters(kube)

	switch {
	case !reflect.DeepEqual(up.VirtualNetworkRuleProperties.VirtualNetworkSubnetID, az.VirtualNetworkRuleProperties.VirtualNetworkSubnetID):
		return true
	case !reflect.DeepEqual(up.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint, az.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint):
		return true
	}

	return false
}

// PostgreSQLVirtualNetworkRuleStatusFromAzure converts an Azure subnet to a SubnetStatus
func PostgreSQLVirtualNetworkRuleStatusFromAzure(az postgresql.VirtualNetworkRule) azuredbv1alpha2.VirtualNetworkRuleStatus {
	return azuredbv1alpha2.VirtualNetworkRuleStatus{
		State: string(az.State),
		ID:    ToString(az.ID),
		Type:  ToString(az.Type),
	}
}

// Helper functions
// NOTE: These helper functions work for both MySQL and PostreSQL, but we cast everything to the MySQL types because
// the generated Azure clients for MySQL and PostgreSQL are exactly the same content, just a different package. See:
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/mysql/mgmt/2017-12-01/mysql/models.go
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/postgresql/mgmt/2017-12-01/postgresql/models.go

// SQLServerSkuName returns the name of the MySQL Server SKU, which is tier + family + cores, e.g. B_Gen4_1, GP_Gen5_8.
func SQLServerSkuName(pricingTier azuredbv1alpha2.PricingTierSpec) (string, error) {
	t, ok := skuShortTiers[mysql.SkuTier(pricingTier.Tier)]
	if !ok {
		return "", fmt.Errorf("tier '%s' is not one of the supported values: %+v", pricingTier.Tier, mysql.PossibleSkuTierValues())
	}

	return fmt.Sprintf("%s_%s_%s", t, pricingTier.Family, strconv.Itoa(pricingTier.VCores)), nil
}

// ToSslEnforcement converts the given bool its corresponding SslEnforcementEnum value
func ToSslEnforcement(sslEnforced bool) mysql.SslEnforcementEnum {
	if sslEnforced {
		return mysql.SslEnforcementEnumEnabled
	}
	return mysql.SslEnforcementEnumDisabled
}

// ToGeoRedundantBackup converts the given bool its corresponding GeoRedundantBackup value
func ToGeoRedundantBackup(geoRedundantBackup bool) mysql.GeoRedundantBackup {
	if geoRedundantBackup {
		return mysql.Enabled
	}
	return mysql.Disabled
}
