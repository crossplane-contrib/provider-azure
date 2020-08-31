/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the c.Specific language governing permissions and
limitations under the License.
*/

package compute

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	authorizationmgmt "github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

const (
	// AgentPoolProfileName is a format string for the name of the automatically
	// created cluster agent pool profile
	AgentPoolProfileName = "agentpool"

	// NetworkContributorRoleID lets the AKS cluster managed networks, but not
	// access them.
	NetworkContributorRoleID = "/providers/Microsoft.Authorization/roleDefinitions/4d97b98b-1d4f-4787-a291-c67834d212e7"

	appCredsValidYears = 5
)

// An AKSClient can create, read, and delete AKS clusters and the various other
// resources they require.
type AKSClient interface {
	GetManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error)
	EnsureManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster, secret string) error
	DeleteManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) error
	GetKubeConfig(ctx context.Context, ac *v1alpha3.AKSCluster) ([]byte, error)
}

// An AggregateClient aggregates the various clients used by the AKS controller.
type AggregateClient struct {
	ManagedClusters   containerservice.ManagedClustersClient
	Applications      graphrbac.ApplicationsClient
	ServicePrincipals graphrbac.ServicePrincipalsClient
	RoleAssignments   authorization.RoleAssignmentsClient
}

// NewAggregateClient produces the various clients used by the AKS controller.
func NewAggregateClient(creds map[string]string, auth autorest.Authorizer) (AKSClient, error) {
	mcc := containerservice.NewManagedClustersClient(creds[azure.CredentialsKeySubscriptionID])
	mcc.Authorizer = auth
	_ = mcc.AddToUserAgent(azure.UserAgent)

	rac := authorization.NewRoleAssignmentsClient(creds[azure.CredentialsKeySubscriptionID])
	rac.Authorizer = auth
	_ = rac.AddToUserAgent(azure.UserAgent)

	cfg, err := adal.NewOAuthConfig(creds[azure.CredentialsKeyActiveDirectoryEndpointURL], creds[azure.CredentialsKeyTenantID])
	if err != nil {
		return nil, errors.Wrap(err, "cannot create OAuth configuration")
	}

	token, err := adal.NewServicePrincipalToken(*cfg,
		creds[azure.CredentialsKeyClientID],
		creds[azure.CredentialsKeyClientSecret],
		creds[azure.CredentialsKeyActiveDirectoryGraphResourceID])
	if err != nil {
		return nil, errors.Wrap(err, "cannot create service principal token")
	}
	if err := token.Refresh(); err != nil {
		return nil, errors.Wrap(err, "cannot refresh service principal token")
	}

	ta := autorest.NewBearerAuthorizer(token)

	ac := graphrbac.NewApplicationsClient(creds[azure.CredentialsKeyTenantID])
	ac.Authorizer = ta
	_ = ac.AddToUserAgent(azure.UserAgent)

	spc := graphrbac.NewServicePrincipalsClient(creds[azure.CredentialsKeyTenantID])
	spc.Authorizer = ta
	_ = spc.AddToUserAgent(azure.UserAgent)

	return AggregateClient{
		ManagedClusters:   mcc,
		Applications:      ac,
		ServicePrincipals: spc,
		RoleAssignments:   rac,
	}, nil
}

// GetManagedCluster returns the requested Azure managed cluster.
func (c AggregateClient) GetManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error) {
	return c.ManagedClusters.Get(ctx, ac.Spec.ResourceGroupName, meta.GetExternalName(ac))
}

// EnsureManagedCluster ensures the supplied AKS cluster exists, including
// ensuring any required service principals and role assignments exist.
func (c AggregateClient) EnsureManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster, secret string) error {
	app, err := c.ensureApplication(ctx, meta.GetExternalName(ac), secret)
	if err != nil {
		return err
	}

	sp, err := c.ensureServicePrincipal(ctx, to.String(app.AppID))
	if err != nil {
		return err
	}

	if err := c.ensureRoleAssignment(ctx, to.String(sp.ObjectID), NetworkContributorRoleID, ac.Spec.VnetSubnetID); err != nil {
		return err
	}

	mc := newManagedCluster(ac, to.String(app.AppID), secret)
	_, err = c.ManagedClusters.CreateOrUpdate(ctx, ac.Spec.ResourceGroupName, meta.GetExternalName(ac), mc)
	return err
}

