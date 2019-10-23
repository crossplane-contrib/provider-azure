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

package compute

import (
	"context"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice"
	"github.com/Azure/go-autorest/autorest/to"

	computev1alpha2 "github.com/crossplaneio/stack-azure/apis/compute/v1alpha2"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
	"github.com/crossplaneio/stack-azure/pkg/clients/authorization"
)

const (
	// AgentPoolProfileName is a format string for the name of the automatically
	// created cluster agent pool profile
	AgentPoolProfileName = "agentpool"

	maxClusterNameLen = 31
)

// AKSSetupClient is a type that implements all of the AKS setup interface
type AKSSetupClient struct {
	AKSClusterAPI
	azure.ApplicationAPI
	azure.ServicePrincipalAPI
	authorization.RoleAssignmentsAPI
}

// AKSSetupAPIFactory is an interface that can create instances of the AKSSetupClient
type AKSSetupAPIFactory interface {
	CreateSetupClient(c *azure.Client) (*AKSSetupClient, error)
}

// AKSSetupClientFactory implements the AKSSetupAPIFactory interface by returning real clients that talk to Azure APIs
type AKSSetupClientFactory struct {
}

// CreateSetupClient creates and returns an AKS setup client that is ready to talk to Azure APIs
func (f *AKSSetupClientFactory) CreateSetupClient(c *azure.Client) (*AKSSetupClient, error) {
	aksClusterClient, err := NewAKSClusterClient(c)
	if err != nil {
		return nil, err
	}

	appClient, err := azure.NewApplicationClient(c)
	if err != nil {
		return nil, err
	}

	spClient, err := azure.NewServicePrincipalClient(c)
	if err != nil {
		return nil, err
	}

	raClient, err := authorization.NewRoleAssignmentsClient(c)
	if err != nil {
		return nil, err
	}

	return &AKSSetupClient{
		AKSClusterAPI:       aksClusterClient,
		ApplicationAPI:      appClient,
		ServicePrincipalAPI: spClient,
		RoleAssignmentsAPI:  raClient,
	}, nil
}

// AKSClusterAPI represents the API interface for a AKS Cluster client
type AKSClusterAPI interface {
	Get(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.ManagedCluster, error)
	CreateOrUpdateBegin(ctx context.Context, instance computev1alpha2.AKSCluster, clusterName, appID, spSecret string) ([]byte, error)
	CreateOrUpdateEnd(op []byte) (bool, error)
	Delete(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.ManagedClustersDeleteFuture, error)
	ListClusterAdminCredentials(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.CredentialResults, error)
}

// AKSClusterClient is the concreate implementation of the AKSClusterAPI interface that calls Azure API.
type AKSClusterClient struct {
	containerservice.ManagedClustersClient
}

// NewAKSClusterClient creates and initializes a AKSClusterClient instance.
func NewAKSClusterClient(c *azure.Client) (*AKSClusterClient, error) {
	aksClustersClient := containerservice.NewManagedClustersClient(c.SubscriptionID)
	aksClustersClient.Authorizer = c.Authorizer
	aksClustersClient.AddToUserAgent(azure.UserAgent)

	return &AKSClusterClient{aksClustersClient}, nil
}

// Get returns the AKS cluster details for the given instance
func (c *AKSClusterClient) Get(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.ManagedCluster, error) {
	return c.ManagedClustersClient.Get(ctx, instance.Spec.ResourceGroupName, instance.Status.ClusterName)
}

// CreateOrUpdateBegin begins the create/update operation for a AKS Cluster with the given properties
func (c *AKSClusterClient) CreateOrUpdateBegin(ctx context.Context, instance computev1alpha2.AKSCluster, clusterName, appID, spSecret string) ([]byte, error) {
	spec := instance.Spec

	enableRBAC := !spec.DisableRBAC

	nodeCount := int32(computev1alpha2.DefaultNodeCount)
	if spec.NodeCount != nil {
		nodeCount = int32(*spec.NodeCount)
	}

	createParams := containerservice.ManagedCluster{
		Name:     &clusterName,
		Location: &spec.Location,
		ManagedClusterProperties: &containerservice.ManagedClusterProperties{
			KubernetesVersion: &spec.Version,
			DNSPrefix:         &spec.DNSNamePrefix,
			AgentPoolProfiles: &[]containerservice.ManagedClusterAgentPoolProfile{
				{
					Name:   to.StringPtr(AgentPoolProfileName),
					Count:  &nodeCount,
					VMSize: containerservice.VMSizeTypes(spec.NodeVMSize),
				},
			},
			ServicePrincipalProfile: &containerservice.ManagedClusterServicePrincipalProfile{
				ClientID: to.StringPtr(appID),
				Secret:   to.StringPtr(spSecret),
			},
			EnableRBAC: &enableRBAC,
		},
	}

	if spec.VnetSubnetID != "" {
		createParams.ManagedClusterProperties.NetworkProfile = &containerservice.NetworkProfile{
			NetworkPlugin: containerservice.Azure,
		}

		createParams.ManagedClusterProperties.AgentPoolProfiles = &[]containerservice.ManagedClusterAgentPoolProfile{
			{
				Name:         to.StringPtr(AgentPoolProfileName),
				Count:        &nodeCount,
				VMSize:       containerservice.VMSizeTypes(spec.NodeVMSize),
				VnetSubnetID: to.StringPtr(spec.VnetSubnetID),
			},
		}
	}

	createFuture, err := c.CreateOrUpdate(ctx, instance.Spec.ResourceGroupName, clusterName, createParams)
	if err != nil {
		return nil, err
	}

	// serialize the create operation
	createFutureJSON, err := createFuture.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return createFutureJSON, nil
}

// CreateOrUpdateEnd checks to see if the given create/update operation is completed and if any error has occurred.
func (c *AKSClusterClient) CreateOrUpdateEnd(op []byte) (done bool, err error) {
	// unmarshal the given create complete data into a future object
	future := &containerservice.ManagedClustersCreateOrUpdateFuture{}
	if err = future.UnmarshalJSON(op); err != nil {
		return false, err
	}

	// check if the operation is done yet
	done, err = future.DoneWithContext(context.Background(), c.Client)
	if !done {
		return false, err
	}

	// check the result of the completed operation
	if _, err = future.Result(c.ManagedClustersClient); err != nil {
		return true, err
	}

	return true, nil
}

// Delete begins the deletion operator for the given AKS cluster instance
func (c *AKSClusterClient) Delete(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.ManagedClustersDeleteFuture, error) {
	return c.ManagedClustersClient.Delete(ctx, instance.Spec.ResourceGroupName, instance.Status.ClusterName)
}

// ListClusterAdminCredentials will return the admin credentials used to connect to the given AKS cluster
func (c *AKSClusterClient) ListClusterAdminCredentials(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.CredentialResults, error) {
	return c.ManagedClustersClient.ListClusterAdminCredentials(ctx, instance.Spec.ResourceGroupName, instance.Status.ClusterName)
}

// SanitizeClusterName sanitizes the given AKS cluster name
func SanitizeClusterName(name string) string {
	if len(name) > maxClusterNameLen {
		name = name[:maxClusterNameLen]
	}

	return strings.TrimSuffix(name, "-")
}
