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

package fake

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac/graphrbacapi"
	"github.com/Azure/go-autorest/autorest"
)

var _ graphrbacapi.ApplicationsClientAPI = &MockApplicationsClient{}

// MockApplicationsClient is a fake implementation of graphrbacapi.ApplicationsClientAPI.
type MockApplicationsClient struct {
	graphrbacapi.ApplicationsClientAPI

	MockCreate func(ctx context.Context, parameters graphrbac.ApplicationCreateParameters) (result graphrbac.Application, err error)
	MockDelete func(ctx context.Context, applicationObjectID string) (result autorest.Response, err error)
	MockGet    func(ctx context.Context, applicationObjectID string) (result graphrbac.Application, err error)
}

// Create calls the MockApplicationsClient's MockCreateOrUpdate method.
func (c *MockApplicationsClient) Create(ctx context.Context, parameters graphrbac.ApplicationCreateParameters) (result graphrbac.Application, err error) {
	return c.MockCreate(ctx, parameters)
}

// Delete calls the MockApplicationsClient's MockDelete method.
func (c *MockApplicationsClient) Delete(ctx context.Context, applicationObjectID string) (result autorest.Response, err error) {
	return c.MockDelete(ctx, applicationObjectID)
}

// Get calls the MockApplicationsClient's MockGet method.
func (c *MockApplicationsClient) Get(ctx context.Context, applicationObjectID string) (result graphrbac.Application, err error) {
	return c.MockGet(ctx, applicationObjectID)
}
