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

package openshift

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	authorizationmgmt "github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/redhatopenshift/mgmt/2022-04-01/redhatopenshift"
	"github.com/crossplane-contrib/provider-azure/apis/containers/v1alpha1"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Resource states
const (
	ProvisioningStateCreating  = "Creating"
	ProvisioningStateDeleting  = "Deleting"
	ProvisioningStateFailed    = "Failed"
	ProvisioningStateSucceeded = "Succeeded"
	errGetPasswordSecretFailed = "Cannot get password secret"
	NetworkContributorRoleID   = "/providers/Microsoft.Authorization/roleDefinitions/4d97b98b-1d4f-4787-a291-c67834d212e7"
)

// NewCreateParameters returns Openshift cluster resource creation parameters suitable for
// use with the Azure API.
func NewCreateParameters(ctx context.Context, cr *v1alpha1.Openshift, c client.Client) redhatopenshift.OpenShiftCluster {
	return redhatopenshift.OpenShiftCluster{
		OpenShiftClusterProperties: &redhatopenshift.OpenShiftClusterProperties{
			IngressProfiles: &[]redhatopenshift.IngressProfile{
				{
					Name:       azure.ToStringPtr("default"),
					Visibility: redhatopenshift.VisibilityPublic,
				},
			},
			ApiserverProfile: &redhatopenshift.APIServerProfile{
				Visibility: redhatopenshift.VisibilityPublic,
			},
			ClusterProfile: &redhatopenshift.ClusterProfile{
				PullSecret:           azure.ToStringPtr(cr.Spec.ForProvider.ClusterProfile.PullSecret),
				Domain:               azure.ToStringPtr(cr.Spec.ForProvider.ClusterProfile.Domain),
				Version:              azure.ToStringPtr(cr.Spec.ForProvider.ClusterProfile.Version),
				ResourceGroupID:      azure.ToStringPtr(cr.Spec.ForProvider.ClusterProfile.ResourceGroupID),
				FipsValidatedModules: redhatopenshift.FipsValidatedModulesDisabled,
			},
			ServicePrincipalProfile: &redhatopenshift.ServicePrincipalProfile{
				ClientID:     azure.ToStringPtr(cr.Spec.ForProvider.ServicePrincipalProfile.ClientID),
				ClientSecret: azure.ToStringPtr(cr.Spec.ForProvider.ServicePrincipalProfile.ClientSecret),
			},
			NetworkProfile: &redhatopenshift.NetworkProfile{
				PodCidr:     azure.ToStringPtr(cr.Spec.ForProvider.NetworkProfile.PodCidr),
				ServiceCidr: azure.ToStringPtr(cr.Spec.ForProvider.NetworkProfile.ServiceCidr),
			},
			MasterProfile: &redhatopenshift.MasterProfile{
				VMSize:           azure.ToStringPtr(cr.Spec.ForProvider.MasterProfile.VMSize),
				SubnetID:         azure.ToStringPtr(cr.Spec.ForProvider.MasterProfile.SubnetID),
				EncryptionAtHost: redhatopenshift.EncryptionAtHostDisabled,
			},
			WorkerProfiles: &[]redhatopenshift.WorkerProfile{
				{
					Name:             azure.ToStringPtr("worker"),
					VMSize:           azure.ToStringPtr(cr.Spec.ForProvider.WorkerProfile.VMSize),
					DiskSizeGB:       azure.ToInt32Ptr(cr.Spec.ForProvider.WorkerProfile.DiskSizeGB),
					SubnetID:         azure.ToStringPtr(cr.Spec.ForProvider.WorkerProfile.SubnetID),
					Count:            azure.ToInt32Ptr(cr.Spec.ForProvider.WorkerProfile.Count),
					EncryptionAtHost: redhatopenshift.EncryptionAtHostDisabled,
				},
			},
		},
		Tags:     azure.ToStringPtrMap(cr.Spec.ForProvider.Tags),
		Location: azure.ToStringPtr(cr.Spec.ForProvider.Location),
	}
}

