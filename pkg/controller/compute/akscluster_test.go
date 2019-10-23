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
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	goerrors "github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	computev1alpha2 "github.com/crossplaneio/stack-azure/apis/compute/v1alpha2"
	azureclients "github.com/crossplaneio/stack-azure/pkg/clients"
	"github.com/crossplaneio/stack-azure/pkg/clients/compute"
)

var _ reconcile.Reconciler = &Reconciler{}

type mockAKSSetupClientFactory struct {
	mockClient *mockAKSSetupClient
}

func (m *mockAKSSetupClientFactory) CreateSetupClient(_ *azureclients.Client) (*compute.AKSSetupClient, error) {
	return &compute.AKSSetupClient{
		AKSClusterAPI:       m.mockClient,
		ApplicationAPI:      m.mockClient,
		ServicePrincipalAPI: m.mockClient,
		RoleAssignmentsAPI:  m.mockClient,
	}, nil
}

type mockAKSSetupClient struct {
	MockGet                         func(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.ManagedCluster, error)
	MockCreateOrUpdateBegin         func(ctx context.Context, instance computev1alpha2.AKSCluster, clusterName, appID, spSecret string) ([]byte, error)
	MockCreateOrUpdateEnd           func(op []byte) (bool, error)
	MockDelete                      func(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.ManagedClustersDeleteFuture, error)
	MockListClusterAdminCredentials func(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.CredentialResults, error)
	MockCreateApplication           func(ctx context.Context, appParams azureclients.ApplicationParameters) (*graphrbac.Application, error)
	MockDeleteApplication           func(ctx context.Context, appObjectID string) error
	MockCreateServicePrincipal      func(ctx context.Context, spID, appID string) (*graphrbac.ServicePrincipal, error)
	MockDeleteServicePrincipal      func(ctx context.Context, spID string) error
	MockCreateRoleAssignment        func(ctx context.Context, sp, vnetSubnetID, name string) (*authorization.RoleAssignment, error)
	MockDeleteRoleAssignment        func(ctx context.Context, vnetSubnetID, name string) error
}

func (m *mockAKSSetupClient) Get(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.ManagedCluster, error) {
	if m.MockGet != nil {
		return m.MockGet(ctx, instance)
	}
	return containerservice.ManagedCluster{}, nil
}

func (m *mockAKSSetupClient) CreateOrUpdateBegin(ctx context.Context, instance computev1alpha2.AKSCluster, clusterName, appID, spSecret string) ([]byte, error) {
	if m.MockCreateOrUpdateBegin != nil {
		return m.MockCreateOrUpdateBegin(ctx, instance, clusterName, appID, spSecret)
	}
	return nil, nil
}

func (m *mockAKSSetupClient) CreateOrUpdateEnd(op []byte) (bool, error) {
	if m.MockCreateOrUpdateEnd != nil {
		return m.MockCreateOrUpdateEnd(op)
	}
	return true, nil
}

func (m *mockAKSSetupClient) Delete(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.ManagedClustersDeleteFuture, error) {
	if m.MockDelete != nil {
		return m.MockDelete(ctx, instance)
	}
	return containerservice.ManagedClustersDeleteFuture{}, nil
}

func (m *mockAKSSetupClient) ListClusterAdminCredentials(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.CredentialResults, error) {
	if m.MockListClusterAdminCredentials != nil {
		return m.MockListClusterAdminCredentials(ctx, instance)
	}
	return containerservice.CredentialResults{}, nil
}

func (m *mockAKSSetupClient) CreateApplication(ctx context.Context, appParams azureclients.ApplicationParameters) (*graphrbac.Application, error) {
	if m.MockCreateApplication != nil {
		return m.MockCreateApplication(ctx, appParams)
	}
	return nil, nil
}

func (m *mockAKSSetupClient) DeleteApplication(ctx context.Context, appObjectID string) error {
	if m.MockDeleteApplication != nil {
		return m.MockDeleteApplication(ctx, appObjectID)
	}
	return nil
}

