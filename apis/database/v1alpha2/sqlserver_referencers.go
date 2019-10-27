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

package v1alpha2

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
)

// A MySQLServerNameReferencer returns the server name of a referenced
// MySQLServer.
type MySQLServerNameReferencer struct {
	corev1.LocalObjectReference `json:",inline"`
}

// GetStatus of the referenced MySQLServer.
func (v *MySQLServerNameReferencer) GetStatus(ctx context.Context, _ resource.CanReference, reader client.Reader) ([]resource.ReferenceStatus, error) {
	refObj := MySQLServer{}
	nn := types.NamespacedName{Name: v.Name}
	if err := reader.Get(ctx, nn, &refObj); err != nil {
		if kerrors.IsNotFound(err) {
			return []resource.ReferenceStatus{{Name: v.Name, Status: resource.ReferenceNotFound}}, nil
		}

		return nil, err
	}

	if !resource.IsConditionTrue(refObj.GetCondition(runtimev1alpha1.TypeReady)) {
		return []resource.ReferenceStatus{{Name: v.Name, Status: resource.ReferenceNotReady}}, nil
	}

	return []resource.ReferenceStatus{{Name: v.Name, Status: resource.ReferenceReady}}, nil
}

// Build returns the server name of the referenced MySQLServer.
func (v *MySQLServerNameReferencer) Build(ctx context.Context, _ resource.CanReference, reader client.Reader) (string, error) {
	refObj := MySQLServer{}
	nn := types.NamespacedName{Name: v.Name}
	if err := reader.Get(ctx, nn, &refObj); err != nil {
		return "", err
	}

	return refObj.GetName(), nil
}

// A PostgreSQLServerNameReferencer returns the server name of a referenced
// PostgreSQLServer.
type PostgreSQLServerNameReferencer struct {
	corev1.LocalObjectReference `json:",inline"`
}

// GetStatus implements GetStatus method of AttributeReferencer interface
func (v *PostgreSQLServerNameReferencer) GetStatus(ctx context.Context, _ resource.CanReference, reader client.Reader) ([]resource.ReferenceStatus, error) {
	refObj := PostgreSQLServer{}
	nn := types.NamespacedName{Name: v.Name}
	if err := reader.Get(ctx, nn, &refObj); err != nil {
		if kerrors.IsNotFound(err) {
			return []resource.ReferenceStatus{{Name: v.Name, Status: resource.ReferenceNotFound}}, nil
		}

		return nil, err
	}

	if !resource.IsConditionTrue(refObj.GetCondition(runtimev1alpha1.TypeReady)) {
		return []resource.ReferenceStatus{{Name: v.Name, Status: resource.ReferenceNotReady}}, nil
	}

	return []resource.ReferenceStatus{{Name: v.Name, Status: resource.ReferenceReady}}, nil
}

// Build returns the server name of the referenced PostgreSQLServer.
func (v *PostgreSQLServerNameReferencer) Build(ctx context.Context, _ resource.CanReference, reader client.Reader) (string, error) {
	refObj := PostgreSQLServer{}
	nn := types.NamespacedName{Name: v.Name}
	if err := reader.Get(ctx, nn, &refObj); err != nil {
		return "", err
	}

	return refObj.GetName(), nil
}
