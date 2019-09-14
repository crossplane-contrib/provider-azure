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

package container

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	storagev1alpha1 "github.com/crossplaneio/crossplane/apis/storage/v1alpha1"

	"github.com/crossplaneio/stack-azure/azure/apis/storage/v1alpha2"
)

// ClaimController is responsible for adding the Container claim controller and its
// corresponding reconciler to the manager with any runtime configuration.
type ClaimController struct{}

// SetupWithManager adds a controller that reconciles Bucket resource claims.
func (c *ClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha2.ContainerKind,
		v1alpha2.Group))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
		resource.ClassKinds{
			Portable:    storagev1alpha1.BucketClassGroupVersionKind,
			NonPortable: v1alpha2.ContainerClassGroupVersionKind,
		},
		resource.ManagedKind(v1alpha2.ContainerGroupVersionKind),
		resource.WithManagedBinder(resource.NewAPIManagedStatusBinder(mgr.GetClient())),
		resource.WithManagedFinalizer(resource.NewAPIManagedStatusUnbinder(mgr.GetClient())),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigureContainer),
			resource.NewObjectMetaConfigurator(mgr.GetScheme()),
		))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha2.ContainerGroupVersionKind)),
		resource.HasDirectClassReferenceKind(resource.NonPortableClassKind(v1alpha2.ContainerClassGroupVersionKind)),
		resource.HasIndirectClassReferenceKind(mgr.GetClient(), mgr.GetScheme(), resource.ClassKinds{
			Portable:    storagev1alpha1.BucketClassGroupVersionKind,
			NonPortable: v1alpha2.ContainerClassGroupVersionKind,
		})))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha2.Container{}}, &resource.EnqueueRequestForClaim{}).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureContainer configures the supplied resource (presumed to be an Container)
// using the supplied resource claim (presumed to be a Bucket) and resource class.
func ConfigureContainer(_ context.Context, cm resource.Claim, cs resource.NonPortableClass, mg resource.Managed) error {
	if _, cmok := cm.(*storagev1alpha1.Bucket); !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), storagev1alpha1.BucketGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha2.ContainerClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.ContainerClassGroupVersionKind)
	}

	a, mgok := mg.(*v1alpha2.Container)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha2.ContainerGroupVersionKind)
	}

	spec := &v1alpha2.ContainerSpec{
		ReclaimPolicy:       runtimev1alpha1.ReclaimRetain,
		ContainerParameters: rs.SpecTemplate.ContainerParameters,
	}

	// NOTE(hasheddan): consider moving defaulting to either CRD or managed reconciler level
	if spec.Metadata == nil {
		spec.Metadata = azblob.Metadata{}
	}

	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	// Azure storage containers read credentials via an Account resource, not an
	// Azure Crossplane provider. We reuse the 'provider' reference field of the
	// resource class.
	spec.AccountReference = corev1.LocalObjectReference{Name: rs.SpecTemplate.ProviderReference.Name}

	a.Spec = *spec

	return nil
}