func (m *mockAKSSetupClient) CreateServicePrincipal(ctx context.Context, spID, appID string) (*graphrbac.ServicePrincipal, error) {
	if m.MockCreateServicePrincipal != nil {
		return m.MockCreateServicePrincipal(ctx, spID, appID)
	}
	return nil, nil
}

func (m *mockAKSSetupClient) DeleteServicePrincipal(ctx context.Context, spID string) error {
	if m.MockDeleteServicePrincipal != nil {
		return m.MockDeleteServicePrincipal(ctx, spID)
	}
	return nil
}

func (m *mockAKSSetupClient) CreateRoleAssignment(ctx context.Context, sp, vnetSubnetID, name string) (result *authorization.RoleAssignment, err error) {
	if m.MockCreateRoleAssignment != nil {
		return m.MockCreateRoleAssignment(ctx, sp, vnetSubnetID, name)
	}
	return nil, goerrors.New("subnet ID was provided but no role assignment creator")
}

func (m *mockAKSSetupClient) DeleteRoleAssignment(ctx context.Context, vnetSubnetID, name string) error {
	if m.MockDeleteRoleAssignment != nil {
		return m.MockDeleteRoleAssignment(ctx, vnetSubnetID, name)
	}
	return nil
}

func TestReconcile(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockAKSSetupClient := &mockAKSSetupClient{}
	mockAKSSetupClientFactory := &mockAKSSetupClientFactory{mockClient: mockAKSSetupClient}

	// setup all the mocked functions for the AKS setup client
	mockAKSSetupClient.MockCreateApplication = func(ctx context.Context, appParams azureclients.ApplicationParameters) (*graphrbac.Application, error) {
		return &graphrbac.Application{
			ObjectID: to.StringPtr("182f8c4a-ad89-4b25-b947-d4026ab183a1"),
			AppID:    to.StringPtr("e163d435-00d2-4ea8-9735-b875990e453e"),
		}, nil
	}
	mockAKSSetupClient.MockCreateServicePrincipal = func(ctx context.Context, spID, appID string) (*graphrbac.ServicePrincipal, error) {
		return &graphrbac.ServicePrincipal{
			ObjectID: to.StringPtr("da804153-3faa-4c73-9fcb-0961387a31f9"),
		}, nil
	}
	mockAKSSetupClient.MockCreateOrUpdateBegin = func(ctx context.Context, instance computev1alpha2.AKSCluster, clusterName, appID, spSecret string) ([]byte, error) {
		return []byte("mocked marshalled create future"), nil
	}
	mockAKSSetupClient.MockGet = func(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.ManagedCluster, error) {
		return containerservice.ManagedCluster{
			ID: to.StringPtr("fcb4e97a-c3ea-4466-9b02-e728d8e6764f"),
			ManagedClusterProperties: &containerservice.ManagedClusterProperties{
				ProvisioningState: to.StringPtr("Succeeded"),
				Fqdn:              to.StringPtr("crossplane-aks.foo.azure.com"),
			},
		}, nil
	}
	mockAKSSetupClient.MockListClusterAdminCredentials = func(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.CredentialResults, error) {
		return containerservice.CredentialResults{
			Kubeconfigs: &[]containerservice.CredentialResult{{Value: &kubecfg}},
		}, nil
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: ":8081"})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	r := NewAKSClusterReconciler(mgr, mockAKSSetupClientFactory)
	r.newClientFn = func(_ []byte) (*azureclients.Client, error) { return nil, nil }
	recFn, requests := SetupTestReconcile(r)
	controller := &AKSClusterController{
		Reconciler: recFn,
	}
	g.Expect(controller.SetupWithManager(mgr)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mgr, g))

	// create the provider object and defer its cleanup
	providerSecret := testProviderSecret([]byte(""))
	err = c.Create(ctx, providerSecret)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(ctx, providerSecret)

	provider := testProvider()
	err = c.Create(ctx, provider)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(ctx, provider)

	// Create the AKS cluster object and defer its clean up
	instance := testInstance(provider)
	err = c.Create(ctx, instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer cleanupAKSCluster(t, g, c, requests, instance)

	// first reconcile loop should start the create operation
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	// after the first reconcile, the create operation should be saved on the running operation field,
	// and the following should be set:
	// 1) cluster name
	// 2) application object ID
	// 3) service principal ID
	// 4) "creating" condition
	expectedStatus := computev1alpha2.AKSClusterStatus{
		RunningOperation:    "mocked marshalled create future",
		ClusterName:         instanceName,
		ApplicationObjectID: "182f8c4a-ad89-4b25-b947-d4026ab183a1",
		ServicePrincipalID:  "da804153-3faa-4c73-9fcb-0961387a31f9",
	}
	expectedStatus.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileSuccess())
	assertAKSClusterStatus(g, c, expectedStatus)

	// the service principal secret (note this is not the connection secret) should have been created
	spSecret := &v1.Secret{}
	n := types.NamespacedName{
		Namespace: instance.Spec.WriteServicePrincipalSecretTo.Namespace,
		Name:      instance.Spec.WriteServicePrincipalSecretTo.Name,
	}
	err = r.Get(ctx, n, spSecret)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	spSecretValue, ok := spSecret.Data[spSecretKey]
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(spSecretValue).ToNot(gomega.BeEmpty())

	// second reconcile should finish the create operation and clear out the running operation field
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	expectedStatus = computev1alpha2.AKSClusterStatus{
		RunningOperation:    "",
		ClusterName:         instanceName,
		ApplicationObjectID: "182f8c4a-ad89-4b25-b947-d4026ab183a1",
		ServicePrincipalID:  "da804153-3faa-4c73-9fcb-0961387a31f9",
	}
	expectedStatus.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileSuccess())
	assertAKSClusterStatus(g, c, expectedStatus)

	// third reconcile should find the AKS cluster instance from Azure and update the full status of the CRD
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	// verify that the CRD status was updated with details about the external AKS cluster and that the
	// CRD conditions show the transition from creating to running
	expectedStatus = computev1alpha2.AKSClusterStatus{
		ClusterName:         instanceName,
		State:               "Succeeded",
		ProviderID:          "fcb4e97a-c3ea-4466-9b02-e728d8e6764f",
		Endpoint:            "crossplane-aks.foo.azure.com",
		ApplicationObjectID: "182f8c4a-ad89-4b25-b947-d4026ab183a1",
		ServicePrincipalID:  "da804153-3faa-4c73-9fcb-0961387a31f9",
	}
	expectedStatus.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	assertAKSClusterStatus(g, c, expectedStatus)

	// verify that a finalizer was added to the CRD
	c.Get(ctx, expectedRequest.NamespacedName, instance)
	g.Expect(len(instance.Finalizers)).To(gomega.Equal(1))
	g.Expect(instance.Finalizers[0]).To(gomega.Equal(finalizer))
}

