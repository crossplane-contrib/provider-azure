/*
Copyright 2020 The Crossplane Authors.

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

	"github.com/crossplaneio/stack-azure/apis/storage/v1alpha3"
)

// A ClaimSchedulingController reconciles Bucket claims that include a class
// selector but omit their class and resource references by picking a random
// matching Azure Container class, if any.
type ClaimSchedulingController struct{}

// SetupWithManager sets up the ClaimSchedulingController using the supplied
// manager.
func (c *ClaimSchedulingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("scheduler.%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha3.ContainerKind,
		v1alpha3.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimSchedulingReconciler(mgr,
			resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
			resource.ClassKind(v1alpha3.ContainerClassGroupVersionKind),
		))
}

// A ClaimDefaultingController reconciles Bucket claims that omit their resource
// ref, class ref, and class selector by choosing a default Azure Container
// resource class if one exists.
type ClaimDefaultingController struct{}

// SetupWithManager sets up the ClaimDefaultingController using the supplied
// manager.
func (c *ClaimDefaultingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("defaulter.%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha3.ContainerKind,
		v1alpha3.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimDefaultingReconciler(mgr,
			resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
			resource.ClassKind(v1alpha3.ContainerClassGroupVersionKind),
		))
}

// A ClaimController reconciles Bucket claims with Azure Account resources,
// dynamically provisioning them if needed.
type ClaimController struct{}

// SetupWithManager sets up the ClaimController using the supplied manager.
func (c *ClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha3.ContainerKind,
		v1alpha3.Group))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
		resource.ClassKind(v1alpha3.ContainerClassGroupVersionKind),
		resource.ManagedKind(v1alpha3.ContainerGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigureContainer),
			resource.ManagedConfiguratorFn(resource.ConfigureReclaimPolicy),
			resource.ManagedConfiguratorFn(resource.ConfigureNames),
		))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha3.ContainerClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha3.ContainerGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha3.ContainerGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha3.Container{}}, &resource.EnqueueRequestForClaim{}).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureContainer configures the supplied resource (presumed to be an Container)
// using the supplied resource claim (presumed to be a Bucket) and resource class.
func ConfigureContainer(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	if _, cmok := cm.(*storagev1alpha1.Bucket); !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), storagev1alpha1.BucketGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha3.ContainerClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha3.ContainerClassGroupVersionKind)
	}

	a, mgok := mg.(*v1alpha3.Container)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha3.ContainerGroupVersionKind)
	}

	spec := &v1alpha3.ContainerSpec{
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
