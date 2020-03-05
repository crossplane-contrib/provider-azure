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
	"fmt"
	"net/http"
	"testing"

	"github.com/crossplane/provider-azure/pkg/clients/database"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/fake"
)

const (
	namespace         = "coolNamespace"
	name              = "coolSubnet"
	uid               = types.UID("definitely-a-uuid")
	serverName        = "coolVnet"
	resourceGroupName = "coolRG"
	vnetSubnetID      = "/the/best/subnet/ever"
	resourceID        = "a-very-cool-id"
	resourceType      = "cooltype"

	providerName       = "cool-aws"
	providerSecretName = "cool-aws-secret"
	providerSecretKey  = "credentials"
	providerSecretData = "definitelyini"
)

var (
	ctx       = context.Background()
	errorBoom = errors.New("boom")

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

type testCase struct {
	name    string
	e       managed.ExternalClient
	r       resource.Managed
	want    resource.Managed
	wantErr error
}

type virtualNetworkRuleModifier func(*v1alpha3.MySQLServerVirtualNetworkRule)

func withConditions(c ...runtimev1alpha1.Condition) virtualNetworkRuleModifier {
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

func withProviderRef(p *corev1.ObjectReference) virtualNetworkRuleModifier {
	return func(r *v1alpha3.MySQLServerVirtualNetworkRule) { r.Spec.ProviderReference = p }
}

func virtualNetworkRule(sm ...virtualNetworkRuleModifier) *v1alpha3.MySQLServerVirtualNetworkRule {
	r := &v1alpha3.MySQLServerVirtualNetworkRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.MySQLVirtualNetworkRuleSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Namespace: namespace, Name: providerName},
			},
			Name:              name,
			ServerName:        serverName,
			ResourceGroupName: resourceGroupName,
			VirtualNetworkRuleProperties: v1alpha3.VirtualNetworkRuleProperties{
				VirtualNetworkSubnetID:           vnetSubnetID,
				IgnoreMissingVnetServiceEndpoint: true,
			},
		},
		Status: v1alpha3.VirtualNetworkRuleStatus{},
	}

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
				withConditions(runtimev1alpha1.Creating()),
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
				withConditions(runtimev1alpha1.Creating()),
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
							State:                            mysql.Ready,
						},
					}, nil
				},
			}},
			r: virtualNetworkRule(),
			want: virtualNetworkRule(
				withConditions(runtimev1alpha1.Available()),
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
				withConditions(runtimev1alpha1.Deleting()),
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
				withConditions(runtimev1alpha1.Deleting()),
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
				withConditions(runtimev1alpha1.Deleting()),
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

func TestConnect(t *testing.T) {
	cases := []struct {
		name    string
		conn    *connecter
		i       resource.Managed
		want    managed.ExternalClient
		wantErr error
	}{
		{
			name:    "NotMySQLVirtualNetworkRule",
			conn:    &connecter{client: &test.MockClient{}},
			i:       &v1alpha3.PostgreSQLServerVirtualNetworkRule{},
			want:    nil,
			wantErr: errors.New(errNotMySQLServerVirtualNetworkRule),
		},
		{
			name: "SuccessfulConnect",
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
				newClientFn: func(_ context.Context, _ []byte) (database.MySQLVirtualNetworkRulesClient, error) {
					return &fake.MockMySQLVirtualNetworkRulesClient{}, nil
				},
			},
			i:    virtualNetworkRule(withProviderRef(&corev1.ObjectReference{Name: providerName})),
			want: &external{client: &fake.MockMySQLVirtualNetworkRulesClient{}},
		},
		{
			name: "GetProviderSecretFailed",
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							*obj.(*azurev1alpha3.Provider) = provider
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							return errorBoom
						}
						return nil
					},
				},
				newClientFn: func(_ context.Context, _ []byte) (database.MySQLVirtualNetworkRulesClient, error) {
					return &fake.MockMySQLVirtualNetworkRulesClient{}, nil
				},
			},
			i:       virtualNetworkRule(withProviderRef(&corev1.ObjectReference{Name: providerName})),
			wantErr: errors.Wrapf(errorBoom, "cannot get provider secret %s", fmt.Sprintf("%s/%s", namespace, providerSecretName)),
		},
		{
			name: "GetProviderSecretNil",
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							nilSecretProvider := provider
							nilSecretProvider.SetCredentialsSecretReference(nil)
							*obj.(*azurev1alpha3.Provider) = nilSecretProvider
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							return errorBoom
						}
						return nil
					},
				},
				newClientFn: func(_ context.Context, _ []byte) (database.MySQLVirtualNetworkRulesClient, error) {
					return &fake.MockMySQLVirtualNetworkRulesClient{}, nil
				},
			},
			i:       virtualNetworkRule(withProviderRef(&corev1.ObjectReference{Name: providerName})),
			wantErr: errors.New(errProviderSecretNil),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := tc.conn.Connect(ctx, tc.i)

			if diff := cmp.Diff(tc.wantErr, gotErr, test.EquateErrors()); diff != "" {
				t.Errorf("tc.conn.Connect(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, got, test.EquateConditions(), cmp.AllowUnexported(external{})); diff != "" {
				t.Errorf("tc.conn.Connect(...): -want, +got:\n%s", diff)
			}
		})
	}
}