func TestReconcileInSubnet(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	mockAKSSetupClient := &mockAKSSetupClient{}
	mockAKSSetupClientFactory := &mockAKSSetupClientFactory{mockClient: mockAKSSetupClient}

	// setup all the mocked functions for the AKS setup client
	mockAKSSetupClient.MockCreateApplication = func(ctx context.Context, appParams azureclients.ApplicationParameters) (*graphrbac.Application, error) {
		return &graphrbac.Application{
			ObjectID: to.StringPtr("182f8c4a-ad89-4b25-b947-d4026ab183a1"),
			AppID:    to.StringPtr("e163d435-00d2-4ea8-9735-b875990e453e"),
		}, nil
	}
	mockAKSSetupClient.MockCreateServicePrincipal = func(ctx context.Context, spID, appID string) (*graphrbac.ServicePrincipal, error) {
		return &graphrbac.ServicePrincipal{
			ObjectID: to.StringPtr("da804153-3faa-4c73-9fcb-0961387a31f9"),
		}, nil
	}
	mockAKSSetupClient.MockCreateRoleAssignment = func(ctx context.Context, sp, vnetSubnetID, name string) (result *authorization.RoleAssignment, err error) {
		return &authorization.RoleAssignment{}, nil
	}
	mockAKSSetupClient.MockCreateOrUpdateBegin = func(ctx context.Context, instance computev1alpha2.AKSCluster, clusterName, appID, spSecret string) ([]byte, error) {
		return []byte("mocked marshalled create future"), nil
	}
	mockAKSSetupClient.MockGet = func(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.ManagedCluster, error) {
		return containerservice.ManagedCluster{
			ID: to.StringPtr("fcb4e97a-c3ea-4466-9b02-e728d8e6764f"),
			ManagedClusterProperties: &containerservice.ManagedClusterProperties{
				ProvisioningState: to.StringPtr("Succeeded"),
				Fqdn:              to.StringPtr("crossplane-aks.foo.azure.com"),
			},
		}, nil
	}
	mockAKSSetupClient.MockListClusterAdminCredentials = func(ctx context.Context, instance computev1alpha2.AKSCluster) (containerservice.CredentialResults, error) {
		return containerservice.CredentialResults{
			Kubeconfigs: &[]containerservice.CredentialResult{{Value: &kubecfg}},
		}, nil
	}

	// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
	// channel when it is finished.
	mgr, err := manager.New(cfg, manager.Options{MetricsBindAddress: ":8081"})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	c := mgr.GetClient()

	r := NewAKSClusterReconciler(mgr, mockAKSSetupClientFactory)
	r.newClientFn = func(_ []byte) (*azureclients.Client, error) { return nil, nil }
	recFn, requests := SetupTestReconcile(r)
	controller := &AKSClusterController{
		Reconciler: recFn,
	}
	g.Expect(controller.SetupWithManager(mgr)).NotTo(gomega.HaveOccurred())
	defer close(StartTestManager(mgr, g))

	// create the provider object and defer its cleanup
	providerSecret := testProviderSecret([]byte(""))
	err = c.Create(ctx, providerSecret)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(ctx, providerSecret)

	provider := testProvider()
	err = c.Create(ctx, provider)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer c.Delete(ctx, provider)

	// Create the AKS cluster object and defer its clean up
	instance := testInstanceInSubnet(provider)
	err = c.Create(ctx, instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	defer cleanupAKSCluster(t, g, c, requests, instance)

	// first reconcile loop should start the create operation
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	// after the first reconcile, the create operation should be saved on the running operation field,
	// and the following should be set:
	// 1) cluster name
	// 2) application object ID
	// 3) service principal ID
	// 4) "creating" condition
	expectedStatus := computev1alpha2.AKSClusterStatus{
		RunningOperation:    "mocked marshalled create future",
		ClusterName:         instanceName,
		ApplicationObjectID: "182f8c4a-ad89-4b25-b947-d4026ab183a1",
		ServicePrincipalID:  "da804153-3faa-4c73-9fcb-0961387a31f9",
	}
	expectedStatus.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileSuccess())
	assertAKSClusterStatus(g, c, expectedStatus)

	// the service principal secret (note this is not the connection secret) should have been created
	spSecret := &v1.Secret{}
	n := types.NamespacedName{
		Namespace: instance.Spec.WriteServicePrincipalSecretTo.Namespace,
		Name:      instance.Spec.WriteServicePrincipalSecretTo.Name,
	}
	err = r.Get(ctx, n, spSecret)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	spSecretValue, ok := spSecret.Data[spSecretKey]
	g.Expect(ok).To(gomega.BeTrue())
	g.Expect(spSecretValue).ToNot(gomega.BeEmpty())

	// second reconcile should finish the create operation and clear out the running operation field
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))
	expectedStatus = computev1alpha2.AKSClusterStatus{
		RunningOperation:    "",
		ClusterName:         instanceName,
		ApplicationObjectID: "182f8c4a-ad89-4b25-b947-d4026ab183a1",
		ServicePrincipalID:  "da804153-3faa-4c73-9fcb-0961387a31f9",
	}
	expectedStatus.SetConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileSuccess())
	assertAKSClusterStatus(g, c, expectedStatus)

	// third reconcile should find the AKS cluster instance from Azure and update the full status of the CRD
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	// verify that the CRD status was updated with details about the external AKS cluster and that the
	// CRD conditions show the transition from creating to running
	expectedStatus = computev1alpha2.AKSClusterStatus{
		ClusterName:         instanceName,
		State:               "Succeeded",
		ProviderID:          "fcb4e97a-c3ea-4466-9b02-e728d8e6764f",
		Endpoint:            "crossplane-aks.foo.azure.com",
		ApplicationObjectID: "182f8c4a-ad89-4b25-b947-d4026ab183a1",
		ServicePrincipalID:  "da804153-3faa-4c73-9fcb-0961387a31f9",
	}
	expectedStatus.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	assertAKSClusterStatus(g, c, expectedStatus)

	// verify that a finalizer was added to the CRD
	c.Get(ctx, expectedRequest.NamespacedName, instance)
	g.Expect(len(instance.Finalizers)).To(gomega.Equal(1))
	g.Expect(instance.Finalizers[0]).To(gomega.Equal(finalizer))
}

