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

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/go-autorest/autorest"

	azuredbv1beta1 "github.com/crossplane/provider-azure/apis/database/v1beta1"
	"github.com/crossplane/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

const (
	// SourceSystemManaged represents the source for system-managed configuration values
	SourceSystemManaged = "system-default"
)

// NOTE: postgresql and mysql structs and functions live in their respective
// packages even though they are exactly the same. However, Crossplane does not
// make that assumption and use the respective package for each type, although,
// they both share the same SQLServerParameters and SQLServerObservation objects.
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/mysql/mgmt/2017-12-01/mysql/models.go
// https://github.com/Azure/azure-sdk-for-go/blob/master/services/postgresql/mgmt/2017-12-01/postgresql/models.go

// PostgreSQLConfigurationAPI represents the API interface for a PostgreSQL Server Configuration client
type PostgreSQLConfigurationAPI interface {
	Get(ctx context.Context, s *azuredbv1beta1.PostgreSQLServerConfiguration) (postgresql.Configuration, error)
	CreateOrUpdate(ctx context.Context, s *azuredbv1beta1.PostgreSQLServerConfiguration) error
	Delete(ctx context.Context, s *azuredbv1beta1.PostgreSQLServerConfiguration) error
	GetRESTClient() autorest.Sender
}

// PostgreSQLConfigurationClient is the concreate implementation of the PostgreSQLConfigurationAPI interface for PostgreSQL that calls Azure API.
type PostgreSQLConfigurationClient struct {
	postgresql.ConfigurationsClient
}

// NewPostgreSQLConfigurationClient creates and initializes a PostgreSQLConfigurationClient instance.
func NewPostgreSQLConfigurationClient(cl postgresql.ConfigurationsClient) *PostgreSQLConfigurationClient {
	return &PostgreSQLConfigurationClient{
		ConfigurationsClient: cl,
	}
}

// GetRESTClient returns the underlying REST client that the client object uses.
func (c *PostgreSQLConfigurationClient) GetRESTClient() autorest.Sender {
	return c.ConfigurationsClient.Client
}

// Get retrieves the requested PostgreSQL Configuration
func (c *PostgreSQLConfigurationClient) Get(ctx context.Context, cr *azuredbv1beta1.PostgreSQLServerConfiguration) (postgresql.Configuration, error) {
	return c.ConfigurationsClient.Get(ctx, cr.Spec.ForProvider.ResourceGroupName, cr.Spec.ForProvider.ServerName, cr.Spec.ForProvider.Name)
}

// CreateOrUpdate creates or updates a PostgreSQL Server Configuration
func (c *PostgreSQLConfigurationClient) CreateOrUpdate(ctx context.Context, cr *azuredbv1beta1.PostgreSQLServerConfiguration) error {
	return c.update(ctx, cr, cr.Spec.ForProvider.Value, nil)
}

// Delete deletes the given PostgreSQL Server Configuration
func (c *PostgreSQLConfigurationClient) Delete(ctx context.Context, cr *azuredbv1beta1.PostgreSQLServerConfiguration) error {
	source := SourceSystemManaged
	// we are mimicking Terraform behavior here: when the configuration object
	// is deleted, we are resetting its value to the system default,
	// and updating its source to "system-default" to declare that
	// we are no longer managing it.
	return c.update(ctx, cr, &cr.Status.AtProvider.DefaultValue, &source)
}

func (c *PostgreSQLConfigurationClient) update(ctx context.Context, cr *azuredbv1beta1.PostgreSQLServerConfiguration, value, source *string) error {
	s := cr.Spec.ForProvider
	config := postgresql.Configuration{
		ConfigurationProperties: &postgresql.ConfigurationProperties{
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

// UpdatePostgreSQLConfigurationObservation produces SQLServerConfigurationObservation from postgresql.Configuration.
func UpdatePostgreSQLConfigurationObservation(o *azuredbv1beta1.SQLServerConfigurationObservation, in postgresql.Configuration) {
	o.ID = azure.ToString(in.ID)
	o.Name = azure.ToString(in.Name)
	o.Type = azure.ToString(in.Type)
	o.DataType = azure.ToString(in.DataType)
	o.Value = azure.ToString(in.Value)
	o.DefaultValue = azure.ToString(in.DefaultValue)
	o.Source = azure.ToString(in.Source)
	o.Description = azure.ToString(in.Description)
}

// LateInitializePostgreSQLConfiguration fills the empty values of SQLServerConfigurationParameters with the
// ones that are retrieved from the Azure API.
func LateInitializePostgreSQLConfiguration(p *azuredbv1beta1.SQLServerConfigurationParameters, in postgresql.Configuration) {
	p.Value = azure.LateInitializeStringPtrFromPtr(p.Value, in.Value)
}

// IsPostgreSQLConfigurationUpToDate is used to report whether given postgresql.Configuration is in
// sync with the SQLServerConfigurationParameters that user desires.
func IsPostgreSQLConfigurationUpToDate(p azuredbv1beta1.SQLServerConfigurationParameters, in postgresql.Configuration) bool {
	return azure.ToString(p.Value) == azure.ToString(in.Value)
}
