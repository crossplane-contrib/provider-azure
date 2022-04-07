/*
Copyright 2022 The Crossplane Authors.

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
// Code generated by angryjet. DO NOT EDIT.

package v1alpha2

import (
	"context"
	v1alpha21 "github.com/crossplane-contrib/provider-jet-azure/apis/azure/v1alpha2"
	v1alpha2 "github.com/crossplane-contrib/provider-jet-azure/apis/storage/v1alpha2"
	reference "github.com/crossplane/crossplane-runtime/pkg/reference"
	errors "github.com/pkg/errors"
	client "sigs.k8s.io/controller-runtime/pkg/client"
)

// ResolveReferences of this IOTHub.
func (mg *IOTHub) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var err error

	for i3 := 0; i3 < len(mg.Spec.ForProvider.Endpoint); i3++ {
		rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Endpoint[i3].ContainerName),
			Extract:      reference.ExternalName(),
			Reference:    mg.Spec.ForProvider.Endpoint[i3].ContainerNameRef,
			Selector:     mg.Spec.ForProvider.Endpoint[i3].ContainerNameSelector,
			To: reference.To{
				List:    &v1alpha2.ContainerList{},
				Managed: &v1alpha2.Container{},
			},
		})
		if err != nil {
			return errors.Wrap(err, "mg.Spec.ForProvider.Endpoint[i3].ContainerName")
		}
		mg.Spec.ForProvider.Endpoint[i3].ContainerName = reference.ToPtrValue(rsp.ResolvedValue)
		mg.Spec.ForProvider.Endpoint[i3].ContainerNameRef = rsp.ResolvedReference

	}
	for i3 := 0; i3 < len(mg.Spec.ForProvider.Endpoint); i3++ {
		rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
			CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.Endpoint[i3].ResourceGroupName),
			Extract:      reference.ExternalName(),
			Reference:    mg.Spec.ForProvider.Endpoint[i3].ResourceGroupNameRef,
			Selector:     mg.Spec.ForProvider.Endpoint[i3].ResourceGroupNameSelector,
			To: reference.To{
				List:    &v1alpha21.ResourceGroupList{},
				Managed: &v1alpha21.ResourceGroup{},
			},
		})
		if err != nil {
			return errors.Wrap(err, "mg.Spec.ForProvider.Endpoint[i3].ResourceGroupName")
		}
		mg.Spec.ForProvider.Endpoint[i3].ResourceGroupName = reference.ToPtrValue(rsp.ResolvedValue)
		mg.Spec.ForProvider.Endpoint[i3].ResourceGroupNameRef = rsp.ResolvedReference

	}
	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ResourceGroupName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To: reference.To{
			List:    &v1alpha21.ResourceGroupList{},
			Managed: &v1alpha21.ResourceGroup{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ResourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this IOTHubConsumerGroup.
func (mg *IOTHubConsumerGroup) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var err error

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.IOTHubName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.IOTHubNameRef,
		Selector:     mg.Spec.ForProvider.IOTHubNameSelector,
		To: reference.To{
			List:    &IOTHubList{},
			Managed: &IOTHub{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.IOTHubName")
	}
	mg.Spec.ForProvider.IOTHubName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.IOTHubNameRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ResourceGroupName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To: reference.To{
			List:    &v1alpha21.ResourceGroupList{},
			Managed: &v1alpha21.ResourceGroup{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ResourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this IOTHubDPS.
func (mg *IOTHubDPS) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var err error

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ResourceGroupName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To: reference.To{
			List:    &v1alpha21.ResourceGroupList{},
			Managed: &v1alpha21.ResourceGroup{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ResourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this IOTHubDPSCertificate.
func (mg *IOTHubDPSCertificate) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var err error

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.IOTDPSName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.IOTDPSNameRef,
		Selector:     mg.Spec.ForProvider.IOTDPSNameSelector,
		To: reference.To{
			List:    &IOTHubDPSList{},
			Managed: &IOTHubDPS{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.IOTDPSName")
	}
	mg.Spec.ForProvider.IOTDPSName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.IOTDPSNameRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ResourceGroupName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To: reference.To{
			List:    &v1alpha21.ResourceGroupList{},
			Managed: &v1alpha21.ResourceGroup{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ResourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this IOTHubDPSSharedAccessPolicy.
func (mg *IOTHubDPSSharedAccessPolicy) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var err error

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.IOTHubDPSName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.IOTHubDPSNameRef,
		Selector:     mg.Spec.ForProvider.IOTHubDPSNameSelector,
		To: reference.To{
			List:    &IOTHubDPSList{},
			Managed: &IOTHubDPS{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.IOTHubDPSName")
	}
	mg.Spec.ForProvider.IOTHubDPSName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.IOTHubDPSNameRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ResourceGroupName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To: reference.To{
			List:    &v1alpha21.ResourceGroupList{},
			Managed: &v1alpha21.ResourceGroup{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ResourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this IOTHubEndpointStorageContainer.
func (mg *IOTHubEndpointStorageContainer) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var err error

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ContainerName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.ContainerNameRef,
		Selector:     mg.Spec.ForProvider.ContainerNameSelector,
		To: reference.To{
			List:    &v1alpha2.ContainerList{},
			Managed: &v1alpha2.Container{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ContainerName")
	}
	mg.Spec.ForProvider.ContainerName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ContainerNameRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.IOTHubName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.IOTHubNameRef,
		Selector:     mg.Spec.ForProvider.IOTHubNameSelector,
		To: reference.To{
			List:    &IOTHubList{},
			Managed: &IOTHub{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.IOTHubName")
	}
	mg.Spec.ForProvider.IOTHubName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.IOTHubNameRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ResourceGroupName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To: reference.To{
			List:    &v1alpha21.ResourceGroupList{},
			Managed: &v1alpha21.ResourceGroup{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ResourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this IOTHubFallbackRoute.
func (mg *IOTHubFallbackRoute) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var mrsp reference.MultiResolutionResponse
	var err error

	mrsp, err = r.ResolveMultiple(ctx, reference.MultiResolutionRequest{
		CurrentValues: reference.FromPtrValues(mg.Spec.ForProvider.EndpointNames),
		Extract:       reference.ExternalName(),
		References:    mg.Spec.ForProvider.EndpointNamesRefs,
		Selector:      mg.Spec.ForProvider.EndpointNamesSelector,
		To: reference.To{
			List:    &IOTHubEndpointStorageContainerList{},
			Managed: &IOTHubEndpointStorageContainer{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.EndpointNames")
	}
	mg.Spec.ForProvider.EndpointNames = reference.ToPtrValues(mrsp.ResolvedValues)
	mg.Spec.ForProvider.EndpointNamesRefs = mrsp.ResolvedReferences

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.IOTHubName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.IOTHubNameRef,
		Selector:     mg.Spec.ForProvider.IOTHubNameSelector,
		To: reference.To{
			List:    &IOTHubList{},
			Managed: &IOTHub{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.IOTHubName")
	}
	mg.Spec.ForProvider.IOTHubName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.IOTHubNameRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ResourceGroupName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To: reference.To{
			List:    &v1alpha21.ResourceGroupList{},
			Managed: &v1alpha21.ResourceGroup{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ResourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}

// ResolveReferences of this IOTHubSharedAccessPolicy.
func (mg *IOTHubSharedAccessPolicy) ResolveReferences(ctx context.Context, c client.Reader) error {
	r := reference.NewAPIResolver(c, mg)

	var rsp reference.ResolutionResponse
	var err error

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.IOTHubName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.IOTHubNameRef,
		Selector:     mg.Spec.ForProvider.IOTHubNameSelector,
		To: reference.To{
			List:    &IOTHubList{},
			Managed: &IOTHub{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.IOTHubName")
	}
	mg.Spec.ForProvider.IOTHubName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.IOTHubNameRef = rsp.ResolvedReference

	rsp, err = r.Resolve(ctx, reference.ResolutionRequest{
		CurrentValue: reference.FromPtrValue(mg.Spec.ForProvider.ResourceGroupName),
		Extract:      reference.ExternalName(),
		Reference:    mg.Spec.ForProvider.ResourceGroupNameRef,
		Selector:     mg.Spec.ForProvider.ResourceGroupNameSelector,
		To: reference.To{
			List:    &v1alpha21.ResourceGroupList{},
			Managed: &v1alpha21.ResourceGroup{},
		},
	})
	if err != nil {
		return errors.Wrap(err, "mg.Spec.ForProvider.ResourceGroupName")
	}
	mg.Spec.ForProvider.ResourceGroupName = reference.ToPtrValue(rsp.ResolvedValue)
	mg.Spec.ForProvider.ResourceGroupNameRef = rsp.ResolvedReference

	return nil
}