func cleanupAKSCluster(t *testing.T, g *gomega.GomegaWithT, c client.Client, requests chan reconcile.Request, instance *computev1alpha2.AKSCluster) {
	deletedInstance := &computev1alpha2.AKSCluster{}
	if err := c.Get(ctx, expectedRequest.NamespacedName, deletedInstance); errors.IsNotFound(err) {
		// instance has already been deleted, bail out
		return
	}

	t.Logf("cleaning up AKS cluster instance %s by deleting it", instance.Name)
	err := c.Delete(ctx, instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// wait for the reconcile to happen that handles the CRD deletion
	g.Eventually(requests, timeout).Should(gomega.Receive(gomega.Equal(expectedRequest)))

	// wait for the finalizer to run and the instance to be deleted for good
	err = wait.ExponentialBackoff(test.DefaultRetry, func() (done bool, err error) {
		deletedInstance := &computev1alpha2.AKSCluster{}
		if err := c.Get(ctx, expectedRequest.NamespacedName, deletedInstance); errors.IsNotFound(err) {
			return true, nil
		}
		return false, nil
	})
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

func assertAKSClusterStatus(g *gomega.GomegaWithT, c client.Client, expectedStatus computev1alpha2.AKSClusterStatus) {
	instance := &computev1alpha2.AKSCluster{}
	err := c.Get(ctx, expectedRequest.NamespacedName, instance)
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// assert the expected status properties
	g.Expect(instance.Status.ClusterName).To(gomega.HavePrefix(expectedStatus.ClusterName))
	g.Expect(instance.Status.State).To(gomega.Equal(expectedStatus.State))
	g.Expect(instance.Status.ProviderID).To(gomega.Equal(expectedStatus.ProviderID))
	g.Expect(instance.Status.Endpoint).To(gomega.Equal(expectedStatus.Endpoint))
	g.Expect(instance.Status.ApplicationObjectID).To(gomega.Equal(expectedStatus.ApplicationObjectID))
	g.Expect(instance.Status.ServicePrincipalID).To(gomega.Equal(expectedStatus.ServicePrincipalID))
	g.Expect(instance.Status.RunningOperation).To(gomega.Equal(expectedStatus.RunningOperation))
	g.Expect(cmp.Diff(expectedStatus.ConditionedStatus, instance.Status.ConditionedStatus, test.EquateConditions())).Should(gomega.BeZero())
}
