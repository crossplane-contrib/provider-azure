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
	"net/http"
	"reflect"
	"strconv"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql/mysqlapi"
	"github.com/Azure/go-autorest/autorest"
	azureautorest "github.com/Azure/go-autorest/autorest/azure"
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

	AsyncOperationStatusInProgress = "InProgress"

	asyncOperationPollingMethod = "AsyncOperation"
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
	GetRESTClient() autorest.Sender
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
	if err := mysqlServersClient.AddToUserAgent(azure.UserAgent); err != nil {
		return nil, err
	}

	nameClient := mysql.NewCheckNameAvailabilityClient(c.SubscriptionID)
	nameClient.Authorizer = c.Authorizer
	if err := nameClient.AddToUserAgent(azure.UserAgent); err != nil {
		return nil, err
	}

	return &MySQLServerClient{
		ServersClient:               mysqlServersClient,
		CheckNameAvailabilityClient: nameClient,
	}, nil
}

// GetRESTClient returns the underlying REST client that the client object uses.
func (c *MySQLServerClient) GetRESTClient() autorest.Sender {
	return c.ServersClient.Client
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
func (c *MySQLServerClient) GetServer(ctx context.Context, cr *azuredbv1beta1.MySQLServer) (mysql.Server, error) {
	return c.ServersClient.Get(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
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
	cr.Status.AtProvider.LastOperation = azuredbv1beta1.AsyncOperation{
		PollingURL: op.PollingURL(),
		Method:     http.MethodPut,
	}
	return FetchAsyncOperation(ctx, c.ServersClient.Client, &cr.Status.AtProvider.LastOperation)
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
	cr.Status.AtProvider.LastOperation = azuredbv1beta1.AsyncOperation{
		PollingURL: op.PollingURL(),
		Method:     http.MethodPatch,
	}
	return FetchAsyncOperation(ctx, c.ServersClient.Client, &cr.Status.AtProvider.LastOperation)
}

// DeleteServer deletes the given MySQLServer resource.
func (c *MySQLServerClient) DeleteServer(ctx context.Context, cr *azuredbv1beta1.MySQLServer) error {
	op, err := c.ServersClient.Delete(ctx, cr.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(cr))
	if err != nil {
		return err
	}
	cr.Status.AtProvider.LastOperation = azuredbv1beta1.AsyncOperation{
		PollingURL: op.PollingURL(),
		Method:     http.MethodDelete,
	}
	return FetchAsyncOperation(ctx, c.ServersClient.Client, &cr.Status.AtProvider.LastOperation)
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

// TODO(muvaf): FetchAsyncOperation can be used by other managed resources as well.

// FetchAsyncOperation updates the given operation object with the most up-to-date
// status retrieved from Azure API.
func FetchAsyncOperation(ctx context.Context, client autorest.Sender, as *azuredbv1beta1.AsyncOperation) error {
	if as == nil || as.PollingURL == "" {
		return nil
	}
	// NOTE(muvaf):There is NewFutureFromResponse method to construct Future
	// object but that requires http.Request object. Even though we construct a
	// fake http.Request object, the poll operation makes decisions based on the
	// response status code and request headers. JSON marshal needs less
	// information and it's safer to cover all types of pollingTrackedBase objects.
	futureJSON, err := json.Marshal(map[string]string{
		"method":        as.Method,
		"pollingMethod": asyncOperationPollingMethod,
		"pollingURI":    as.PollingURL,
	})
	if err != nil {
		return err
	}
	op := &azureautorest.Future{}
	if err := op.UnmarshalJSON(futureJSON); err != nil {
		return err
	}
	// NOTE(muvaf): This function is meant to fetch the operation status, meaning
	// it shouldn't fail if the operation reports error. It should fail if an
	// error appears during the HTTP calls that are made to fetch operation
	// status. But DoneWithContext returns uses the same error variable for both
	// cases, so, we make a compromise and not return the error even if it's
	// related to fetch call.
	_, err = op.DoneWithContext(ctx, client)
	as.Status = op.Status()
	if err != nil {
		as.ErrorMessage = err.Error()
	}
	return nil
}
