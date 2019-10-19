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

package database

import (
	"testing"
	"time"

	"github.com/crossplaneio/stack-azure/apis"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	"github.com/crossplaneio/stack-azure/apis/database/v1alpha2"
	azuredbv1alpha2 "github.com/crossplaneio/stack-azure/apis/database/v1alpha2"
	azurev1alpha2 "github.com/crossplaneio/stack-azure/apis/v1alpha2"
	localtest "github.com/crossplaneio/stack-azure/pkg/test"
)

const (
	timeout       = 5 * time.Second
	namespace     = "test-namespace"
	instanceName  = "test-db-instance"
	secretName    = "test-secret"
	secretDataKey = "credentials"
	providerName  = "test-provider"
)

var (
	cfg             *rest.Config
	expectedRequest = reconcile.Request{NamespacedName: types.NamespacedName{Name: instanceName}}
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

func testSecret(data []byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Data: map[string][]byte{
			secretDataKey: data,
		},
	}
}

func testProvider(s *corev1.Secret) *azurev1alpha2.Provider {
	return &azurev1alpha2.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: providerName},
		Spec: azurev1alpha2.ProviderSpec{
			Secret: runtimev1alpha1.SecretKeySelector{
				SecretReference: runtimev1alpha1.SecretReference{
					Namespace: s.GetNamespace(),
					Name:      s.GetName(),
				},
				Key: secretDataKey,
			},
		},
	}
}

func testInstance(p *azurev1alpha2.Provider) *azuredbv1alpha2.MysqlServer {
	return &azuredbv1alpha2.MysqlServer{
		ObjectMeta: metav1.ObjectMeta{Name: instanceName},
		Spec: azuredbv1alpha2.SQLServerSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: p.GetName()},
				WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
					Namespace: namespace,
					Name:      "coolsecret",
				},
			},
			SQLServerParameters: v1alpha2.SQLServerParameters{
				AdminLoginName: "myadmin",
				PricingTier: azuredbv1alpha2.PricingTierSpec{
					Tier: "Basic", VCores: 1, Family: "Gen4",
				},
			},
		},
	}
}
