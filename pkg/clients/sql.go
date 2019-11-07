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
	"reflect"

	"github.com/crossplaneio/crossplane-runtime/pkg/meta"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql/mysqlapi"
	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql/postgresqlapi"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/pkg/errors"

	azuredbv1alpha3 "github.com/crossplaneio/stack-azure/apis/database/v1alpha3"
)

// MySQLServerAPI represents the API interface for a MySQL Server client
type MySQLServerAPI interface {
	ServerNameTaken(ctx context.Context, s *azuredbv1alpha3.MySQLServer) (bool, error)
	GetServer(ctx context.Context, s *azuredbv1alpha3.MySQLServer) (mysql.Server, error)
	CreateServer(ctx context.Context, s *azuredbv1alpha3.MySQLServer, adminPassword string) error
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
func (c *MySQLServerClient) ServerNameTaken(ctx context.Context, s *azuredbv1alpha3.MySQLServer) (bool, error) {
	r, err := c.Execute(ctx, mysql.NameAvailabilityRequest{Name: ToStringPtr(meta.GetExternalName(s))})
	if err != nil {
		return false, err
	}
	return !ToBool(r.NameAvailable), nil
}

// GetServer retrieves the requested MySQL Server
func (c *MySQLServerClient) GetServer(ctx context.Context, s *azuredbv1alpha3.MySQLServer) (mysql.Server, error) {
	return c.ServersClient.Get(ctx, s.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(s))
}

// CreateServer creates a MySQL Server.
func (c *MySQLServerClient) CreateServer(ctx context.Context, s *azuredbv1alpha3.MySQLServer, adminPassword string) error {
	// initialize all the parameters that specify how to configure the server during creation
	properties := &mysql.ServerPropertiesForDefaultCreate{
		AdministratorLogin:         ToStringPtr(s.Spec.ForProvider.AdministratorLogin),
		AdministratorLoginPassword: &adminPassword,
		Version:                    mysql.ServerVersion(ToString(s.Spec.ForProvider.Version)),
		SslEnforcement:             mysql.SslEnforcementEnum(ToString(s.Spec.ForProvider.SslEnforcement)),
		CreateMode:                 mysql.CreateModeDefault,
	}
	if s.Spec.ForProvider.StorageProfile != nil {
		properties.StorageProfile = &mysql.StorageProfile{
			BackupRetentionDays: ToInt32PtrFromIntPtr(s.Spec.ForProvider.StorageProfile.BackupRetentionDays),
			GeoRedundantBackup:  mysql.GeoRedundantBackup(ToString(s.Spec.ForProvider.StorageProfile.GeoRedundantBackup)),
			StorageMB:           ToInt32PtrFromIntPtr(s.Spec.ForProvider.StorageProfile.StorageMB),
			StorageAutogrow:     mysql.StorageAutogrow(ToString(s.Spec.ForProvider.StorageProfile.StorageAutogrow)),
		}
	}
	createParams := mysql.ServerForCreate{
		Sku: &mysql.Sku{
			Name:     ToStringPtr(s.Spec.ForProvider.SKU.Name),
			Tier:     mysql.SkuTier(s.Spec.ForProvider.SKU.Tier),
			Capacity: ToInt32Ptr(s.Spec.ForProvider.SKU.Capacity),
			Family:   ToStringPtr(s.Spec.ForProvider.SKU.Family),
			Size:     ToStringPtr(s.Spec.ForProvider.SKU.Size),
		},
		Properties: properties,
		Location:   &s.Spec.ForProvider.Location,
		Tags:       ToStringPtrMap(s.Spec.ForProvider.Tags),
	}
	_, err := c.Create(ctx, s.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(s), createParams)
	return err
}

// DeleteServer deletes the given MySQLServer resource.
func (c *MySQLServerClient) DeleteServer(ctx context.Context, s *azuredbv1alpha3.MySQLServer) error {
	_, err := c.ServersClient.Delete(ctx, s.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(s))
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
func NewMySQLVirtualNetworkRuleParameters(v *azuredbv1alpha3.MySQLServerVirtualNetworkRule) mysql.VirtualNetworkRule {
	return mysql.VirtualNetworkRule{
		Name: ToStringPtr(v.Spec.Name),
		VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
			VirtualNetworkSubnetID:           ToStringPtr(v.Spec.VirtualNetworkRuleProperties.VirtualNetworkSubnetID),
			IgnoreMissingVnetServiceEndpoint: ToBoolPtr(v.Spec.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint, FieldRequired),
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
	v.Status.ID = ToString(az.ID)
	v.Status.Type = ToString(az.Type)
}

//---------------------------------------------------------------------------------------------------------------------
// PostgreSQLServerClient

// PostgreSQLServerAPI represents the API interface for a MySQL Server client
type PostgreSQLServerAPI interface {
	ServerNameTaken(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) (bool, error)
	GetServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) (postgresql.Server, error)
	CreateServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer, adminPassword string) error
	DeleteServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) error
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
func (c *PostgreSQLServerClient) ServerNameTaken(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) (bool, error) {
	r, err := c.Execute(ctx, postgresql.NameAvailabilityRequest{Name: ToStringPtr(meta.GetExternalName(s))})
	if err != nil {
		return false, err
	}
	return !ToBool(r.NameAvailable), nil
}

// GetServer retrieves the requested PostgreSQL Server
func (c *PostgreSQLServerClient) GetServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) (postgresql.Server, error) {
	return c.ServersClient.Get(ctx, s.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(s))
}

