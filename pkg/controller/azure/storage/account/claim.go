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

package account

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	storagev1alpha1 "github.com/crossplaneio/crossplane/apis/storage/v1alpha1"

	"github.com/crossplaneio/stack-azure/azure/apis/storage/v1alpha2"
)

// ClaimController is responsible for adding the Account claim controller and its
// corresponding reconciler to the manager with any runtime configuration.
type ClaimController struct{}

// SetupWithManager adds a controller that reconciles Bucket resource claims.
func (c *ClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		storagev1alpha1.BucketKind,
		v1alpha2.AccountKind,
		v1alpha2.Group))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(storagev1alpha1.BucketGroupVersionKind),
		resource.ClassKinds{
			Portable:    storagev1alpha1.BucketClassGroupVersionKind,
			NonPortable: v1alpha2.AccountClassGroupVersionKind,
		},
		resource.ManagedKind(v1alpha2.AccountGroupVersionKind),
		resource.WithManagedBinder(resource.NewAPIManagedStatusBinder(mgr.GetClient())),
		resource.WithManagedFinalizer(resource.NewAPIManagedStatusUnbinder(mgr.GetClient())),
		resource.WithManagedConfigurators(resource.ManagedConfiguratorFn(ConfigureAccount)))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha2.AccountGroupVersionKind)),
		resource.HasDirectClassReferenceKind(resource.NonPortableClassKind(v1alpha2.AccountClassGroupVersionKind)),
		resource.HasIndirectClassReferenceKind(mgr.GetClient(), mgr.GetScheme(), resource.ClassKinds{
			Portable:    storagev1alpha1.BucketClassGroupVersionKind,
			NonPortable: v1alpha2.AccountClassGroupVersionKind,
		})))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha2.Account{}}, &resource.EnqueueRequestForClaim{}).
		For(&storagev1alpha1.Bucket{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureAccount configures the supplied resource (presumed to be an Account)
// using the supplied resource claim (presumed to be a Bucket) and resource class.
func ConfigureAccount(_ context.Context, cm resource.Claim, cs resource.NonPortableClass, mg resource.Managed) error {
	b, cmok := cm.(*storagev1alpha1.Bucket)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), storagev1alpha1.BucketGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha2.AccountClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.AccountClassGroupVersionKind)
	}

	a, mgok := mg.(*v1alpha2.Account)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha2.AccountGroupVersionKind)
	}

	if b.Spec.Name == "" {
		return errors.Errorf("invalid account claim: %s spec, name property is required", b.GetName())
	}

	spec := &v1alpha2.AccountSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		AccountParameters: rs.SpecTemplate.AccountParameters,
	}

	// NOTE(hasheddan): consider moving defaulting to either CRD or managed reconciler level
	if spec.StorageAccountSpec == nil {
		spec.StorageAccountSpec = &v1alpha2.StorageAccountSpec{}
	}

	spec.StorageAccountName = b.Spec.Name

	spec.WriteConnectionSecretToReference = corev1.LocalObjectReference{Name: string(cm.GetUID())}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	a.Spec = *spec

	// Accounts do not follow the typical pattern of creating a managed resource
	// named claimkind-claimuuid because their associated container needs a
	// predictably named account from which to load its connection secret.
	// Instead we create an account with the same name as the claim.
	a.SetNamespace(cs.GetNamespace())
	a.SetName(b.GetName())

	// TODO(negz): Don't set this potentially cross-namespace owner reference.
	// We probably want to use the resource's reclaim policy, not Kubernetes
	// garbage collection, to determine whether to delete the managed resource
	// when the claim is deleted per
	// https://github.com/crossplaneio/crossplane/issues/550
	a.SetOwnerReferences([]v1.OwnerReference{meta.AsOwner(meta.ReferenceTo(b, storagev1alpha1.BucketGroupVersionKind))})

	return nil
}
