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

package mysqlservervirtualnetworkrule

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
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

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/fake"
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

type virtualNetworkRuleModifier func(*v1alpha3.MySQLServerVirtualNetworkRule)

func withConditions(c ...xpv1.Condition) virtualNetworkRuleModifier {
	return func(r *v1alpha3.MySQLServerVirtualNetworkRule) { r.Status.ConditionedStatus.Conditions = c }
}

func withType(s string) virtualNetworkRuleModifier {
	return func(r *v1alpha3.MySQLServerVirtualNetworkRule) { r.Status.Type = s }
}

func withID(s string) virtualNetworkRuleModifier {
	return func(r *v1alpha3.MySQLServerVirtualNetworkRule) { r.Status.ID = s }
}

func withState(s string) virtualNetworkRuleModifier {
	return func(r *v1alpha3.MySQLServerVirtualNetworkRule) { r.Status.State = s }
}

func virtualNetworkRule(sm ...virtualNetworkRuleModifier) *v1alpha3.MySQLServerVirtualNetworkRule {
	r := &v1alpha3.MySQLServerVirtualNetworkRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.MySQLVirtualNetworkRuleSpec{
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
			name:    "NotMysqServerlVirtualNetworkRule",
			e:       &external{client: &fake.MockMySQLVirtualNetworkRulesClient{}},
			r:       &v1alpha3.PostgreSQLServerVirtualNetworkRule{},
			want:    &v1alpha3.PostgreSQLServerVirtualNetworkRule{},
			wantErr: errors.New(errNotMySQLServerVirtualNetworkRule),
		},
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ mysql.VirtualNetworkRule) (mysql.VirtualNetworkRulesCreateOrUpdateFuture, error) {
					return mysql.VirtualNetworkRulesCreateOrUpdateFuture{}, nil
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ mysql.VirtualNetworkRule) (mysql.VirtualNetworkRulesCreateOrUpdateFuture, error) {
					return mysql.VirtualNetworkRulesCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Creating()),
			),
			wantErr: errors.Wrap(errorBoom, errCreateMySQLServerVirtualNetworkRule),
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
			name:    "NotMysqServerlVirtualNetworkRule",
			e:       &external{client: &fake.MockMySQLVirtualNetworkRulesClient{}},
			r:       &v1alpha3.PostgreSQLServerVirtualNetworkRule{},
			want:    &v1alpha3.PostgreSQLServerVirtualNetworkRule{},
			wantErr: errors.New(errNotMySQLServerVirtualNetworkRule),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result mysql.VirtualNetworkRule, err error) {
					return mysql.VirtualNetworkRule{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			r:    virtualNetworkRule(),
			want: virtualNetworkRule(),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result mysql.VirtualNetworkRule, err error) {
					return mysql.VirtualNetworkRule{
						ID:   azure.ToStringPtr(resourceID),
						Type: azure.ToStringPtr(resourceType),
						VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
							VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
							IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(true),
							State:                            mysql.VirtualNetworkRuleStateReady,
						},
					}, nil
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Available()),
				withState(string(mysql.Ready)),
				withType(resourceType),
				withID(resourceID),
			),
		},
		{
			name: "FailedObserve",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result mysql.VirtualNetworkRule, err error) {
					return mysql.VirtualNetworkRule{}, errorBoom
				},
			}},
			r:       virtualNetworkRule(),
			want:    virtualNetworkRule(),
			wantErr: errors.Wrap(errorBoom, errGetMySQLServerVirtualNetworkRule),
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
			name:    "NotMysqServerlVirtualNetworkRule",
			e:       &external{client: &fake.MockMySQLVirtualNetworkRulesClient{}},
			r:       &v1alpha3.PostgreSQLServerVirtualNetworkRule{},
			want:    &v1alpha3.PostgreSQLServerVirtualNetworkRule{},
			wantErr: errors.New(errNotMySQLServerVirtualNetworkRule),
		},
		{
			name: "SuccessfulDoesNotNeedUpdate",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result mysql.VirtualNetworkRule, err error) {
					return mysql.VirtualNetworkRule{
						VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
							VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
							IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(true),
						},
					}, nil
				},
			}},
			r:    virtualNetworkRule(),
			want: virtualNetworkRule(),
		},
		{
			name: "SuccessfulNeedsUpdate",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result mysql.VirtualNetworkRule, err error) {
					return mysql.VirtualNetworkRule{
						VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
							VirtualNetworkSubnetID:           azure.ToStringPtr("/wrong/subnet"),
							IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(true),
						},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ mysql.VirtualNetworkRule) (mysql.VirtualNetworkRulesCreateOrUpdateFuture, error) {
					return mysql.VirtualNetworkRulesCreateOrUpdateFuture{}, nil
				},
			}},
			r:    virtualNetworkRule(),
			want: virtualNetworkRule(),
		},
		{
			name: "UnsuccessfulGet",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result mysql.VirtualNetworkRule, err error) {
					return mysql.VirtualNetworkRule{
						VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
							VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
							IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(true),
						},
					}, errorBoom
				},
			}},
			r:       virtualNetworkRule(),
			want:    virtualNetworkRule(),
			wantErr: errors.Wrap(errorBoom, errGetMySQLServerVirtualNetworkRule),
		},
		{
			name: "UnsuccessfulUpdate",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result mysql.VirtualNetworkRule, err error) {
					return mysql.VirtualNetworkRule{
						VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
							VirtualNetworkSubnetID:           azure.ToStringPtr("wrong/subnet"),
							IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(true),
						},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ mysql.VirtualNetworkRule) (mysql.VirtualNetworkRulesCreateOrUpdateFuture, error) {
					return mysql.VirtualNetworkRulesCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r:       virtualNetworkRule(),
			want:    virtualNetworkRule(),
			wantErr: errors.Wrap(errorBoom, errUpdateMySQLServerVirtualNetworkRule),
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
			name:    "NotMysqServerlVirtualNetworkRule",
			e:       &external{client: &fake.MockMySQLVirtualNetworkRulesClient{}},
			r:       &v1alpha3.PostgreSQLServerVirtualNetworkRule{},
			want:    &v1alpha3.PostgreSQLServerVirtualNetworkRule{},
			wantErr: errors.New(errNotMySQLServerVirtualNetworkRule),
		},
		{
			name: "Successful",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockDelete: func(_ context.Context, _ string, _ string, _ string) (result mysql.VirtualNetworkRulesDeleteFuture, err error) {
					return mysql.VirtualNetworkRulesDeleteFuture{}, nil
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockDelete: func(_ context.Context, _ string, _ string, _ string) (result mysql.VirtualNetworkRulesDeleteFuture, err error) {
					return mysql.VirtualNetworkRulesDeleteFuture{}, autorest.DetailedError{
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
			e: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{
				MockDelete: func(_ context.Context, _ string, _ string, _ string) (result mysql.VirtualNetworkRulesDeleteFuture, err error) {
					return mysql.VirtualNetworkRulesDeleteFuture{}, errorBoom
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(xpv1.Deleting()),
			),
			wantErr: errors.Wrap(errorBoom, errDeleteMySQLServerVirtualNetworkRule),
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
