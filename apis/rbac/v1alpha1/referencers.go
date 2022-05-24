/*
Copyright 2021 The Crossplane Authors.

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

package v1alpha1

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
)

// ResolveReferences of this ServicePrincipal
func (mg *ServicePrincipal) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.applicationID
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.ApplicationID,
		Reference:    mg.Spec.ForProvider.ApplicationIDRef,
		Selector:     mg.Spec.ForProvider.ApplicationIDSelector,
		To:           reference.To{Managed: &Application{}, List: &ApplicationList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.applicationID")
	}
	mg.Spec.ForProvider.ApplicationID = rsp.ResolvedValue
	mg.Spec.ForProvider.ApplicationIDRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this ServicePrincipal
func (mg *RoleAssignment) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ForProvider.PrincipalID,
		Reference:    mg.Spec.ForProvider.PrincipalIDRef,
		Selector:     mg.Spec.ForProvider.PrincipalIDSelector,
		To:           reference.To{Managed: &ServicePrincipal{}, List: &ServicePrincipalList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.principalID")
	}
	mg.Spec.ForProvider.PrincipalID = rsp.ResolvedValue
	mg.Spec.ForProvider.PrincipalIDRef = rsp.ResolvedReference

	return nil
}
