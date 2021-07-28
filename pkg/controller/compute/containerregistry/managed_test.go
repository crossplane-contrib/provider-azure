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

package containerregistry

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/containerregistry/mgmt/2019-05-01/containerregistry"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	"github.com/crossplane/provider-azure/pkg/clients/registry/fake"
)

type modifier func(*v1alpha3.Registry)

func withState(state string) modifier {
	return func(c *v1alpha3.Registry) {
		c.Status.State = state
	}
}

func withProviderID(id string) modifier {
	return func(c *v1alpha3.Registry) {
		c.Status.ProviderID = id
	}
}

func testRegistry(m ...modifier) *v1alpha3.Registry {
	ac := &v1alpha3.Registry{}

	for _, mod := range m {
		mod(ac)
	}

	return ac
}

func TestObserve(t *testing.T) {
	errBoom := errors.New("boom")
	id := "koolAD"
	stateWat := "Wat"

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		eo  managed.ExternalObservation
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"ErrNotRegistry": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotRegistry),
			},
		},
		"ErrRegistryNotFound": {
			e: &external{
				client: &fake.MockContainerRegistry{
					MockGet: func(ctx context.Context, resourceGroupName string, registryName string) (result containerregistry.Registry, err error) {
						return containerregistry.Registry{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  testRegistry(),
			},
			want: want{
				eo: managed.ExternalObservation{ResourceExists: false},
				mg: testRegistry(),
			},
		},
		"ErrGetRegistry": {
			e: &external{
				client: &fake.MockContainerRegistry{
					MockGet: func(ctx context.Context, resourceGroupName string, registryName string) (result containerregistry.Registry, err error) {
						return containerregistry.Registry{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  testRegistry(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetRegistry),
				mg:  testRegistry(),
			},
		},
		"NotReady": {
			e: &external{
				client: &fake.MockContainerRegistry{
					MockGet: func(ctx context.Context, resourceGroupName string, registryName string) (result containerregistry.Registry, err error) {
						return containerregistry.Registry{
							ID: to.StringPtr(id),
							RegistryProperties: &containerregistry.RegistryProperties{
								ProvisioningState: containerregistry.ProvisioningState(stateWat),
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  testRegistry(),
			},
			want: want{
				eo: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
				mg: testRegistry(
					withProviderID(id),
					withState(stateWat),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eo, err := tc.e.Observe(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("tc.e.Observe(...): -want managed, +got managed:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.eo, eo); diff != "" {
				t.Errorf("tc.e.Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		ec  managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"ErrNotRegistry": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotRegistry),
			},
		},
		"ErrCreateRegistry": {
			e: &external{
				client: &fake.MockContainerRegistry{
					MockCreate: func(ctx context.Context, resourceGroupName string, registryName string, registry containerregistry.Registry) (result containerregistry.RegistriesCreateFuture, err error) {
						return containerregistry.RegistriesCreateFuture{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  testRegistry(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateRegistry),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ec, err := tc.e.Create(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Create(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.ec, ec); diff != "" {
				t.Errorf("tc.e.Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want error
	}{
		"ErrNotRegistry": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotRegistry),
		},
		"ErrDeleteRegistry": {
			e: &external{
				client: &fake.MockContainerRegistry{
					MockDelete: func(ctx context.Context, resourceGroupName string, registryName string) (result containerregistry.RegistriesDeleteFuture, err error) {
						return containerregistry.RegistriesDeleteFuture{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  testRegistry(),
			},
			want: errors.Wrap(errBoom, errDeleteRegistry),
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := tc.e.Delete(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want, got, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}
