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

package v1alpha3

import (
	"context"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/v1alpha3"
)

// SubnetID extracts status.ID from the supplied managed resource, which must be
// a Subnet.
func SubnetID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		s, ok := mg.(*Subnet)
		if !ok {
			return ""
		}
		return s.Status.ID
	}
}

// NetworkInterfaceID extracts status.ID from the supplied managed resource, which must be
// a NetworkInterface.
func NetworkInterfaceID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		s, ok := mg.(*NetworkInterface)
		if !ok {
			return ""
		}
		return s.Status.ID
	}
}

// PublicIPAddressID extracts status.ID from the supplied managed resource, which must be
// a PublicIPAddress.
func PublicIPAddressID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		s, ok := mg.(*PublicIPAddress)
		if !ok {
			return ""
		}
		return s.Status.ID
	}
}

// ResolveReferences of this VirtualNetwork
func (mg *VirtualNetwork) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.resourceGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ResourceGroupName,
		Reference:    mg.Spec.ResourceGroupNameRef,
		Selector:     mg.Spec.ResourceGroupNameSelector,
		To:           reference.To{Managed: &v1alpha3.ResourceGroup{}, List: &v1alpha3.ResourceGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.resourceGroupName")
	}
	mg.Spec.ResourceGroupName = rsp.ResolvedValue
	mg.Spec.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this Subnet
func (mg *Subnet) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.resourceGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ResourceGroupName,
		Reference:    mg.Spec.ResourceGroupNameRef,
		Selector:     mg.Spec.ResourceGroupNameSelector,
		To:           reference.To{Managed: &v1alpha3.ResourceGroup{}, List: &v1alpha3.ResourceGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.resourceGroupName")
	}
	mg.Spec.ResourceGroupName = rsp.ResolvedValue
	mg.Spec.ResourceGroupNameRef = rsp.ResolvedReference

	// Resolve spec.virtualNetworkName
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.VirtualNetworkName,
		Reference:    mg.Spec.VirtualNetworkNameRef,
		Selector:     mg.Spec.VirtualNetworkNameSelector,
		To:           reference.To{Managed: &VirtualNetwork{}, List: &VirtualNetworkList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.virtualNetworkName")
	}
	mg.Spec.VirtualNetworkName = rsp.ResolvedValue
	mg.Spec.VirtualNetworkNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this PublicIPAddress
func (mg *PublicIPAddress) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.resourceGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ResourceGroupName,
		Reference:    mg.Spec.ResourceGroupNameRef,
		Selector:     mg.Spec.ResourceGroupNameSelector,
		To:           reference.To{Managed: &v1alpha3.ResourceGroup{}, List: &v1alpha3.ResourceGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.resourceGroupName")
	}
	mg.Spec.ResourceGroupName = rsp.ResolvedValue
	mg.Spec.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this PublicIPAddress
func (mg *NetworkInterface) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	// Resolve spec.resourceGroupName
	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.ResourceGroupName,
		Reference:    mg.Spec.ResourceGroupNameRef,
		Selector:     mg.Spec.ResourceGroupNameSelector,
		To:           reference.To{Managed: &v1alpha3.ResourceGroup{}, List: &v1alpha3.ResourceGroupList{}},
		Extract:      reference.ExternalName(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.resourceGroupName")
	}
	mg.Spec.ResourceGroupName = rsp.ResolvedValue
	mg.Spec.ResourceGroupNameRef = rsp.ResolvedReference

	// Resolve spec.properties.interfaceIPConfigurations[].publicIPAddress
	for i, iface := range mg.Spec.NetworkInterfaceFormat.IPConfigurations {
		rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: iface.PublicIPAddressID,
			Reference:    iface.PublicIPAddressIDRef,
			Selector:     iface.PublicIPAddressIDSelector,
			To:           reference.To{Managed: &PublicIPAddress{}, List: &PublicIPAddressList{}},
			Extract:      PublicIPAddressID(),
		})
		if err != nil {
			return errors.Wrap(err, "spec.properties.interfaceIPConfigurations[].publicIPAddress")
		}
		mg.Spec.NetworkInterfaceFormat.IPConfigurations[i].PublicIPAddressID = rsp.ResolvedValue
		mg.Spec.NetworkInterfaceFormat.IPConfigurations[i].PublicIPAddressIDRef = rsp.ResolvedReference
	}

	// Resolve spec.properties.interfaceIPConfigurations[].subnet
	for i, iface := range mg.Spec.NetworkInterfaceFormat.IPConfigurations {
		rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: iface.SubnetID,
			Reference:    iface.SubnetIDRef,
			Selector:     iface.SubnetIDSelector,
			To:           reference.To{Managed: &Subnet{}, List: &SubnetList{}},
			Extract:      SubnetID(),
		})
		if err != nil {
			return errors.Wrap(err, "spec.properties.interfaceIPConfigurations[].subnet")
		}
		mg.Spec.NetworkInterfaceFormat.IPConfigurations[i].SubnetID = rsp.ResolvedValue
		mg.Spec.NetworkInterfaceFormat.IPConfigurations[i].SubnetIDRef = rsp.ResolvedReference
	}

	return nil
}
