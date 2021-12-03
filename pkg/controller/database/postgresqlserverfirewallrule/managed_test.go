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

package postgresqlserverfirewallrule

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/fake"
)

const (
	name              = "coolSubnet"
	uid               = types.UID("definitely-a-uuid")
	serverName        = "coolSrv"
	resourceGroupName = "coolRG"
	resourceID        = "a-very-cool-id"
	resourceType      = "cooltype"
)

type firewallRuleModifier func(*v1alpha3.PostgreSQLServerFirewallRule)

func withConditions(c ...xpv1.Condition) firewallRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerFirewallRule) { r.Status.ConditionedStatus.Conditions = c }
}

func withType(s string) firewallRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerFirewallRule) { r.Status.AtProvider.Type = s }
}

func withID(s string) firewallRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerFirewallRule) { r.Status.AtProvider.ID = s }
}

func firewallRule(sm ...firewallRuleModifier) *v1alpha3.PostgreSQLServerFirewallRule {
	r := &v1alpha3.PostgreSQLServerFirewallRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.FirewallRuleSpec{
			ForProvider: v1alpha3.FirewallRuleParameters{
				ServerName:        serverName,
				ResourceGroupName: resourceGroupName,
				FirewallRuleProperties: v1alpha3.FirewallRuleProperties{
					StartIPAddress: "127.0.0.1",
					EndIPAddress:   "127.0.0.1",
				},
			},
		},
		Status: v1alpha3.FirewallRuleStatus{},
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

func TestObserve(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		mg  resource.Managed
		err error
	}

	errBoom := errors.New("boom")

	cases := map[string]struct {
		ec   managed.ExternalClient
		args args
		want want
	}{
		"NotPostgreSQLServerFirewallRule": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{}},
			want: want{
				err: errors.New(errNotPostgreSQLServerFirewallRule),
			},
		},
		"SuccessfulObserveNotExist": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result postgresql.FirewallRule, err error) {
					return postgresql.FirewallRule{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
				},
			}},
			args: args{
				mg: firewallRule(),
			},
			want: want{
				mg: firewallRule(),
			},
		},
		"SuccessfulObserveExists": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result postgresql.FirewallRule, err error) {
					return postgresql.FirewallRule{
						ID:                     azure.ToStringPtr(resourceID),
						Type:                   azure.ToStringPtr(resourceType),
						FirewallRuleProperties: &postgresql.FirewallRuleProperties{},
					}, nil
				},
			}},
			args: args{
				mg: firewallRule(),
			},
			want: want{
				mg: firewallRule(
					withConditions(xpv1.Available()),
					withType(resourceType),
					withID(resourceID),
				),
			},
		},
		"FailedObserve": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result postgresql.FirewallRule, err error) {
					return postgresql.FirewallRule{}, errBoom
				},
			}},
			args: args{
				mg: firewallRule(),
			},
			want: want{
				mg:  firewallRule(),
				err: errors.Wrap(errBoom, errGetPostgreSQLServerFirewallRule),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := tc.ec.Observe(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		mg  resource.Managed
		err error
	}

	errBoom := errors.New("boom")

	cases := map[string]struct {
		ec   managed.ExternalClient
		args args
		want want
	}{
		"NotPostgreSQLServerFirewallRule": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{}},
			want: want{
				err: errors.New(errNotPostgreSQLServerFirewallRule),
			},
		},
		"ErrorCreate": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ postgresql.FirewallRule) (postgresql.FirewallRulesCreateOrUpdateFuture, error) {
					return postgresql.FirewallRulesCreateOrUpdateFuture{}, errBoom
				},
			}},
			args: args{
				mg: firewallRule(),
			},
			want: want{
				mg: firewallRule(
					withConditions(xpv1.Creating()),
				),
				err: errors.Wrap(errBoom, errCreatePostgreSQLServerFirewallRule),
			},
		},
		"Successful": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ postgresql.FirewallRule) (postgresql.FirewallRulesCreateOrUpdateFuture, error) {
					return postgresql.FirewallRulesCreateOrUpdateFuture{}, nil
				},
			}},
			args: args{
				mg: firewallRule(),
			},
			want: want{
				mg: firewallRule(
					withConditions(xpv1.Creating()),
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := tc.ec.Create(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Create(...): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		mg  resource.Managed
		err error
	}

	errBoom := errors.New("boom")

	cases := map[string]struct {
		ec   managed.ExternalClient
		args args
		want want
	}{
		"NotPostgreSQLServerFirewallRule": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{}},
			want: want{
				err: errors.New(errNotPostgreSQLServerFirewallRule),
			},
		},
		"UpdateError": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result postgresql.FirewallRule, err error) {
					return postgresql.FirewallRule{
						FirewallRuleProperties: &postgresql.FirewallRuleProperties{},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ postgresql.FirewallRule) (postgresql.FirewallRulesCreateOrUpdateFuture, error) {
					return postgresql.FirewallRulesCreateOrUpdateFuture{}, errBoom
				},
			}},
			args: args{
				mg: firewallRule(),
			},
			want: want{
				mg:  firewallRule(),
				err: errors.Wrap(errBoom, errUpdatePostgreSQLServerFirewallRule),
			},
		},
		"Successful": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result postgresql.FirewallRule, err error) {
					return postgresql.FirewallRule{
						FirewallRuleProperties: &postgresql.FirewallRuleProperties{},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ postgresql.FirewallRule) (postgresql.FirewallRulesCreateOrUpdateFuture, error) {
					return postgresql.FirewallRulesCreateOrUpdateFuture{}, nil
				},
			}},

			args: args{
				mg: firewallRule(),
			},
			want: want{
				mg: firewallRule(),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := tc.ec.Update(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Update(...): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		mg  resource.Managed
		err error
	}

	errBoom := errors.New("boom")

	cases := map[string]struct {
		ec   managed.ExternalClient
		args args
		want want
	}{
		"NotPostgreSQLServerFirewallRule": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{}},
			want: want{
				err: errors.New(errNotPostgreSQLServerFirewallRule),
			},
		},
		"Successful": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{
				MockDelete: func(_ context.Context, _ string, _ string, _ string) (result postgresql.FirewallRulesDeleteFuture, err error) {
					return postgresql.FirewallRulesDeleteFuture{}, nil
				},
			}},
			args: args{
				mg: firewallRule(),
			},
			want: want{
				mg: firewallRule(
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"SuccessfulNotFound": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{
				MockDelete: func(_ context.Context, _ string, _ string, _ string) (result postgresql.FirewallRulesDeleteFuture, err error) {
					return postgresql.FirewallRulesDeleteFuture{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			args: args{
				mg: firewallRule(),
			},
			want: want{
				mg: firewallRule(
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"Failed": {
			ec: &external{client: &fake.MockPostgreSQLFirewallRulesClient{
				MockDelete: func(_ context.Context, _ string, _ string, _ string) (result postgresql.FirewallRulesDeleteFuture, err error) {
					return postgresql.FirewallRulesDeleteFuture{}, errBoom
				},
			}},
			args: args{
				mg: firewallRule(),
			},
			want: want{
				mg: firewallRule(
					withConditions(xpv1.Deleting()),
				),
				err: errors.Wrap(errBoom, errDeletePostgreSQLServerFirewallRule),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.ec.Delete(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
