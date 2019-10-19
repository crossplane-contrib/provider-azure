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
	"encoding/base64"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/crossplaneio/stack-azure/apis"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	localtest "github.com/crossplaneio/stack-azure/pkg/test"

	"github.com/crossplaneio/stack-azure/apis/compute/v1alpha2"
	computev1alpha1 "github.com/crossplaneio/stack-azure/apis/compute/v1alpha2"
	azurev1alpha2 "github.com/crossplaneio/stack-azure/apis/v1alpha2"
)

const (
	timeout      = 5 * time.Second
	namespace    = "test-compute-namespace"
	instanceName = "test-compute-instance"

	providerName          = "test-provider"
	providerSecretName    = "test-provider-secret"
	providerSecretDataKey = "credentials"

	connectionSecretName = "test-connection-secret"
	principalSecretName  = "test-principal-secret"

	clientEndpoint = "https://example.org"
	clientCAdata   = "DEFINITELYPEMENCODED"
	clientCert     = "SOMUCHPEM"
	clientKey      = "WOWVERYENCODED"
)

const kubeconfigTemplate = `
---
apiVersion: v1
kind: Config
contexts:
- context:
    cluster: aks
    user: aks
  name: %s
clusters:
- cluster:
    server: %s
    certificate-authority-data: %s
  name: aks
users:
- name: aks
  user:
    client-certificate-data: %s
    client-key-data: %s
current-context: aks
preferences: {}
`

var (
	cfg             *rest.Config
	expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: instanceName}}

	kubecfg = []byte(fmt.Sprintf(kubeconfigTemplate,
		instanceName,
		clientEndpoint,
		base64.StdEncoding.EncodeToString([]byte(clientCAdata)),
		base64.StdEncoding.EncodeToString([]byte(clientCert)),
		base64.StdEncoding.EncodeToString([]byte(clientKey)),
	))
)

func TestMain(m *testing.M) {
	t := test.NewEnv(namespace, apis.AddToSchemes, localtest.CRDs())
	cfg = t.Start()
	t.StopAndExit(m.Run())
}

// SetupTestReconcile returns a reconcile.Reconcile implementation that delegates to inner and
// writes the request to requests after Reconcile is finished.
func SetupTestReconcile(inner reconcile.Reconciler) (reconcile.Reconciler, chan reconcile.Request) {
	requests := make(chan reconcile.Request)
	fn := reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) {
		result, err := inner.Reconcile(req)
		requests <- req
		return result, err
	})
	return fn, requests
}

// StartTestManager adds recFn
func StartTestManager(mgr manager.Manager, g *gomega.GomegaWithT) chan struct{} {
	stop := make(chan struct{})
	go func() {
		g.Expect(mgr.Start(stop)).NotTo(gomega.HaveOccurred())
	}()
	return stop
}

func testProviderSecret(data []byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      providerSecretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			providerSecretDataKey: data,
		},
	}
}

func testProvider() *azurev1alpha2.Provider {
	return &azurev1alpha2.Provider{
		ObjectMeta: metav1.ObjectMeta{
			Name: providerName,
		},
		Spec: azurev1alpha2.ProviderSpec{
			Secret: runtimev1alpha1.SecretKeySelector{
				SecretReference: runtimev1alpha1.SecretReference{
					Namespace: namespace,
					Name:      providerSecretName,
				},
				Key: providerSecretDataKey,
			},
		},
	}
}

func testInstance(p *azurev1alpha2.Provider) *computev1alpha1.AKSCluster {
	return &computev1alpha1.AKSCluster{
		ObjectMeta: metav1.ObjectMeta{Name: instanceName},
		Spec: computev1alpha1.AKSClusterSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ReclaimPolicy:     runtimev1alpha1.ReclaimDelete,
				ProviderReference: meta.ReferenceTo(p, azurev1alpha2.ProviderGroupVersionKind),
				WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
					Namespace: namespace,

					// NOTE(negz): There appears to be a race in these tests
					// in which garbage collection has not collected the first
					// collection secret before the next test runs. This
					// prevents two clusters using the same secret.
					Name: connectionSecretName + strconv.Itoa(rand.Int()),
				},
			},
			AKSClusterParameters: v1alpha2.AKSClusterParameters{
				WriteServicePrincipalSecretTo: runtimev1alpha1.SecretReference{
					Namespace: namespace,
					Name:      principalSecretName,
				},
				ResourceGroupName: "rg1",
				Location:          "loc1",
				Version:           "1.12.5",
				NodeCount:         to.IntPtr(3),
				NodeVMSize:        "Standard_F2s_v2",
				DNSNamePrefix:     "crossplane-aks",
				DisableRBAC:       false,
			},
		},
	}
}

func testInstanceInSubnet(p *azurev1alpha2.Provider) *computev1alpha1.AKSCluster {
	instance := testInstance(p)
	instance.Spec.AKSClusterParameters.VnetSubnetID = "/path/to/cool/subnet"
	return instance
}
