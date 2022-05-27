/*
Copyright 2021 The Crossplane Authors.

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

package configuration

import (
	"context"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/go-autorest/autorest"

	azuredbv1beta1 "github.com/crossplane-contrib/provider-azure/apis/database/v1beta1"
	"github.com/crossplane-contrib/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
)

// NOTE: postgresql and mysql structs and functions live in their respective
// packages even though they are exactly the same. However, Crossplane does not
// make that assumption and use the respective package for each type, although,
// they both share the same SQLServerParameters and SQLServerObservation objects.
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/mysql/mgmt/2017-12-01/mysql/models.go
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/postgresql/mgmt/2017-12-01/postgresql/models.go

// MySQLConfigurationAPI represents the API interface for a MySQL Server Configuration client
type MySQLConfigurationAPI interface {
	Get(ctx context.Context, s *azuredbv1beta1.MySQLServerConfiguration) (mysql.Configuration, error)
	CreateOrUpdate(ctx context.Context, s *azuredbv1beta1.MySQLServerConfiguration) error
	Delete(ctx context.Context, s *azuredbv1beta1.MySQLServerConfiguration) error
	GetRESTClient() autorest.Sender
}

// MySQLConfigurationClient is the concreate implementation of the MySQLConfigurationAPI interface for MySQL that calls Azure API.
type MySQLConfigurationClient struct {
	mysql.ConfigurationsClient
}

// NewMySQLConfigurationClient creates and initializes a MySQLConfigurationClient instance.
func NewMySQLConfigurationClient(cl mysql.ConfigurationsClient) *MySQLConfigurationClient {
	return &MySQLConfigurationClient{
		ConfigurationsClient: cl,
	}
}

// GetRESTClient returns the underlying REST client that the client object uses.
func (c *MySQLConfigurationClient) GetRESTClient() autorest.Sender {
	return c.ConfigurationsClient.Client
}

// Get retrieves the requested MySQL Configuration
func (c *MySQLConfigurationClient) Get(ctx context.Context, cr *azuredbv1beta1.MySQLServerConfiguration) (mysql.Configuration, error) {
	return c.ConfigurationsClient.Get(ctx, cr.Spec.ForProvider.ResourceGroupName, cr.Spec.ForProvider.ServerName, cr.Spec.ForProvider.Name)
}

// CreateOrUpdate creates or updates a MySQL Server Configuration
func (c *MySQLConfigurationClient) CreateOrUpdate(ctx context.Context, cr *azuredbv1beta1.MySQLServerConfiguration) error {
	return c.update(ctx, cr, cr.Spec.ForProvider.Value, nil)
}

// Delete deletes the given MySQL Server Configuration
func (c *MySQLConfigurationClient) Delete(ctx context.Context, cr *azuredbv1beta1.MySQLServerConfiguration) error {
	source := SourceSystemManaged
	// we are mimicking Terraform behavior here: when the configuration object
	// is deleted, we are resetting its value to the system default,
	// and updating its source to "system-default" to declare that
	// we are no longer managing it.
	return c.update(ctx, cr, &cr.Status.AtProvider.DefaultValue, &source)
}

func (c *MySQLConfigurationClient) update(ctx context.Context, cr *azuredbv1beta1.MySQLServerConfiguration, value, source *string) error {
	s := cr.Spec.ForProvider
	config := mysql.Configuration{
		ConfigurationProperties: &mysql.ConfigurationProperties{
			Value:  value,
			Source: source,
		},
	}
	op, err := c.ConfigurationsClient.CreateOrUpdate(ctx, s.ResourceGroupName, cr.Spec.ForProvider.ServerName, cr.Spec.ForProvider.Name, config)
	if err != nil {
		return err
	}
	cr.Status.AtProvider.LastOperation = v1alpha3.AsyncOperation{
		PollingURL: op.PollingURL(),
		Method:     http.MethodPut,
	}
	return nil
}

// UpdateMySQLConfigurationObservation produces SQLServerConfigurationObservation from mysql.Configuration.
func UpdateMySQLConfigurationObservation(o *azuredbv1beta1.SQLServerConfigurationObservation, in mysql.Configuration) {
	o.ID = azure.ToString(in.ID)
	o.Name = azure.ToString(in.Name)
	o.Type = azure.ToString(in.Type)
	o.DataType = azure.ToString(in.DataType)
	o.Value = azure.ToString(in.Value)
	o.DefaultValue = azure.ToString(in.DefaultValue)
	o.Source = azure.ToString(in.Source)
	o.Description = azure.ToString(in.Description)
}

// IsMySQLConfigurationUpToDate is used to report whether given mysql.Configuration is in
// sync with the SQLServerConfigurationParameters that user desires.
func IsMySQLConfigurationUpToDate(p azuredbv1beta1.SQLServerConfigurationParameters, in mysql.Configuration) bool {
	return azure.ToString(p.Value) == azure.ToString(in.Value)
}
