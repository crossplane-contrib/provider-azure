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
	"fmt"

	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reference"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	cache "github.com/crossplane-contrib/provider-azure/apis/cache/v1beta1"
	database "github.com/crossplane-contrib/provider-azure/apis/database/v1beta1"
	storage "github.com/crossplane-contrib/provider-azure/apis/storage/v1alpha3"
	"github.com/crossplane-contrib/provider-azure/apis/v1alpha3"
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

// MySQLServerID extracts status.ID from the supplied managed resource, which must be
// a Subnet.
func MySQLServerID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		s, ok := mg.(*database.MySQLServer)
		if !ok {
			return ""
		}
		return s.Status.AtProvider.ID
	}
}

func PostgreSQLServerID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		s, ok := mg.(*database.PostgreSQLServer)
		if !ok {
			return ""
		}
		return s.Status.AtProvider.ID
	}
}

func RedisCacheID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		s, ok := mg.(*cache.Redis)
		if !ok {
			return ""
		}
		return s.Status.AtProvider.ID
	}
}

func StorageAccountID() reference.ExtractValueFn {
	return func(mg resource.Managed) string {
		s, ok := mg.(*storage.Account)
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

// ResolveReferences of this PrivateEndpoint
func (mg *PrivateEndpoint) ResolveReferences(ctx context.Context, c client.Reader) error {
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

	// Resolve spec.subnetId
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.SubnetId,
		Reference:    mg.Spec.SubnetIdRef,
		Selector:     mg.Spec.SubnetIdSelector,
		To:           reference.To{Managed: &Subnet{}, List: &SubnetList{}},
		Extract:      SubnetID(),
	})
	if err != nil {
		return errors.Wrap(err, "spec.subnetId")
	}
	mg.Spec.SubnetId = rsp.ResolvedValue
	mg.Spec.SubnetIdRef = rsp.ResolvedReference

	// Resolve the connection resource id reference by type
	// Note we explicitly define which resources we support
	if !privateEndpointResourceSupported(mg.Spec.ResourceType) {
		return fmt.Errorf("unable to resolve refernce for %s, not implemented yet", mg.Spec.ResourceType)
	}
	id, ref, err := mg.getEndpointResourceReference(ctx, c, r)

	mg.Spec.PrivateConnectionDetails.PrivateConnectionResourceId = id
	mg.Spec.PrivateConnectionDetails.PrivateConnectionResourceIdRef = ref

	return err
}

func privateEndpointResourceSupported(t string) bool {
	switch t {
	case "mysqlServer", "postgresqlServer", "redisCache", "blob":
		return true
	}
	return false
}

func (mg *PrivateEndpoint) getEndpointResourceReference(ctx context.Context, c client.Reader, r *reference.APIResolver) (string, *v1.Reference, error) {
	resType := mg.Spec.PrivateConnectionDetails.ResourceType

	var target resource.Managed
	var targetList resource.ManagedList
	var extractFn func(mg resource.Managed) string

	switch resType {
	case "mysqlServer":
		target = &database.MySQLServer{}
		targetList = &database.MySQLServerList{}
		extractFn = MySQLServerID()
	case "postgresqlServer":
		target = &database.PostgreSQLServer{}
		targetList = &database.PostgreSQLServerList{}
		extractFn = PostgreSQLServerID()
	case "redisCache":
		target = &cache.Redis{}
		targetList = &cache.RedisList{}
		extractFn = RedisCacheID()
	case "blob":
		target = &storage.Account{}
		targetList = &storage.AccountList{}
		extractFn = StorageAccountID()
	}

	rsp, err := r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: mg.Spec.PrivateConnectionDetails.PrivateConnectionResourceId,
		Reference:    mg.Spec.PrivateConnectionDetails.PrivateConnectionResourceIdRef,
		Selector:     mg.Spec.PrivateConnectionDetails.PrivateConnectionResourceIdSelector,
		To:           reference.To{Managed: target, List: targetList},
		Extract:      extractFn,
	})

	if err != nil {
		return "", nil, errors.Wrap(err, "spec.privateConnectionDetails.privateConnectionResourceId")
	}

	return rsp.ResolvedValue, rsp.ResolvedReference, err
}
