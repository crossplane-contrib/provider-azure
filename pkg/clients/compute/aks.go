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

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2022-01-01/containerservice"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	"github.com/crossplane-contrib/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
)

const (
	// AgentPoolProfileName is a format string for the name of the automatically
	// created cluster agent pool profile
	AgentPoolProfileName = "agentpool"
)

const (
	// error strings
	errInvalidUserAssignedManagedIdentity = "at least one user-assigned managed identity resource name must be specified when identity.type is UserAssigned"
)

// An AKSClient can create, read, and delete AKS clusters and the various other
// resources they require.
type AKSClient interface {
	GetManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error)
	EnsureManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) error
	DeleteManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) error
	GetKubeConfig(ctx context.Context, ac *v1alpha3.AKSCluster) ([]byte, error)
}

// An AggregateClient aggregates the various clients used by the AKS controller.
type AggregateClient struct {
	ManagedClusters containerservice.ManagedClustersClient
	subscriptionID  string
}

// NewAggregateClient produces the various clients used by the AKS controller.
func NewAggregateClient(subscriptionID string, auth autorest.Authorizer) (AKSClient, error) {
	mcc := containerservice.NewManagedClustersClient(subscriptionID)
	mcc.Authorizer = auth
	_ = mcc.AddToUserAgent(azure.UserAgent)

	return AggregateClient{
		ManagedClusters: mcc,
		subscriptionID:  subscriptionID,
	}, nil
}

// GetManagedCluster returns the requested Azure managed cluster.
func (c AggregateClient) GetManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error) {
	return c.ManagedClusters.Get(ctx, ac.Spec.ResourceGroupName, meta.GetExternalName(ac))
}

// EnsureManagedCluster ensures the supplied AKS cluster exists.
func (c AggregateClient) EnsureManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) error {
	mc, err := newManagedCluster(ac)
	if err != nil {
		return err
	}
	_, err = c.ManagedClusters.CreateOrUpdate(ctx, ac.Spec.ResourceGroupName, meta.GetExternalName(ac), mc)
	return err
}

// DeleteManagedCluster deletes the supplied AKS cluster.
func (c AggregateClient) DeleteManagedCluster(ctx context.Context, ac *v1alpha3.AKSCluster) error {
	_, err := c.ManagedClusters.Delete(ctx, ac.Spec.ResourceGroupName, meta.GetExternalName(ac))
	return err
}

// GetKubeConfig produces a kubeconfig file that configures access to the
// supplied AKS cluster.
func (c AggregateClient) GetKubeConfig(ctx context.Context, ac *v1alpha3.AKSCluster) ([]byte, error) {
	creds, err := c.ManagedClusters.ListClusterAdminCredentials(ctx, ac.Spec.ResourceGroupName, meta.GetExternalName(ac), "")
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

func newManagedCluster(c *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error) {
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
					VMSize: to.StringPtr(c.Spec.NodeVMSize),
					Mode:   containerservice.AgentPoolModeSystem,
				},
			},
			EnableRBAC: to.BoolPtr(!c.Spec.DisableRBAC),
		},
		Identity: &containerservice.ManagedClusterIdentity{},
	}
	switch containerservice.ResourceIdentityType(c.Spec.Identity.Type) { //nolint:exhaustive
	case containerservice.ResourceIdentityTypeSystemAssigned:
		p.Identity.Type = containerservice.ResourceIdentityTypeSystemAssigned
	case containerservice.ResourceIdentityTypeUserAssigned:
		p.Identity.Type = containerservice.ResourceIdentityTypeUserAssigned
		if len(c.Spec.Identity.ResourceIDs) == 0 {
			return p, errors.New(errInvalidUserAssignedManagedIdentity)
		}
		p.Identity.UserAssignedIdentities = make(map[string]*containerservice.ManagedClusterIdentityUserAssignedIdentitiesValue, len(c.Spec.Identity.ResourceIDs))
		for _, resourceID := range c.Spec.Identity.ResourceIDs {
			p.Identity.UserAssignedIdentities[resourceID] = &containerservice.ManagedClusterIdentityUserAssignedIdentitiesValue{}
		}
	}

	if c.Spec.VnetSubnetID != "" {
		p.ManagedClusterProperties.NetworkProfile = &containerservice.NetworkProfile{NetworkPlugin: containerservice.NetworkPluginAzure}
		(*p.ManagedClusterProperties.AgentPoolProfiles)[0].VnetSubnetID = to.StringPtr(c.Spec.VnetSubnetID)
	}

	return p, nil
}