// DeleteManagedCluster deletes the supplied AKS cluster, including its service
// principals and any role assignments.
func (c AggregateClient) DeleteManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) error {
	if err := c.deleteApplication(ctx, meta.GetExternalName(ac)); err != nil {
		return err
	}
	_, err := c.ManagedClusters.Delete(ctx, ac.Spec.ResourceGroupName, meta.GetExternalName(ac))
	return err
}

// GetKubeConfig produces a kubeconfig file that configures access to the
// supplied AKS cluster.
func (c AggregateClient) GetKubeConfig(ctx context.Context, ac *v1alpha3.AKSCluster) ([]byte, error) {
	creds, err := c.ManagedClusters.ListClusterAdminCredentials(ctx, ac.Spec.ResourceGroupName, meta.GetExternalName(ac))
	if err != nil {
		return nil, err
	}

	// TODO(negz): It's not clear in what case this would contain more than one kubeconfig file.
	// https://docs.microsoft.com/en-us/rest/api/aks/managedclusters/listclusteradmincredentials#credentialresults
	if creds.Kubeconfigs == nil || len(*creds.Kubeconfigs) == 0 || (*creds.Kubeconfigs)[0].Value == nil {
		return nil, errors.Errorf("zero kubeconfig credentials returned")
	}
	// Azure's generated Godoc claims Value is a 'base64 encoded kubeconfig'.
	// This is true on the wire, but not true in the actual struct because
	// encoding/json automatically base64 encodes and decodes byte slices.
	return *((*creds.Kubeconfigs)[0].Value), nil
}

func (c AggregateClient) ensureApplication(ctx context.Context, name, secret string) (graphrbac.Application, error) {
	pc, err := newPasswordCredential(secret)
	if err != nil {
		return graphrbac.Application{}, err
	}

	filter := fmt.Sprintf("displayName eq '%s'", name)
	for l, err := c.Applications.ListComplete(ctx, filter); l.NotDone(); err = l.NextWithContext(ctx) {
		if err != nil {
			return graphrbac.Application{}, err
		}
		p := graphrbac.PasswordCredentialsUpdateParameters{Value: &[]graphrbac.PasswordCredential{pc}}
		if _, err := c.Applications.UpdatePasswordCredentials(ctx, to.String(l.Value().ObjectID), p); err != nil {
			return graphrbac.Application{}, err
		}

		// We really do want to stop here if we found an app with our desired
		// display name. We presume it's one we created earlier.
		return l.Value(), nil // nolint:staticcheck
	}

	url := fmt.Sprintf("https://%s.aks.crossplane.io", name)
	p := graphrbac.ApplicationCreateParameters{
		AvailableToOtherTenants: to.BoolPtr(false),
		DisplayName:             to.StringPtr(name),
		Homepage:                to.StringPtr(url),
		IdentifierUris:          &[]string{url},
		PasswordCredentials:     &[]graphrbac.PasswordCredential{pc},
	}
	if err != nil {
		return graphrbac.Application{}, err
	}

	return c.Applications.Create(ctx, p)
}

func (c AggregateClient) ensureServicePrincipal(ctx context.Context, appID string) (graphrbac.ServicePrincipal, error) {
	r, err := c.Applications.GetServicePrincipalsIDByAppID(ctx, appID)
	if azure.IsNotFound(err) {
		// Create it.
		p := graphrbac.ServicePrincipalCreateParameters{AppID: to.StringPtr(appID), AccountEnabled: to.BoolPtr(true)}
		return c.ServicePrincipals.Create(ctx, p)
	}
	if err != nil {
		return graphrbac.ServicePrincipal{}, err
	}

	return c.ServicePrincipals.Get(ctx, to.String(r.Value))
}