// CreateServer creates a PostgreSQL Server
func (c *PostgreSQLServerClient) CreateServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer, adminPassword string) error {
	// initialize all the parameters that s.Specify how to configure the server during creation
	properties := &postgresql.ServerPropertiesForDefaultCreate{
		AdministratorLogin:         ToStringPtr(s.Spec.ForProvider.AdministratorLogin),
		AdministratorLoginPassword: &adminPassword,
		Version:                    postgresql.ServerVersion(ToString(s.Spec.ForProvider.Version)),
		SslEnforcement:             postgresql.SslEnforcementEnum(ToString(s.Spec.ForProvider.SslEnforcement)),
		CreateMode:                 postgresql.CreateModeDefault,
	}
	if s.Spec.ForProvider.StorageProfile != nil {
		properties.StorageProfile = &postgresql.StorageProfile{
			BackupRetentionDays: ToInt32PtrFromIntPtr(s.Spec.ForProvider.StorageProfile.BackupRetentionDays),
			GeoRedundantBackup:  postgresql.GeoRedundantBackup(ToString(s.Spec.ForProvider.StorageProfile.GeoRedundantBackup)),
			StorageMB:           ToInt32PtrFromIntPtr(s.Spec.ForProvider.StorageProfile.StorageMB),
			StorageAutogrow:     postgresql.StorageAutogrow(ToString(s.Spec.ForProvider.StorageProfile.StorageAutogrow)),
		}
	}
	createParams := postgresql.ServerForCreate{
		Sku: &postgresql.Sku{
			Name:     ToStringPtr(s.Spec.ForProvider.SKU.Name),
			Tier:     postgresql.SkuTier(s.Spec.ForProvider.SKU.Tier),
			Capacity: ToInt32Ptr(s.Spec.ForProvider.SKU.Capacity),
			Family:   ToStringPtr(s.Spec.ForProvider.SKU.Family),
			Size:     ToStringPtr(s.Spec.ForProvider.SKU.Size),
		},
		Properties: properties,
		Location:   &s.Spec.ForProvider.Location,
		Tags:       ToStringPtrMap(s.Spec.ForProvider.Tags),
	}
	_, err := c.Create(ctx, s.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(s), createParams)
	return err
}

// DeleteServer deletes the given PostgreSQL resource
func (c *PostgreSQLServerClient) DeleteServer(ctx context.Context, s *azuredbv1alpha3.PostgreSQLServer) error {
	_, err := c.ServersClient.Delete(ctx, s.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(s))
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
func NewPostgreSQLVirtualNetworkRuleParameters(v *azuredbv1alpha3.PostgreSQLServerVirtualNetworkRule) postgresql.VirtualNetworkRule {
	return postgresql.VirtualNetworkRule{
		Name: ToStringPtr(v.Spec.Name),
		VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
			VirtualNetworkSubnetID:           ToStringPtr(v.Spec.VirtualNetworkRuleProperties.VirtualNetworkSubnetID),
			IgnoreMissingVnetServiceEndpoint: ToBoolPtr(v.Spec.VirtualNetworkRuleProperties.IgnoreMissingVnetServiceEndpoint, FieldRequired),
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
	v.Status.ID = ToString(az.ID)
	v.Status.Type = ToString(az.Type)
}
