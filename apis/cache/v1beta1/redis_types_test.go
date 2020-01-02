/*
Copyright 2020 The Crossplane Authors.

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
	"testing"

	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
)

var _ resource.AttributeReferencer = &ResourceGroupNameReferencerForRedis{}

func TestResourceGroupNameReferencerForRedis_Assign(t *testing.T) {
	r := &ResourceGroupNameReferencerForRedis{}
	expectedErr := errors.New(errNotRedis)

	err := r.Assign(&struct{ resource.CanReference }{}, "mockValue")
	if diff := cmp.Diff(expectedErr, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
	cr := &Redis{}
	err = r.Assign(cr, "mockValue")
	if diff := cmp.Diff(nil, err, test.EquateErrors()); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
	if diff := cmp.Diff(&Redis{Spec: RedisSpec{ForProvider: RedisParameters{ResourceGroupName: "mockValue"}}}, cr); diff != "" {
		t.Errorf("Assign(...): -want error, +got error:\n%s", diff)
	}
}