func (c AggregateClient) ensureRoleAssignment(ctx context.Context, principalID, roleID, scope string) error {
	// If scope was the empty string we probably needed a role assignment for
	// an optional scope, for example a subnetwork.
	if scope == "" {
		return nil
	}

	name, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	filter := fmt.Sprintf("principalId eq '%s'", principalID)
	for l, err := c.RoleAssignments.ListForScopeComplete(ctx, scope, filter); l.NotDone(); err = l.NextWithContext(ctx) {
		if err != nil {
			return err
		}

		// We really do want to stop here if our principal already has a role
		// definition for this scope; we presume it's one we created earlier.
		return nil // nolint:staticcheck
	}

	p := authorizationmgmt.RoleAssignmentCreateParameters{Properties: &authorizationmgmt.RoleAssignmentProperties{
		RoleDefinitionID: azure.ToStringPtr(fmt.Sprintf("/subscriptions/%s%s", c.RoleAssignments.SubscriptionID, roleID)),
		PrincipalID:      azure.ToStringPtr(principalID),
	}}
	_, err = c.RoleAssignments.Create(ctx, scope, name.String(), p)
	return err
}

func (c AggregateClient) deleteApplication(ctx context.Context, name string) error {
	filter := fmt.Sprintf("displayName eq '%s'", name)
	for l, err := c.Applications.ListComplete(ctx, filter); l.NotDone(); err = l.NextWithContext(ctx) {
		if err != nil {
			return err
		}

		// We really do want to delete the first matching application we find.
		_, err := c.Applications.Delete(ctx, to.String(l.Value().ObjectID))
		return resource.Ignore(azure.IsNotFound, err) // nolint:staticcheck
	}

	return nil
}

func newManagedCluster(c *v1alpha3.AKSCluster, appID, secret string) containerservice.ManagedCluster {
	nodeCount := int32(v1alpha3.DefaultNodeCount)
	if c.Spec.NodeCount != nil {
		nodeCount = int32(*c.Spec.NodeCount)
	}

	p := containerservice.ManagedCluster{
		Name:     to.StringPtr(meta.GetExternalName(c)),
		Location: to.StringPtr(c.Spec.Location),
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			KubernetesVersion: to.StringPtr(c.Spec.Version),
			DNSPrefix:         to.StringPtr(c.Spec.DNSNamePrefix),
			AgentPoolProfiles: &[]containerservice.ManagedClusterAgentPoolProfile{
				{
					Name:   to.StringPtr(AgentPoolProfileName),
					Count:  &nodeCount,
					VMSize: containerservice.VMSizeTypes(c.Spec.NodeVMSize),
				},
			},
			ServicePrincipalProfile: &containerservice.ManagedClusterServicePrincipalProfile{
				ClientID: to.StringPtr(appID),
				Secret:   to.StringPtr(secret),
			},
			EnableRBAC: to.BoolPtr(!c.Spec.DisableRBAC),
		},
	}

	if c.Spec.VnetSubnetID != "" {
		p.ManagedClusterProperties.NetworkProfile = &containerservice.NetworkProfile{NetworkPlugin: containerservice.Azure}
		p.ManagedClusterProperties.AgentPoolProfiles = &[]containerservice.ManagedClusterAgentPoolProfile{
			{
				Name:         to.StringPtr(AgentPoolProfileName),
				Count:        &nodeCount,
				VMSize:       containerservice.VMSizeTypes(c.Spec.NodeVMSize),
				VnetSubnetID: to.StringPtr(c.Spec.VnetSubnetID),
			},
		}
	}

	return p
}

func newPasswordCredential(secret string) (graphrbac.PasswordCredential, error) {
	keyID, err := uuid.NewRandom()
	return graphrbac.PasswordCredential{
		StartDate: &date.Time{Time: time.Now()},
		EndDate:   &date.Time{Time: time.Now().AddDate(appCredsValidYears, 0, 0)},
		KeyID:     to.StringPtr(keyID.String()),
		Value:     to.StringPtr(secret),
	}, err
}
