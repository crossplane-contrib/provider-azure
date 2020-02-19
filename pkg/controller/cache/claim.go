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

package cache

import (
	"context"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimbinding"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimdefaulting"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/claimscheduling"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cachev1alpha1 "github.com/crossplane/crossplane/apis/cache/v1alpha1"

	"github.com/crossplane/stack-azure/apis/cache/v1beta1"
)

// SetupRedisClaimScheduling adds a controller that reconciles RedisCluster
// claims that include a class selector but omit their class and resource
// references by picking a random matching Azure Redis class, if any.
func SetupRedisClaimScheduling(mgr ctrl.Manager, l logging.Logger) error {
	name := claimscheduling.ControllerName(cachev1alpha1.RedisClusterGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&cachev1alpha1.RedisCluster{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimscheduling.NewReconciler(mgr,
			resource.ClaimKind(cachev1alpha1.RedisClusterGroupVersionKind),
			resource.ClassKind(v1beta1.RedisClassGroupVersionKind),
			claimscheduling.WithLogger(l.WithValues("controller", name)),
			claimscheduling.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupRedisClaimDefaulting adds a controller that reconciles RedisCluster
// claims that omit their resource ref, class ref, and class selector by
// choosing a default Azure Redis resource class if one exists.
func SetupRedisClaimDefaulting(mgr ctrl.Manager, l logging.Logger) error {
	name := claimdefaulting.ControllerName(cachev1alpha1.RedisClusterGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&cachev1alpha1.RedisCluster{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(claimdefaulting.NewReconciler(mgr,
			resource.ClaimKind(cachev1alpha1.RedisClusterGroupVersionKind),
			resource.ClassKind(v1beta1.RedisClassGroupVersionKind),
			claimdefaulting.WithLogger(l.WithValues("controller", name)),
			claimdefaulting.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

// SetupRedisClaimBinding adds a controller that reconciles RedisCluster claims
// with Azure Redis resources, dynamically provisioning them if needed.
func SetupRedisClaimBinding(mgr ctrl.Manager, l logging.Logger) error {
	name := claimbinding.ControllerName(cachev1alpha1.RedisClusterGroupKind)

	r := claimbinding.NewReconciler(mgr,
		resource.ClaimKind(cachev1alpha1.RedisClusterGroupVersionKind),
		resource.ClassKind(v1beta1.RedisClassGroupVersionKind),
		resource.ManagedKind(v1beta1.RedisGroupVersionKind),
		claimbinding.WithBinder(claimbinding.NewAPIBinder(mgr.GetClient(), mgr.GetScheme())),
		claimbinding.WithManagedConfigurators(
			claimbinding.ManagedConfiguratorFn(ConfigureRedis),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureReclaimPolicy),
			claimbinding.ManagedConfiguratorFn(claimbinding.ConfigureNames)),
		claimbinding.WithLogger(l.WithValues("controller", name)),
		claimbinding.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1beta1.RedisClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1beta1.RedisGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1beta1.RedisGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1beta1.Redis{}}, &resource.EnqueueRequestForClaim{}).
		For(&cachev1alpha1.RedisCluster{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureRedis configures the supplied resource (presumed to be a Redis)
// using the supplied resource claim (presumed to be a RedisCluster) and
// resource class.
func ConfigureRedis(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	rc, cmok := cm.(*cachev1alpha1.RedisCluster)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), cachev1alpha1.RedisClusterGroupVersionKind)
	}

	rs, csok := cs.(*v1beta1.RedisClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1beta1.RedisClassGroupVersionKind)
	}

	i, mgok := mg.(*v1beta1.Redis)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1beta1.RedisGroupVersionKind)
	}

	spec := &v1beta1.RedisSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		ForProvider: rs.SpecTemplate.ForProvider,
	}
	if err := resolveAzureClassValues(rc); err != nil {
		return errors.Wrap(err, "cannot resolve Azure class instance values")
	}

	spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{
		Namespace: rs.SpecTemplate.WriteConnectionSecretsToNamespace,
		Name:      string(cm.GetUID()),
	}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	i.Spec = *spec

	return nil
}

func resolveAzureClassValues(rc *cachev1alpha1.RedisCluster) error {
	// EngineVersion is currently the only option we expose at the claim level,
	// and Azure only supports Redis 3.2.
	if rc.Spec.EngineVersion != "" && rc.Spec.EngineVersion != v1beta1.SupportedRedisVersion {
		return errors.Errorf("Azure supports only Redis version %s", v1beta1.SupportedRedisVersion)
	}
	return nil
}
