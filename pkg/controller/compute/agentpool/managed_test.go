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

package agentpool

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/containerservice/mgmt/containerservice"
	original "github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2020-03-01/containerservice"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	"github.com/crossplane/provider-azure/pkg/clients/compute/agentpool"
)

type modifier func(*v1alpha3.AgentPool)

func withState(state string) modifier {
	return func(c *v1alpha3.AgentPool) {
		c.Status.State = state
	}
}

func withProviderID(id string) modifier {
	return func(c *v1alpha3.AgentPool) {
		c.Status.ProviderID = id
	}
}

func pool(m ...modifier) *v1alpha3.AgentPool {
	ac := &v1alpha3.AgentPool{}
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
		"ErrNotAgentPool": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotAgentPool),
			},
		},
		"ErrAgentPoolNotFound": {
			e: &external{
				c: &agentpool.Mock{
					MockGet: func(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string) (result containerservice.AgentPool, err error) {
						return containerservice.AgentPool{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  pool(),
			},
			want: want{
				eo: managed.ExternalObservation{ResourceExists: false},
				mg: pool(),
			},
		},
		"ErrGetAgentPool": {
			e: &external{
				c: &agentpool.Mock{
					MockGet: func(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string) (result containerservice.AgentPool, err error) {
						return containerservice.AgentPool{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  pool(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetAgentPool),
				mg:  pool(),
			},
		},
		"NotReady": {
			e: &external{
				c: &agentpool.Mock{
					MockGet: func(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string) (result containerservice.AgentPool, err error) {
						return containerservice.AgentPool{
							ID: to.StringPtr(id),
							ManagedClusterAgentPoolProfileProperties: &containerservice.ManagedClusterAgentPoolProfileProperties{
								ProvisioningState: to.StringPtr(stateWat),
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  pool(),
			},
			want: want{
				eo: managed.ExternalObservation{ResourceExists: true},
				mg: pool(
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
		"ErrNotAgentPool": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotAgentPool),
			},
		},
		"ErrCreateAgentPool": {
			e: &external{
				c: &agentpool.Mock{
					MockCreateOrUpdate: func(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string, parameters original.AgentPool) (result original.AgentPoolsCreateOrUpdateFuture, err error) {
						return original.AgentPoolsCreateOrUpdateFuture{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  pool(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateAgentPool),
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
		"ErrNotAgentPool": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotAgentPool),
		},
		"ErrDeleteAgentPool": {
			e: &external{
				c: &agentpool.Mock{MockDelete: func(ctx context.Context, resourceGroupName string, resourceName string, agentPoolName string) (result original.AgentPoolsDeleteFuture, err error) {
					return original.AgentPoolsDeleteFuture{}, errBoom
				}},
			},
			args: args{
				ctx: context.Background(),
				mg:  pool(),
			},
			want: errors.Wrap(errBoom, errDeleteAgentPool),
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