// NewUpdateParameters returns a redhatopenshift.OpenShiftClusterUpdate object only with changed
// fields.
// nolint:gocyclo
func NewUpdateParameters(spec v1alpha1.OpenshiftParameters, state redhatopenshift.OpenShiftCluster) redhatopenshift.OpenShiftClusterUpdate {
	patch := redhatopenshift.OpenShiftClusterUpdate{
		Tags: azure.ToStringPtrMap(spec.Tags),
		OpenShiftClusterProperties: &redhatopenshift.OpenShiftClusterProperties{
			IngressProfiles: &[]redhatopenshift.IngressProfile{
				{
					Name:       azure.ToStringPtr("default"),
					Visibility: redhatopenshift.VisibilityPublic,
				},
			},
			ApiserverProfile: &redhatopenshift.APIServerProfile{
				Visibility: redhatopenshift.VisibilityPublic,
			},
			ClusterProfile: &redhatopenshift.ClusterProfile{
				Domain:  azure.ToStringPtr(spec.ClusterProfile.Domain),
				Version: azure.ToStringPtr(spec.ClusterProfile.Version),
			},
			NetworkProfile: &redhatopenshift.NetworkProfile{
				PodCidr:     azure.ToStringPtr(spec.NetworkProfile.PodCidr),
				ServiceCidr: azure.ToStringPtr(spec.NetworkProfile.ServiceCidr),
			},
			MasterProfile: &redhatopenshift.MasterProfile{
				VMSize:   azure.ToStringPtr(spec.MasterProfile.VMSize),
				SubnetID: azure.ToStringPtr(spec.MasterProfile.SubnetID),
			},
		},
	}

	for k, v := range state.Tags {
		if patch.Tags[k] == v {
			delete(patch.Tags, k)
		}
	}
	if len(patch.Tags) == 0 {
		patch.Tags = nil
	}

	return patch
}

func LateInitialize(spec *v1alpha1.OpenshiftParameters, az redhatopenshift.OpenShiftCluster) {
	spec.Location = *azure.LateInitializeStringPtrFromVal(&spec.Location, *az.Location)
	spec.Tags = azure.LateInitializeStringMap(spec.Tags, az.Tags)
	spec.ClusterProfile.Domain = *azure.LateInitializeStringPtrFromVal(&spec.ClusterProfile.Domain, *az.OpenShiftClusterProperties.ClusterProfile.Domain)
	spec.ClusterProfile.ResourceGroupID = *azure.LateInitializeStringPtrFromVal(&spec.ClusterProfile.ResourceGroupID, *az.OpenShiftClusterProperties.ClusterProfile.ResourceGroupID)
	spec.ClusterProfile.Version = *azure.LateInitializeStringPtrFromVal(&spec.ClusterProfile.Version, *az.OpenShiftClusterProperties.ClusterProfile.Version)
	spec.NetworkProfile.PodCidr = *azure.LateInitializeStringPtrFromVal(&spec.NetworkProfile.PodCidr, *az.OpenShiftClusterProperties.NetworkProfile.PodCidr)
	spec.NetworkProfile.ServiceCidr = *azure.LateInitializeStringPtrFromVal(&spec.NetworkProfile.ServiceCidr, *az.OpenShiftClusterProperties.NetworkProfile.ServiceCidr)
	for _, worker := range *az.OpenShiftClusterProperties.WorkerProfiles {
		spec.WorkerProfile.Count = *azure.LateInitializeIntPtrFromInt32Ptr(&spec.WorkerProfile.Count, worker.Count)
		spec.WorkerProfile.DiskSizeGB = *azure.LateInitializeIntPtrFromInt32Ptr(&spec.WorkerProfile.DiskSizeGB, worker.DiskSizeGB)
		spec.WorkerProfile.SubnetID = *azure.LateInitializeStringPtrFromVal(&spec.WorkerProfile.SubnetID, *worker.SubnetID)
		spec.WorkerProfile.VMSize = *azure.LateInitializeStringPtrFromVal(&spec.WorkerProfile.VMSize, *worker.VMSize)
	}
	spec.MasterProfile.SubnetID = *azure.LateInitializeStringPtrFromVal(&spec.MasterProfile.SubnetID, *az.OpenShiftClusterProperties.MasterProfile.SubnetID)
	spec.MasterProfile.VMSize = *azure.LateInitializeStringPtrFromVal(&spec.MasterProfile.VMSize, *az.OpenShiftClusterProperties.MasterProfile.VMSize)
}

