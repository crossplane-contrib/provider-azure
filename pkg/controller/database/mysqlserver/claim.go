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

package mysqlserver

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/source"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	databasev1alpha1 "github.com/crossplaneio/crossplane/apis/database/v1alpha1"

	"github.com/crossplaneio/stack-azure/apis/database/v1alpha2"
)

// A ClaimSchedulingController reconciles MySQLInstance claims that include a
// class selector but omit their class and resource references by picking a
// random matching Azure SQLServer class, if any.
type ClaimSchedulingController struct{}

// SetupWithManager sets up the ClaimSchedulingController using the supplied
// manager.
func (c *ClaimSchedulingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("scheduler.%s.%s.%s",
		databasev1alpha1.MySQLInstanceKind,
		v1alpha2.MysqlServerKind,
		v1alpha2.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimSchedulingReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
			resource.ClassKind(v1alpha2.SQLServerClassGroupVersionKind),
		))
}

// A ClaimDefaultingController reconciles MySQLInstance claims that omit their
// resource ref, class ref, and class selector by choosing a default Azure
// SQLServer resource class if one exists.
type ClaimDefaultingController struct{}

// SetupWithManager sets up the ClaimDefaultingController using the supplied
// manager.
func (c *ClaimDefaultingController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("defaulter.%s.%s.%s",
		databasev1alpha1.MySQLInstanceKind,
		v1alpha2.MysqlServerKind,
		v1alpha2.Group))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(resource.NewPredicates(resource.AllOf(
			resource.HasNoClassSelector(),
			resource.HasNoClassReference(),
			resource.HasNoManagedResourceReference(),
		))).
		Complete(resource.NewClaimDefaultingReconciler(mgr,
			resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
			resource.ClassKind(v1alpha2.SQLServerClassGroupVersionKind),
		))
}

// A ClaimController reconciles MySQLInstance claims with Azure MysqlServer
// resources, dynamically provisioning them if needed.
type ClaimController struct{}

// SetupWithManager sets up the ClaimController using the supplied manager.
func (c *ClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		databasev1alpha1.MySQLInstanceKind,
		v1alpha2.MysqlServerKind,
		v1alpha2.Group))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.MySQLInstanceGroupVersionKind),
		resource.ClassKind(v1alpha2.SQLServerClassGroupVersionKind),
		resource.ManagedKind(v1alpha2.MysqlServerGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigureMysqlServer),
			resource.NewObjectMetaConfigurator(mgr.GetScheme()),
		))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha2.SQLServerClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha2.MysqlServerGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha2.MysqlServerGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha2.MysqlServer{}}, &resource.EnqueueRequestForClaim{}).
		For(&databasev1alpha1.MySQLInstance{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigureMysqlServer configures the supplied resource (presumed to be
// a MysqlServer) using the supplied resource claim (presumed to be a
// MySQLInstance) and resource class.
func ConfigureMysqlServer(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	my, cmok := cm.(*databasev1alpha1.MySQLInstance)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), databasev1alpha1.MySQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha2.SQLServerClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.SQLServerClassGroupVersionKind)
	}

	s, mgok := mg.(*v1alpha2.MysqlServer)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha2.MysqlServerGroupVersionKind)
	}

	spec := &v1alpha2.SQLServerSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		SQLServerParameters: rs.SpecTemplate.SQLServerParameters,
	}

	if my.Spec.EngineVersion != "" {
		spec.Version = my.Spec.EngineVersion
	}

	spec.WriteConnectionSecretToReference = &runtimev1alpha1.SecretReference{
		Namespace: rs.SpecTemplate.WriteConnectionSecretsToNamespace,
		Name:      string(cm.GetUID()),
	}
	spec.ProviderReference = rs.SpecTemplate.ProviderReference
	spec.ReclaimPolicy = rs.SpecTemplate.ReclaimPolicy

	s.Spec = *spec

	return nil
}
