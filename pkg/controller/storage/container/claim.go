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

	"github.com/Azure/azure-storage-blob-go/azblob"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimbinding"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimdefaulting"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimscheduling"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	storagev1alpha1 "github.com/crossplane/crossplane/apis/storage/v1alpha1"

	"github.com/crossplane/provider-azure/apis/storage/v1alpha3"
)

// SetupClaimScheduling adds a controller that reconciles Bucket claims that
// include a class selector but omit their class and resource references by
// picking a random matching Azure Container class, if any.
func SetupClaimScheduling(mgr ctrl.Manager, l logging.Logger) error {
	name := claimscheduling.ControllerName(storagev1alpha1.BucketGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimscheduling.NewReconciler(mgr,
			resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
			resource.ClassKind(v1alpha3.ContainerClassGroupVersionKind),
			claimscheduling.WithLogger(l.WithValues("controller", name)),
			claimscheduling.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupClaimDefaulting adds a controller that reconciles Bucket claims that
// omit their resource ref, class ref, and class selector by choosing a default
// Azure Container resource class if one exists.
func SetupClaimDefaulting(mgr ctrl.Manager, l logging.Logger) error {
	name := claimdefaulting.ControllerName(storagev1alpha1.BucketGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimdefaulting.NewReconciler(mgr,
			resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
			resource.ClassKind(v1alpha3.ContainerClassGroupVersionKind),
			claimdefaulting.WithLogger(l.WithValues("controller", name)),
			claimdefaulting.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupClaimBinding adds a controller that reconciles Bucket claims with Azure
// Account resources, dynamically provisioning them if needed.
func SetupClaimBinding(mgr ctrl.Manager, l logging.Logger) error {
	name := claimbinding.ControllerName(storagev1alpha1.BucketGroupKind)

	r := claimbinding.NewReconciler(mgr,
		resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
		resource.ClassKind(v1alpha3.ContainerClassGroupVersionKind),
		resource.ManagedKind(v1alpha3.ContainerGroupVersionKind),
		claimbinding.WithManagedConfigurators(
			claimbinding.ManagedConfiguratorFn(ConfigureContainer),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureReclaimPolicy),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureNames)),
		claimbinding.WithLogger(l.WithValues("controller", name)),
		claimbinding.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

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
		ContainerParameters: rs.SpecTemplate.ContainerParameters,
	}

	// NOTE(hasheddan): consider moving defaulting to either CRD or managed reconciler level
	if spec.Metadata == nil {
		spec.Metadata = azblob.Metadata{}
	}

	spec.ProviderReference = rs.SpecTemplate.ProviderReference.DeepCopy()
	a.Spec = *spec

	return nil
}