func GenerateObservation(az redhatopenshift.OpenShiftCluster) v1alpha1.OpenshiftObservation {
	o := v1alpha1.OpenshiftObservation{
		ID:   azure.ToString(az.ID),
		Name: azure.ToString(az.Name),
	}
	if az.OpenShiftClusterProperties == nil {
		return o
	}
	o.ProvisioningState = string(az.OpenShiftClusterProperties.ProvisioningState)
	return o
}

func ExtractSecrets(ctx context.Context, c client.Client, cr *v1alpha1.Openshift) (*v1alpha1.Openshift, error) {
	pullSecret, err := GetPassword(ctx, c, cr.Spec.ForProvider.ClusterProfile.PullSecretRef)
	if err != nil {
		return nil, err
	}
	clientID, err := GetPassword(ctx, c, cr.Spec.ForProvider.ServicePrincipalProfile.ClientIDRef)
	if err != nil {
		return nil, err
	}
	clientSecret, err := GetPassword(ctx, c, cr.Spec.ForProvider.ServicePrincipalProfile.ClientSecretRef)
	if err != nil {
		return nil, err
	}
	principalID, err := GetPassword(ctx, c, cr.Spec.ForProvider.ServicePrincipalProfile.PrincipalIDRef)
	if err != nil {
		return nil, err
	}
	cr.Spec.ForProvider.ClusterProfile.PullSecret = pullSecret
	cr.Spec.ForProvider.ServicePrincipalProfile.ClientID = clientID
	cr.Spec.ForProvider.ServicePrincipalProfile.ClientSecret = clientSecret
	cr.Spec.ForProvider.ServicePrincipalProfile.PrincipalID = principalID
	return cr, nil
}

func GetPassword(ctx context.Context, kube client.Client, in *xpv1.SecretKeySelector) (pwd string, err error) {
	if in == nil {
		return "", nil
	}
	nn := types.NamespacedName{
		Name:      in.Name,
		Namespace: in.Namespace,
	}
	s := &corev1.Secret{}
	if err := kube.Get(ctx, nn, s); err != nil {
		return "", errors.Wrap(err, errGetPasswordSecretFailed)
	}

	return string(s.Data[in.Key]), nil
}

func ExtractScopeFromSubnetID(subnetID string) string {
	if subnetID == "" {
		return ""
	}
	parts := strings.Split(subnetID, "/subnets")
	fmt.Println(len(parts))
	if len(parts) != 2 {
		return ""
	}
	return parts[0]
}

func EnsureResourceGroup(ctx context.Context, principalID, scope string, c authorization.RoleAssignmentsClient) error {
	// If scope was the empty string we probably needed a role assignment for
	// an optional scope, for example a subnetwork.
	if scope == "" {
		return nil
	}

	p := authorizationmgmt.RoleAssignmentCreateParameters{Properties: &authorizationmgmt.RoleAssignmentProperties{
		RoleDefinitionID: azure.ToStringPtr(NetworkContributorRoleID),
		PrincipalID:      azure.ToStringPtr(principalID),
	}}

	name, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	_, err = c.Create(ctx, scope, name.String(), p)
	return err
}
