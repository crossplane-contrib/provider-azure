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

package postgresqlservervirtualnetworkrule

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/go-autorest/autorest"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	azure "github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients"
	"github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients/fake"

	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/database/v1alpha3"
)

const (
	name              = "coolSubnet"
	uid               = types.UID("definitely-a-uuid")
	serverName        = "coolVnet"
	resourceGroupName = "coolRG"
	vnetSubnetID      = "/the/best/subnet/ever"
	resourceID        = "a-very-cool-id"
	resourceType      = "cooltype"
)

var (
	ctx       = context.Background()
	errorBoom = errors.New("boom")
)

type testCase struct {
	name    string
	e       managed.ExternalClient
	r       resource.Managed
	want    resource.Managed
	wantErr error
}

type virtualNetworkRuleModifier func(*v1alpha3.PostgreSQLServerVirtualNetworkRule)

func withConditions(c ...xpv1.Condition) virtualNetworkRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerVirtualNetworkRule) { r.Status.ConditionedStatus.Conditions = c }
}

func withType(s string) virtualNetworkRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerVirtualNetworkRule) { r.Status.Type = s }
}

func withID(s string) virtualNetworkRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerVirtualNetworkRule) { r.Status.ID = s }
}

func withState(s string) virtualNetworkRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerVirtualNetworkRule) { r.Status.State = s }
}

func virtualNetworkRule(sm ...virtualNetworkRuleModifier) *v1alpha3.PostgreSQLServerVirtualNetworkRule {
	r := &v1alpha3.PostgreSQLServerVirtualNetworkRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.PostgreSQLVirtualNetworkRuleSpec{
			ServerName:        serverName,
			ResourceGroupName: resourceGroupName,
			VirtualNetworkRuleProperties: v1alpha3.VirtualNetworkRuleProperties{
				VirtualNetworkSubnetID:           vnetSubnetID,
				IgnoreMissingVnetServiceEndpoint: true,
			},
		},
		Status: v1alpha3.VirtualNetworkRuleStatus{},
	}

	meta.SetExternalName(r, name)

	for _, m := range sm {
		m(r)
	}

	return r
}

// Test that our Reconciler implementation satisfies the Reconciler interface.
var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connecter{}

func TestCreate(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotPostgreSQLServerlVirtualNetworkRule",
			e:       &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{}},
			r:       &v1alpha3.MySQLServerVirtualNetworkRule{},
			want:    &v1alpha3.MySQLServerVirtualNetworkRule{},
			wantErr: errors.New(errNotPostgreSQLServerVirtualNetworkRule),
		},
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ postgresql.VirtualNetworkRule) (postgresql.VirtualNetworkRulesCreateOrUpdateFuture, error) {
					return postgresql.VirtualNetworkRulesCreateOrUpdateFuture{}, nil
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ postgresql.VirtualNetworkRule) (postgresql.VirtualNetworkRulesCreateOrUpdateFuture, error) {
					return postgresql.VirtualNetworkRulesCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Creating()),
			),
			wantErr: errors.Wrap(errorBoom, errCreatePostgreSQLServerVirtualNetworkRule),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Create(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Create(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotPostgreSQLServerlVirtualNetworkRule",
			e:       &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{}},
			r:       &v1alpha3.MySQLServerVirtualNetworkRule{},
			want:    &v1alpha3.MySQLServerVirtualNetworkRule{},
			wantErr: errors.New(errNotPostgreSQLServerVirtualNetworkRule),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result postgresql.VirtualNetworkRule, err error) {
					return postgresql.VirtualNetworkRule{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			r:    virtualNetworkRule(),
			want: virtualNetworkRule(),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result postgresql.VirtualNetworkRule, err error) {
					return postgresql.VirtualNetworkRule{
						ID:   azure.ToStringPtr(resourceID),
						Type: azure.ToStringPtr(resourceType),
						VirtualNetworkRuleProperties: &postgresql.VirtualNetworkRuleProperties{
							VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
							IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(true),
							State:                            postgresql.VirtualNetworkRuleStateReady,
						},
					}, nil
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Available()),
				withState(string(postgresql.Ready)),
				withType(resourceType),
				withID(resourceID),
			),
		},
		{
			name: "FailedObserve",
			e: &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result postgresql.VirtualNetworkRule, err error) {
					return postgresql.VirtualNetworkRule{}, errorBoom
				},
			}},
			r:       virtualNetworkRule(),
			want:    virtualNetworkRule(),
			wantErr: errors.Wrap(errorBoom, errGetPostgreSQLServerVirtualNetworkRule),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Observe(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotPostgreSQLServerlVirtualNetworkRule",
			e:       &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{}},
			r:       &v1alpha3.MySQLServerVirtualNetworkRule{},
			want:    &v1alpha3.MySQLServerVirtualNetworkRule{},
			wantErr: errors.New(errNotPostgreSQLServerVirtualNetworkRule),
		},
		{
			name: "UnsuccessfulUpdate",
			e: &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ postgresql.VirtualNetworkRule) (postgresql.VirtualNetworkRulesCreateOrUpdateFuture, error) {
					return postgresql.VirtualNetworkRulesCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r:       virtualNetworkRule(),
			want:    virtualNetworkRule(),
			wantErr: errors.Wrap(errorBoom, errUpdatePostgreSQLServerVirtualNetworkRule),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Update(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Update(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotPostgreSQLServerlVirtualNetworkRule",
			e:       &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{}},
			r:       &v1alpha3.MySQLServerVirtualNetworkRule{},
			want:    &v1alpha3.MySQLServerVirtualNetworkRule{},
			wantErr: errors.New(errNotPostgreSQLServerVirtualNetworkRule),
		},
		{
			name: "Successful",
			e: &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{
				MockDelete: func(_ context.Context, _ string, _ string, _ string) (result postgresql.VirtualNetworkRulesDeleteFuture, err error) {
					return postgresql.VirtualNetworkRulesDeleteFuture{}, nil
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{
				MockDelete: func(_ context.Context, _ string, _ string, _ string) (result postgresql.VirtualNetworkRulesDeleteFuture, err error) {
					return postgresql.VirtualNetworkRulesDeleteFuture{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "Failed",
			e: &external{client: &fake.MockPostgreSQLVirtualNetworkRulesClient{
				MockDelete: func(_ context.Context, _ string, _ string, _ string) (result postgresql.VirtualNetworkRulesDeleteFuture, err error) {
					return postgresql.VirtualNetworkRulesDeleteFuture{}, errorBoom
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Deleting()),
			),
			wantErr: errors.Wrap(errorBoom, errDeletePostgreSQLServerVirtualNetworkRule),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.e.Delete(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
