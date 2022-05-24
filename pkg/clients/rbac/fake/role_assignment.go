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

	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization/authorizationapi"
)

var _ authorizationapi.RoleAssignmentsClientAPI = &MockRoleAssignmentClient{}

// MockRoleAssignmentClient is a fake implementation of graphrbacapi.ApplicationsClientAPI.
type MockRoleAssignmentClient struct {
	authorizationapi.RoleAssignmentsClientAPI

	MockCreate               func(ctx context.Context, scope string, roleAssignmentName string, parameters authorization.RoleAssignmentCreateParameters) (result authorization.RoleAssignment, err error)
	MockListForScopeComplete func(ctx context.Context, scope string, filter string) (result authorization.RoleAssignmentListResultIterator, err error)
	MockDelete               func(ctx context.Context, scope string, roleAssignmentName string) (result authorization.RoleAssignment, err error)
}

// Create calls the MockRoleAssignmentClient's MockCreate method.
func (c *MockRoleAssignmentClient) Create(ctx context.Context, scope string, roleAssignmentName string, parameters authorization.RoleAssignmentCreateParameters) (result authorization.RoleAssignment, err error) {
	return c.MockCreate(ctx, scope, roleAssignmentName, parameters)
}

// Delete calls the MockRoleAssignmentClient's MockDelete method.
func (c *MockRoleAssignmentClient) Delete(ctx context.Context, scope string, roleAssignmentName string) (result authorization.RoleAssignment, err error) {
	return c.MockDelete(ctx, scope, roleAssignmentName)
}

// ListForScopeComplete calls the MockRoleAssignmentClient's MockListForScopeComplete method.
func (c *MockRoleAssignmentClient) ListForScopeComplete(ctx context.Context, scope string, filter string) (result authorization.RoleAssignmentListResultIterator, err error) {
	return c.MockListForScopeComplete(ctx, scope, filter)
}
