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
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/database"
	"github.com/crossplane/provider-azure/pkg/clients/fake"
)

const (
	namespace         = "coolNamespace"
	name              = "coolSubnet"
	uid               = types.UID("definitely-a-uuid")
	serverName        = "coolSrv"
	resourceGroupName = "coolRG"
	resourceID        = "a-very-cool-id"
	resourceType      = "cooltype"

	providerName       = "cool-aws"
	providerSecretName = "cool-aws-secret"
	providerSecretKey  = "credentials"
	providerSecretData = "definitelyini"
)

var (
	provider = azurev1alpha3.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: providerName},
		Spec: azurev1alpha3.ProviderSpec{
			ProviderSpec: runtimev1alpha1.ProviderSpec{
				CredentialsSecretRef: &runtimev1alpha1.SecretKeySelector{
					SecretReference: runtimev1alpha1.SecretReference{
						Namespace: namespace,
						Name:      providerSecretName,
					},
					Key: providerSecretKey,
				},
			},
		},
	}

	providerSecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data:       map[string][]byte{providerSecretKey: []byte(providerSecretData)},
	}
)

type firewallRuleModifier func(*v1alpha3.PostgreSQLServerFirewallRule)

func withConditions(c ...runtimev1alpha1.Condition) firewallRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerFirewallRule) { r.Status.ConditionedStatus.Conditions = c }
}

func withType(s string) firewallRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerFirewallRule) { r.Status.AtProvider.Type = s }
}

func withID(s string) firewallRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerFirewallRule) { r.Status.AtProvider.ID = s }
}

func withProviderRef(p *corev1.ObjectReference) firewallRuleModifier {
	return func(r *v1alpha3.PostgreSQLServerFirewallRule) { r.Spec.ProviderReference = p }
}

func firewallRule(sm ...firewallRuleModifier) *v1alpha3.PostgreSQLServerFirewallRule {
	r := &v1alpha3.PostgreSQLServerFirewallRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.FirewallRuleSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Namespace: namespace, Name: providerName},
			},
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

func TestConnect(t *testing.T) {
	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	errBoom := errors.New("boom")

	cases := map[string]struct {
		conn managed.ExternalConnecter
		args args
		want error
	}{
		"NotPostgreSQLFirewallRule": {
			conn: &connecter{client: &test.MockClient{}},
			want: errors.New(errNotPostgreSQLServerFirewallRule),
		},
		"GetProviderError": {
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						return errBoom
					},
				},
				newClientFn: func(_ context.Context, _ []byte) (database.PostgreSQLFirewallRulesClient, error) {
					return &fake.MockPostgreSQLFirewallRulesClient{}, nil
				},
			},
			args: args{
				mg: firewallRule(withProviderRef(&corev1.ObjectReference{Name: providerName})),
			},
			want: errors.Wrapf(errBoom, "cannot get provider %s", providerName),
		},
		"GetProviderSecretError": {
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							*obj.(*azurev1alpha3.Provider) = provider
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							return errBoom
						}
						return nil
					},
				},
				newClientFn: func(_ context.Context, _ []byte) (database.PostgreSQLFirewallRulesClient, error) {
					return &fake.MockPostgreSQLFirewallRulesClient{}, nil
				},
			},
			args: args{
				mg: firewallRule(withProviderRef(&corev1.ObjectReference{Name: providerName})),
			},
			want: errors.Wrapf(errBoom, "cannot get provider secret %s", fmt.Sprintf("%s/%s", namespace, providerSecretName)),
		},
		"GetProviderSecretNil": {
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							nilSecretProvider := provider
							nilSecretProvider.SetCredentialsSecretReference(nil)
							*obj.(*azurev1alpha3.Provider) = nilSecretProvider
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							return errBoom
						}
						return nil
					},
				},
				newClientFn: func(_ context.Context, _ []byte) (database.PostgreSQLFirewallRulesClient, error) {
					return &fake.MockPostgreSQLFirewallRulesClient{}, nil
				},
			},
			args: args{
				mg: firewallRule(withProviderRef(&corev1.ObjectReference{Name: providerName})),
			},
			want: errors.New(errProviderSecretNil),
		},
		"Successful": {
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							*obj.(*azurev1alpha3.Provider) = provider
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							*obj.(*corev1.Secret) = providerSecret
						}
						return nil
					},
				},
				newClientFn: func(_ context.Context, _ []byte) (database.PostgreSQLFirewallRulesClient, error) {
					return &fake.MockPostgreSQLFirewallRulesClient{}, nil
				},
			},
			args: args{
				mg: firewallRule(withProviderRef(&corev1.ObjectReference{Name: providerName})),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, got := tc.conn.Connect(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, got, test.EquateErrors()); diff != "" {
				t.Errorf("tc.conn.Connect(...): want error != got error:\n%s", diff)
			}
		})
	}
}

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
					withConditions(runtimev1alpha1.Available()),
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
					withConditions(runtimev1alpha1.Creating()),
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
					withConditions(runtimev1alpha1.Creating()),
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
					withConditions(runtimev1alpha1.Deleting()),
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
					withConditions(runtimev1alpha1.Deleting()),
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
					withConditions(runtimev1alpha1.Deleting()),
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
