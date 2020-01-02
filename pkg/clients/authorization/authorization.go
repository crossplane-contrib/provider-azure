/*
Copyright 2020 The Crossplane Authors.

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

package authorization

import (
	"context"
	"fmt"

	authorizationmgmt "github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization/authorizationapi"

	azure "github.com/crossplaneio/stack-azure/pkg/clients"
)

const (
	networkContributorRoleID = "/providers/Microsoft.Authorization/roleDefinitions/4d97b98b-1d4f-4787-a291-c67834d212e7"
)

// RoleAssignmentsAPI defines methods available to a Role Assignments client
type RoleAssignmentsAPI interface {
	CreateRoleAssignment(ctx context.Context, sp, vnetSubnetID, name string) (result *authorizationmgmt.RoleAssignment, err error)
	DeleteRoleAssignment(ctx context.Context, vnetSubnetID, name string) error
}

// A RoleAssignmentsClient handles CRUD operations for Azure Role Assignments.
type RoleAssignmentsClient struct {
	subscriptionID string
	client         authorizationapi.RoleAssignmentsClientAPI
}

// NewRoleAssignmentsClient returns a new Azure Role Assignments client.
func NewRoleAssignmentsClient(c *azure.Client) (*RoleAssignmentsClient, error) {
	roleAssignmentsClient := authorizationmgmt.NewRoleAssignmentsClient(c.SubscriptionID)
	roleAssignmentsClient.Authorizer = c.Authorizer
	roleAssignmentsClient.AddToUserAgent(azure.UserAgent)

	return &RoleAssignmentsClient{
		subscriptionID: c.SubscriptionID,
		client:         roleAssignmentsClient,
	}, nil
}

// CreateRoleAssignment creates an Azure network contributor role assignment for
// a service principal
func (r *RoleAssignmentsClient) CreateRoleAssignment(ctx context.Context, sp, vnetSubnetID, name string) (*authorizationmgmt.RoleAssignment, error) {

	subscriptionRole := fmt.Sprintf("/subscriptions/%s", r.subscriptionID) + networkContributorRoleID

	parameters := authorizationmgmt.RoleAssignmentCreateParameters{
		Properties: &authorizationmgmt.RoleAssignmentProperties{
			RoleDefinitionID: azure.ToStringPtr(subscriptionRole),
			PrincipalID:      azure.ToStringPtr(sp),
		},
	}

	role, err := r.client.Create(ctx, vnetSubnetID, name, parameters)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

// DeleteRoleAssignment will delete the given role assignemt
func (r *RoleAssignmentsClient) DeleteRoleAssignment(ctx context.Context, vnetSubnetID, name string) error {
	_, err := r.client.Delete(ctx, vnetSubnetID, name)
	return err
}
