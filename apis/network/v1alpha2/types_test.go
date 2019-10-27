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
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"
)

var _ resource.AttributeReferencer = (*VirtualNetworkNameReferencerForSubnet)(nil)
var _ resource.AttributeReferencer = (*ResourceGroupNameReferencerForVirtualNetwork)(nil)
var _ resource.AttributeReferencer = (*ResourceGroupNameReferencerForSubnet)(nil)

func TestVirtualNetworkNameReferencerForSubnet_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &VirtualNetworkNameReferencerForSubnet{}
	expectedErr := errors.New(errResourceIsNotSubnet)

	err := r.Assign(&struct{ resource.CanReference }{}, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestVirtualNetworkNameReferencerForSubnet_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &VirtualNetworkNameReferencerForSubnet{}
	res := &Subnet{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.VirtualNetworkName, "mockValue"); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}

func TestResourceGroupNameReferencerForVirtualNetwork_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &ResourceGroupNameReferencerForVirtualNetwork{}
	expectedErr := errors.New(errResourceIsNotVirtualNetwork)

	err := r.Assign(&struct{ resource.CanReference }{}, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestResourceGroupNameReferencerForVirtualNetwork_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &ResourceGroupNameReferencerForVirtualNetwork{}
	res := &VirtualNetwork{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.ResourceGroupName, "mockValue"); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}

func TestResourceGroupNameReferencerForSubnet_AssignInvalidType_ReturnsErr(t *testing.T) {

	r := &ResourceGroupNameReferencerForSubnet{}
	expectedErr := errors.New(errResourceIsNotSubnet)

	err := r.Assign(&struct{ resource.CanReference }{}, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}

func TestResourceGroupNameReferencerForSubnet_AssignValidType_ReturnsExpected(t *testing.T) {

	r := &ResourceGroupNameReferencerForSubnet{}
	res := &Subnet{}
	var expectedErr error

	err := r.Assign(res, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}

	if diff := cmp.Diff(res.Spec.ResourceGroupName, "mockValue"); diff != "" {
		t.Errorf("Assign(...): -want value, +got value:\n%s", diff)
	}
}
