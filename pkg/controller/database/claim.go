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

// NOTE(hasheddan): consider combining into single controller

// PostgreSQLInstanceClaimController is responsible for adding the PostgreSQLInstance
// claim controller and its corresponding reconciler to the manager with any runtime configuration.
type PostgreSQLInstanceClaimController struct{}

// SetupWithManager adds a controller that reconciles PostgreSQLInstance instance claims.
func (c *PostgreSQLInstanceClaimController) SetupWithManager(mgr ctrl.Manager) error {
	name := strings.ToLower(fmt.Sprintf("%s.%s.%s",
		databasev1alpha1.PostgreSQLInstanceKind,
		v1alpha2.PostgresqlServerKind,
		v1alpha2.Group))

	r := resource.NewClaimReconciler(mgr,
		resource.ClaimKind(databasev1alpha1.PostgreSQLInstanceGroupVersionKind),
		resource.ClassKind(v1alpha2.SQLServerClassGroupVersionKind),
		resource.ManagedKind(v1alpha2.PostgresqlServerGroupVersionKind),
		resource.WithManagedConfigurators(
			resource.ManagedConfiguratorFn(ConfigurePostgresqlServer),
			resource.NewObjectMetaConfigurator(mgr.GetScheme()),
		))

	p := resource.NewPredicates(resource.AnyOf(
		resource.HasClassReferenceKind(resource.ClassKind(v1alpha2.SQLServerClassGroupVersionKind)),
		resource.HasManagedResourceReferenceKind(resource.ManagedKind(v1alpha2.PostgresqlServerGroupVersionKind)),
		resource.IsManagedKind(resource.ManagedKind(v1alpha2.PostgresqlServerGroupVersionKind), mgr.GetScheme()),
	))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		Watches(&source.Kind{Type: &v1alpha2.PostgresqlServer{}}, &resource.EnqueueRequestForClaim{}).
		For(&databasev1alpha1.PostgreSQLInstance{}).
		WithEventFilter(p).
		Complete(r)
}

// ConfigurePostgresqlServer configures the supplied resource (presumed to be a
// PostgresqlServer) using the supplied resource claim (presumed to be a
// PostgreSQLInstance) and resource class.
func ConfigurePostgresqlServer(_ context.Context, cm resource.Claim, cs resource.Class, mg resource.Managed) error {
	pg, cmok := cm.(*databasev1alpha1.PostgreSQLInstance)
	if !cmok {
		return errors.Errorf("expected resource claim %s to be %s", cm.GetName(), databasev1alpha1.PostgreSQLInstanceGroupVersionKind)
	}

	rs, csok := cs.(*v1alpha2.SQLServerClass)
	if !csok {
		return errors.Errorf("expected resource class %s to be %s", cs.GetName(), v1alpha2.SQLServerClassGroupVersionKind)
	}

	s, mgok := mg.(*v1alpha2.PostgresqlServer)
	if !mgok {
		return errors.Errorf("expected managed resource %s to be %s", mg.GetName(), v1alpha2.PostgresqlServerGroupVersionKind)
	}

	spec := &v1alpha2.SQLServerSpec{
		ResourceSpec: runtimev1alpha1.ResourceSpec{
			ReclaimPolicy: runtimev1alpha1.ReclaimRetain,
		},
		SQLServerParameters: rs.SpecTemplate.SQLServerParameters,
	}

	if pg.Spec.EngineVersion != "" {
		spec.Version = pg.Spec.EngineVersion
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

// MySQLInstanceClaimController is responsible for adding the MySQLInstance
// claim controller and its corresponding reconciler to the manager with any runtime configuration.
type MySQLInstanceClaimController struct{}

// SetupWithManager adds a controller that reconciles MySQLInstance instance claims.
func (c *MySQLInstanceClaimController) SetupWithManager(mgr ctrl.Manager) error {
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
