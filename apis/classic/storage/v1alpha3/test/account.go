/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance With the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package test

import (
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	storagev1alpha3 "github.com/crossplane/provider-azure/apis/classic/storage/v1alpha3"
)

// MockAccount builder for testing account object
type MockAccount struct {
	*storagev1alpha3.Account
}

// NewMockAccount creates new account wrapper
func NewMockAccount(name string) *MockAccount {
	a := &MockAccount{Account: &storagev1alpha3.Account{
		ObjectMeta: metav1.ObjectMeta{Name: name},
	}}
	meta.SetExternalName(a, name)
	return a
}

// WithTypeMeta sets TypeMeta value
func (ta *MockAccount) WithTypeMeta(tm metav1.TypeMeta) *MockAccount {
	ta.TypeMeta = tm
	return ta
}

// WithObjectMeta sets ObjectMeta value
func (ta *MockAccount) WithObjectMeta(om metav1.ObjectMeta) *MockAccount {
	ta.ObjectMeta = om
	return ta
}

// WithUID sets UID value
func (ta *MockAccount) WithUID(uid string) *MockAccount {
	ta.ObjectMeta.UID = types.UID(uid)
	return ta
}

// WithDeleteTimestamp sets metadata deletion timestamp
func (ta *MockAccount) WithDeleteTimestamp(t metav1.Time) *MockAccount {
	ta.Account.ObjectMeta.DeletionTimestamp = &t
	return ta
}

// WithFinalizer sets finalizer
func (ta *MockAccount) WithFinalizer(f string) *MockAccount {
	ta.Account.ObjectMeta.Finalizers = append(ta.Account.ObjectMeta.Finalizers, f)
	return ta
}

// WithFinalizers sets finalizers list
func (ta *MockAccount) WithFinalizers(f []string) *MockAccount {
	ta.Account.ObjectMeta.Finalizers = f
	return ta
}

// WithSpecProvider set a provider
func (ta *MockAccount) WithSpecProvider(name string) *MockAccount {
	ta.Spec.ProviderReference = &xpv1.Reference{Name: name}
	return ta
}

// WithSpecDeletionPolicy sets resource deletion policy
func (ta *MockAccount) WithSpecDeletionPolicy(policy xpv1.DeletionPolicy) *MockAccount {
	ta.Spec.DeletionPolicy = policy
	return ta
}

// WithSpecStorageAccountSpec sets storage account specs
func (ta *MockAccount) WithSpecStorageAccountSpec(spec *storagev1alpha3.StorageAccountSpec) *MockAccount {
	ta.Spec.StorageAccountSpec = spec
	return ta
}

// WithStorageAccountStatus set storage account status
func (ta *MockAccount) WithStorageAccountStatus(status *storagev1alpha3.StorageAccountStatus) *MockAccount {
	ta.Status.StorageAccountStatus = status
	return ta
}

// WithSpecStatusFromProperties set storage account spec status from storage properties
func (ta *MockAccount) WithSpecStatusFromProperties(props *storage.AccountProperties) *MockAccount {
	acct := &storage.Account{
		AccountProperties: props,
	}
	ta.WithSpecStorageAccountSpec(storagev1alpha3.NewStorageAccountSpec(acct)).
		WithStorageAccountStatus(storagev1alpha3.NewStorageAccountStatus(acct))
	return ta
}

// WithSpecWriteConnectionSecretToReference sets where the storage account will write its
// connection secret.
func (ta *MockAccount) WithSpecWriteConnectionSecretToReference(ns, name string) *MockAccount {
	ta.Spec.WriteConnectionSecretToReference = &xpv1.SecretReference{
		Namespace: ns,
		Name:      name,
	}
	return ta
}

// WithStatusConditions sets the storage account's conditioned status.
func (ta *MockAccount) WithStatusConditions(c ...xpv1.Condition) *MockAccount {
	ta.Status.SetConditions(c...)
	return ta
}
