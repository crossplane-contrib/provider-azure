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

package fake

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql/mysqlapi"
	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql/postgresqlapi"
)

var _ mysqlapi.VirtualNetworkRulesClientAPI = &MockMySQLVirtualNetworkRulesClient{}

// MockMySQLVirtualNetworkRulesClient is a fake implementation of mysql.VirtualNetworkRulesClient.
type MockMySQLVirtualNetworkRulesClient struct {
	mysqlapi.VirtualNetworkRulesClientAPI

	MockCreateOrUpdate func(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string, parameters mysql.VirtualNetworkRule) (result mysql.VirtualNetworkRulesCreateOrUpdateFuture, err error)
	MockDelete         func(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string) (result mysql.VirtualNetworkRulesDeleteFuture, err error)
	MockGet            func(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string) (result mysql.VirtualNetworkRule, err error)
}

// CreateOrUpdate calls the MockMySQLVirtualNetworkRulesClient's MockCreateOrUpdate method.
func (c *MockMySQLVirtualNetworkRulesClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string, parameters mysql.VirtualNetworkRule) (result mysql.VirtualNetworkRulesCreateOrUpdateFuture, err error) {
	return c.MockCreateOrUpdate(ctx, resourceGroupName, serverName, virtualNetworkRuleName, parameters)
}

// Delete calls the MockMySQLVirtualNetworkRulesClient's MockDelete method.
func (c *MockMySQLVirtualNetworkRulesClient) Delete(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string) (result mysql.VirtualNetworkRulesDeleteFuture, err error) {
	return c.MockDelete(ctx, resourceGroupName, serverName, virtualNetworkRuleName)
}

// Get calls the MockMySQLVirtualNetworkRulesClient's MockGet method.
func (c *MockMySQLVirtualNetworkRulesClient) Get(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string) (result mysql.VirtualNetworkRule, err error) {
	return c.MockGet(ctx, resourceGroupName, serverName, virtualNetworkRuleName)
}

var _ postgresqlapi.VirtualNetworkRulesClientAPI = &MockPostgreSQLVirtualNetworkRulesClient{}

// MockPostgreSQLVirtualNetworkRulesClient is a fake implementation of postgresql.VirtualNetworkRulesClient.
type MockPostgreSQLVirtualNetworkRulesClient struct {
	postgresqlapi.VirtualNetworkRulesClientAPI

	MockCreateOrUpdate func(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string, parameters postgresql.VirtualNetworkRule) (result postgresql.VirtualNetworkRulesCreateOrUpdateFuture, err error)
	MockDelete         func(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string) (result postgresql.VirtualNetworkRulesDeleteFuture, err error)
	MockGet            func(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string) (result postgresql.VirtualNetworkRule, err error)
}

// CreateOrUpdate calls the MockPostgreSQLVirtualNetworkRulesClient's MockCreateOrUpdate method.
func (c *MockPostgreSQLVirtualNetworkRulesClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string, parameters postgresql.VirtualNetworkRule) (result postgresql.VirtualNetworkRulesCreateOrUpdateFuture, err error) {
	return c.MockCreateOrUpdate(ctx, resourceGroupName, serverName, virtualNetworkRuleName, parameters)
}

// Delete calls the MockPostgreSQLVirtualNetworkRulesClient's MockDelete method.
func (c *MockPostgreSQLVirtualNetworkRulesClient) Delete(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string) (result postgresql.VirtualNetworkRulesDeleteFuture, err error) {
	return c.MockDelete(ctx, resourceGroupName, serverName, virtualNetworkRuleName)
}

// Get calls the MockPostgreSQLVirtualNetworkRulesClient's MockGet method.
func (c *MockPostgreSQLVirtualNetworkRulesClient) Get(ctx context.Context, resourceGroupName string, serverName string, virtualNetworkRuleName string) (result postgresql.VirtualNetworkRule, err error) {
	return c.MockGet(ctx, resourceGroupName, serverName, virtualNetworkRuleName)
}

var _ mysqlapi.FirewallRulesClientAPI = &MockMySQLFirewallRulesClient{}

// MockMySQLFirewallRulesClient is a fake implementation of mysql.FirewallRulesClient.
type MockMySQLFirewallRulesClient struct {
	mysqlapi.FirewallRulesClientAPI

	MockCreateOrUpdate func(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string, parameters mysql.FirewallRule) (result mysql.FirewallRulesCreateOrUpdateFuture, err error)
	MockDelete         func(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string) (result mysql.FirewallRulesDeleteFuture, err error)
	MockGet            func(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string) (result mysql.FirewallRule, err error)
}

// CreateOrUpdate calls the MockMySQLFirewallRulesClient's MockCreateOrUpdate method.
func (c *MockMySQLFirewallRulesClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string, parameters mysql.FirewallRule) (result mysql.FirewallRulesCreateOrUpdateFuture, err error) {
	return c.MockCreateOrUpdate(ctx, resourceGroupName, serverName, firewallRuleName, parameters)
}

// Delete calls the MockMySQLFirewallRulesClient's MockDelete method.
func (c *MockMySQLFirewallRulesClient) Delete(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string) (result mysql.FirewallRulesDeleteFuture, err error) {
	return c.MockDelete(ctx, resourceGroupName, serverName, firewallRuleName)
}

// Get calls the MockMySQLFirewallRulesClient's MockGet method.
func (c *MockMySQLFirewallRulesClient) Get(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string) (result mysql.FirewallRule, err error) {
	return c.MockGet(ctx, resourceGroupName, serverName, firewallRuleName)
}

var _ postgresqlapi.FirewallRulesClientAPI = &MockPostgreSQLFirewallRulesClient{}

// MockPostgreSQLFirewallRulesClient is a fake implementation of postgresql.FirewallRulesClient.
type MockPostgreSQLFirewallRulesClient struct {
	postgresqlapi.FirewallRulesClientAPI

	MockCreateOrUpdate func(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string, parameters postgresql.FirewallRule) (result postgresql.FirewallRulesCreateOrUpdateFuture, err error)
	MockDelete         func(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string) (result postgresql.FirewallRulesDeleteFuture, err error)
	MockGet            func(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string) (result postgresql.FirewallRule, err error)
}

// CreateOrUpdate calls the MockPostgreSQLFirewallRulesClient's MockCreateOrUpdate method.
func (c *MockPostgreSQLFirewallRulesClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string, parameters postgresql.FirewallRule) (result postgresql.FirewallRulesCreateOrUpdateFuture, err error) {
	return c.MockCreateOrUpdate(ctx, resourceGroupName, serverName, firewallRuleName, parameters)
}

// Delete calls the MockPostgreSQLFirewallRulesClient's MockDelete method.
func (c *MockPostgreSQLFirewallRulesClient) Delete(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string) (result postgresql.FirewallRulesDeleteFuture, err error) {
	return c.MockDelete(ctx, resourceGroupName, serverName, firewallRuleName)
}

// Get calls the MockPostgreSQLFirewallRulesClient's MockGet method.
func (c *MockPostgreSQLFirewallRulesClient) Get(ctx context.Context, resourceGroupName string, serverName string, firewallRuleName string) (result postgresql.FirewallRule, err error) {
	return c.MockGet(ctx, resourceGroupName, serverName, firewallRuleName)
}
