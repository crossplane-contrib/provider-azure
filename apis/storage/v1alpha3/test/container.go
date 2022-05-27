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
	"time"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	storagev1alpha3 "github.com/crossplane-contrib/provider-azure/apis/storage/v1alpha3"

	"github.com/Azure/azure-storage-blob-go/azblob"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// MockContainer builder to create a continer object for testing
type MockContainer struct {
	*storagev1alpha3.Container
}

// NewMockContainer new container builcer
func NewMockContainer(name string) *MockContainer {
	c := &MockContainer{
		Container: &storagev1alpha3.Container{
			ObjectMeta: metav1.ObjectMeta{Name: name},
		},
	}
	meta.SetExternalName(c, name)
	return c
}

// WithResourceVersion sets ResourceVersion value
func (tc *MockContainer) WithResourceVersion(v string) *MockContainer {
	tc.ObjectMeta.ResourceVersion = v
	return tc
}

// WithTypeMeta sets TypeMeta value
func (tc *MockContainer) WithTypeMeta(tm metav1.TypeMeta) *MockContainer {
	tc.TypeMeta = tm
	return tc
}

// WithObjectMeta sets ObjectMeta value
func (tc *MockContainer) WithObjectMeta(om metav1.ObjectMeta) *MockContainer {
	tc.ObjectMeta = om
	return tc
}

// WithUID sets UID value
func (tc *MockContainer) WithUID(uid string) *MockContainer {
	tc.ObjectMeta.UID = types.UID(uid)
	return tc
}

// WithDeleteTimestamp sets deletion timestamp value
func (tc *MockContainer) WithDeleteTimestamp(t time.Time) *MockContainer {
	tc.Container.ObjectMeta.DeletionTimestamp = &metav1.Time{Time: t}
	return tc
}

// WithFinalizer sets finalizer
func (tc *MockContainer) WithFinalizer(f string) *MockContainer {
	tc.Container.ObjectMeta.Finalizers = append(tc.Container.ObjectMeta.Finalizers, f)
	return tc
}

// WithFinalizers sets finalizers list
func (tc *MockContainer) WithFinalizers(f []string) *MockContainer {
	tc.Container.ObjectMeta.Finalizers = f
	return tc
}

// WithSpecProviderRef sets spec account reference value
func (tc *MockContainer) WithSpecProviderRef(name string) *MockContainer {
	tc.Container.Spec.ProviderReference = &xpv1.Reference{Name: name}
	return tc
}

// WithSpecProviderConfigRef sets spec account reference value
func (tc *MockContainer) WithSpecProviderConfigRef(name string) *MockContainer {
	tc.Container.Spec.ProviderConfigReference = &xpv1.Reference{Name: name}
	return tc
}

// WithSpecDeletionPolicy sets spec deletion policy value
func (tc *MockContainer) WithSpecDeletionPolicy(p xpv1.DeletionPolicy) *MockContainer {
	tc.Container.Spec.DeletionPolicy = p
	return tc
}

// WithSpecPAC sets spec public access type value
func (tc *MockContainer) WithSpecPAC(pac azblob.PublicAccessType) *MockContainer {
	tc.Container.Spec.PublicAccessType = pac
	return tc
}

// WithSpecMetadata sets spec metadata value
func (tc *MockContainer) WithSpecMetadata(meta map[string]string) *MockContainer {
	tc.Container.Spec.Metadata = meta
	return tc
}

// WithStatusConditions sets the conditioned status.
func (tc *MockContainer) WithStatusConditions(c ...xpv1.Condition) *MockContainer {
	tc.Status.SetConditions(c...)
	return tc
}
