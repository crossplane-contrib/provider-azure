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

package agentpool

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-03-01/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-03-01/containerservice/containerserviceapi"
)

var _ containerserviceapi.AgentPoolsClientAPI = &Mock{}

// Mock is mocked AgentPoolsClientAPI
type Mock struct {
	containerserviceapi.AgentPoolsClientAPI

	MockCreateOrUpdate func(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string, parameters containerservice.AgentPool) (result containerservice.AgentPoolsCreateOrUpdateFuture, err error)
	MockDelete         func(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string) (result containerservice.AgentPoolsDeleteFuture, err error)
	MockGet            func(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string) (result containerservice.AgentPool, err error)
}

// CreateOrUpdate calls the Mock's MockCreateOrUpdate method.
func (c *Mock) CreateOrUpdate(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string, parameters containerservice.AgentPool) (result containerservice.AgentPoolsCreateOrUpdateFuture, err error) {
	return c.MockCreateOrUpdate(ctx, resourceGroupName, resourceName, agentPoolName, parameters)
}

// Delete calls the Mock's MockDelete method.
func (c *Mock) Delete(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string) (result containerservice.AgentPoolsDeleteFuture, err error) {
	return c.MockDelete(ctx, resourceGroupName, resourceName, agentPoolName)
}

// Get calls the Mock's MockGet method.
func (c *Mock) Get(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string) (result containerservice.AgentPool, err error) {
	return c.MockGet(ctx, resourceGroupName, resourceName, agentPoolName)
}
