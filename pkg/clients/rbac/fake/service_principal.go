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

var _ graphrbacapi.ServicePrincipalsClientAPI = &MockServicePrincipalClient{}

// MockServicePrincipalClient is a fake implementation of graphrbacapi.ServicePrincipalsClientAPI.
type MockServicePrincipalClient struct {
	graphrbacapi.ServicePrincipalsClientAPI
	MockCreate func(ctx context.Context, parameters graphrbac.ServicePrincipalCreateParameters) (result graphrbac.ServicePrincipal, err error)
	MockDelete func(ctx context.Context, objectID string) (result autorest.Response, err error)
	MockGet    func(ctx context.Context, objectID string) (result graphrbac.ServicePrincipal, err error)
}

// Create calls the MockServicePrincipalClient's MockCreate method.
func (c *MockServicePrincipalClient) Create(ctx context.Context, parameters graphrbac.ServicePrincipalCreateParameters) (result graphrbac.ServicePrincipal, err error) {
	return c.MockCreate(ctx, parameters)
}

// Delete calls the MockServicePrincipalClient's MockDelete method.
func (c *MockServicePrincipalClient) Delete(ctx context.Context, objectID string) (result autorest.Response, err error) {
	return c.MockDelete(ctx, objectID)
}

// Get calls the MockServicePrincipalClient's MockGet method.
func (c *MockServicePrincipalClient) Get(ctx context.Context, objectID string) (result graphrbac.ServicePrincipal, err error) {
	return c.MockGet(ctx, objectID)
}
