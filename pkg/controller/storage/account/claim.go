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

package account

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	storagev1alpha1 "github.com/crossplaneio/crossplane/apis/storage/v1alpha1"

	"github.com/crossplaneio/stack-azure/apis/storage/v1alpha3"
)

// A ClaimSchedulingController reconciles Bucket claims that include a class
// selector but omit their class and resource references by picking a random
// matching Azure Account class, if any.
type ClaimSchedulingController struct{}

// SetupWithManager sets up the ClaimSchedulingController using the supplied
// manager.
func (c *ClaimSchedulingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("scheduler.%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha3.AccountKind,
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
			resource.ClassKind(v1alpha3.AccountClassGroupVersionKind),
		))
}

// A ClaimDefaultingController reconciles Bucket claims that omit their resource
// ref, class ref, and class selector by choosing a default Azure Account
// resource class if one exists.
type ClaimDefaultingController struct{}

// SetupWithManager sets up the ClaimDefaultingController using the supplied
// manager.
func (c *ClaimDefaultingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("defaulter.%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha3.AccountKind,
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
			resource.ClassKind(v1alpha3.AccountClassGroupVersionKind),
		))
}

// A ClaimController reconciles Bucket claims with Azure Account resources,
// dynamically provisioning them if needed.
type ClaimController struct{}

// SetupWithManager sets up the ClaimController using the supplied manager.
func (c *ClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha3.AccountKind,
		v1alpha3.Group))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
		resource.ClassKind(v1alpha3.AccountClassGroupVersionKind),
		resource.ManagedKind(v1alpha3.AccountGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigureAccount),
			resource.ManagedConfiguratorFn(resource.ConfigureReclaimPolicy),
		))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha3.AccountClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha3.AccountGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha3.AccountGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha3.Account{}}, &resource.EnqueueRequestForClaim{}).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureAccount configures the supplied resource (presumed to be an Account)
// using the supplied resource claim (presumed to be a Bucket) and resource class.
func ConfigureAccount(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	b, cmok := cm.(*storagev1alpha1.Bucket)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), storagev1alpha1.BucketGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha3.AccountClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha3.AccountClassGroupVersionKind)
	}

	a, mgok := mg.(*v1alpha3.Account)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha3.AccountGroupVersionKind)
	}

	if b.Spec.Name == "" {
		return errors.Errorf("invalid account claim: %s spec, name property is required", b.GetName())
	}

	spec := &v1alpha3.AccountSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		AccountParameters: rs.SpecTemplate.AccountParameters,
	}

	// NOTE(hasheddan): consider moving defaulting to either CRD or managed reconciler level
	if spec.StorageAccountSpec == nil {
		spec.StorageAccountSpec = &v1alpha3.StorageAccountSpec{}
	}

	spec.StorageAccountName = b.Spec.Name

	spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{
		Namespace: rs.SpecTemplate.WriteConnectionSecretsToNamespace,
		Name:      string(cm.GetUID()),
	}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	a.Spec = *spec

	// Accounts do not follow the typical pattern of creating a managed resource
	// named claimkind-claimuuid because their associated container needs a
	// predictably named account from which to load its connection secret.
	// Instead we create an account with the same name as the claim.
	a.SetNamespace(cs.GetNamespace())
	a.SetName(b.GetName())

	return nil
}
