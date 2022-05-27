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

package v1beta1

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"

	"github.com/crossplane-contrib/provider-azure/apis/v1alpha3"
)

// ResolveReferences of this MySQLServer.
func (mg *MySQLServer) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.resourceGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.ResourceGroupName,
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To:           reference.To{Managed: &v1alpha3.ResourceGroup{}, List: &v1alpha3.ResourceGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.resourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = rsp.ResolvedValue
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this MySQLServerConfiguration.
func (mg *MySQLServerConfiguration) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.resourceGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.ResourceGroupName,
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To:           reference.To{Managed: &v1alpha3.ResourceGroup{}, List: &v1alpha3.ResourceGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.resourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = rsp.ResolvedValue
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	// Resolve spec.forProvider.resourceGroupName
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.ServerName,
		Reference:    mg.Spec.ForProvider.ServerNameRef,
		Selector:     mg.Spec.ForProvider.ServerNameSelector,
		To:           reference.To{Managed: &MySQLServer{}, List: &MySQLServerList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.serverName")
	}
	mg.Spec.ForProvider.ServerName = rsp.ResolvedValue
	mg.Spec.ForProvider.ServerNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this PostgreSQLServer.
func (mg *PostgreSQLServer) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.resourceGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.ResourceGroupName,
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To:           reference.To{Managed: &v1alpha3.ResourceGroup{}, List: &v1alpha3.ResourceGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.resourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = rsp.ResolvedValue
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this PostgreSQLServerConfiguration.
func (mg *PostgreSQLServerConfiguration) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.forProvider.resourceGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.ResourceGroupName,
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To:           reference.To{Managed: &v1alpha3.ResourceGroup{}, List: &v1alpha3.ResourceGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.resourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = rsp.ResolvedValue
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	// Resolve spec.forProvider.resourceGroupName
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.ServerName,
		Reference:    mg.Spec.ForProvider.ServerNameRef,
		Selector:     mg.Spec.ForProvider.ServerNameSelector,
		To:           reference.To{Managed: &PostgreSQLServer{}, List: &PostgreSQLServerList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.forProvider.serverName")
	}
	mg.Spec.ForProvider.ServerName = rsp.ResolvedValue
	mg.Spec.ForProvider.ServerNameRef = rsp.ResolvedReference

	return nil
}
