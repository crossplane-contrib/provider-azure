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

package cosmosdb

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	cosmosdbclient "github.com/crossplane/provider-azure/pkg/clients/database/cosmosdb"
)

const (
	uid               = types.UID("definitely-a-uuid")
	id                = "myid"
	name              = "mycosmosaccount"
	resourcegroupname = "cool-rg"
	location          = "coolplace"
	kind              = "mongodb"

	stateSucceeded = "Succeeded"
)

type cosmosDBAccountModifier func(*v1alpha3.CosmosDBAccount)

// MockClient is a fake implementation of the azure cosmosdb client.
type MockClient struct {
	cosmosdbclient.AccountClient

	MockCreateOrUpdate  func(ctx context.Context, resourceGroupName string, accountName string, createUpdateParameters documentdb.DatabaseAccountCreateUpdateParameters) (result documentdb.DatabaseAccountsCreateOrUpdateFuture, err error)
	MockCheckNameExists func(ctx context.Context, accountName string) (result autorest.Response, err error)
	MockGet             func(ctx context.Context, resourceGroupName string, accountName string) (result documentdb.DatabaseAccount, err error)
	MockDelete          func(ctx context.Context, resourceGroupName string, accountName string) (result documentdb.DatabaseAccountsDeleteFuture, err error)
}

// CreateOrUpdate calls the underlying MockCreateOrUpdate method.
func (m *MockClient) CreateOrUpdate(ctx context.Context, resourceGroupName string, accountName string, createUpdateParameters documentdb.DatabaseAccountCreateUpdateParameters) (result documentdb.DatabaseAccountsCreateOrUpdateFuture, err error) {
	return m.MockCreateOrUpdate(ctx, resourceGroupName, accountName, createUpdateParameters)
}

// CheckExistence calls the underlying MockCheckExistence method.
func (m *MockClient) CheckNameExists(ctx context.Context, accountName string) (result autorest.Response, err error) {
	return m.MockCheckNameExists(ctx, accountName)
}

// CheckExistence calls the underlying MockCheckExistence method.
func (m *MockClient) Get(ctx context.Context, resourceGroupName string, accountName string) (result documentdb.DatabaseAccount, err error) {
	return m.MockGet(ctx, resourceGroupName, accountName)
}

// Delete calls the underlying MockDeleteGroup method.
func (m *MockClient) Delete(ctx context.Context, resourceGroupName string, accountName string) (result documentdb.DatabaseAccountsDeleteFuture, err error) {
	return m.MockDelete(ctx, resourceGroupName, accountName)
}

func withConditions(c ...xpv1.Condition) cosmosDBAccountModifier {
	return func(r *v1alpha3.CosmosDBAccount) { r.Status.ConditionedStatus.Conditions = c }
}

func cosmosDBAccount(rm ...cosmosDBAccountModifier) *v1alpha3.CosmosDBAccount {
	r := &v1alpha3.CosmosDBAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.CosmosDBAccountSpec{
			ForProvider: v1alpha3.CosmosDBAccountParameters{
				Location:          location,
				Kind:              kind,
				ResourceGroupName: resourcegroupname,
				Properties: v1alpha3.CosmosDBAccountProperties{
					Locations: []v1alpha3.CosmosDBAccountLocation{
						{
							LocationName:     location,
							IsZoneRedundant:  true,
							FailoverPriority: 0,
						},
					},
				},
			},
		},
		Status: v1alpha3.CosmosDBAccountStatus{
			AtProvider: &v1alpha3.CosmosDBAccountObservation{
				State: stateSucceeded,
				ID:    id,
			},
		},
	}

	meta.SetExternalName(r, name)

	for _, m := range rm {
		m(r)
	}

	return r
}

func TestObserve(t *testing.T) {
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		mg  resource.Managed
		err error
	}

	mockKube := &test.MockClient{
		MockGet: func(_ context.Context, key client.ObjectKey, obj client.Object) error {
			return nil
		},
		MockUpdate: func(ctx context.Context, obj client.Object, opts ...client.UpdateOption) error {
			return nil
		},
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"NotCosmosDBAccount": {
			e: &external{},
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotNoSQLAccount),
			},
		},
		"CheckExistenceError": {
			e: &external{
				kube: mockKube,
				client: &MockClient{
					MockCheckNameExists: func(_ context.Context, _ string) (result autorest.Response, err error) {
						return autorest.Response{}, errBoom
					},
				},
			},
			args: args{
				mg: cosmosDBAccount(),
			},
			want: want{
				mg:  cosmosDBAccount(),
				err: errors.Wrap(errBoom, errGetNoSQLAccount),
			},
		},
		"AccountNotFound": {
			e: &external{
				kube: mockKube,
				client: &MockClient{
					MockCheckNameExists: func(_ context.Context, _ string) (result autorest.Response, err error) {
						return autorest.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, nil
					},
				},
			},
			args: args{
				mg: cosmosDBAccount(),
			},
			want: want{
				o:  managed.ExternalObservation{ResourceExists: false},
				mg: cosmosDBAccount(),
			},
		},
		"Success": {
			e: &external{
				kube: mockKube,
				client: &MockClient{
					MockCheckNameExists: func(_ context.Context, _ string) (result autorest.Response, err error) {
						return autorest.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					},
					MockGet: func(_ context.Context, _ string, _ string) (result documentdb.DatabaseAccount, err error) {
						return documentdb.DatabaseAccount{
							ID:       azure.ToStringPtr(id),
							Kind:     kind,
							Location: azure.ToStringPtr(location),
							DatabaseAccountProperties: &documentdb.DatabaseAccountProperties{
								ProvisioningState: azure.ToStringPtr(stateSucceeded),
								ReadLocations: &[]documentdb.Location{
									{
										LocationName:     azure.ToStringPtr(location),
										FailoverPriority: azure.ToInt32Ptr(0, azure.FieldRequired),
										IsZoneRedundant:  azure.ToBoolPtr(true),
									},
								},
							},
						}, nil
					},
				},
			},
			args: args{
				mg: cosmosDBAccount(),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				mg: cosmosDBAccount(
					withConditions(xpv1.Available())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.e.Observe(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("tc.e.Observe(...): -want managed, +got managed:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.o, got); diff != "" {
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
		c   managed.ExternalCreation
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"NotCosmosDBAccount": {
			e: &external{},
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotNoSQLAccount),
			},
		},
		"CreateOrUpdateError": {
			e: &external{
				client: &MockClient{
					MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ documentdb.DatabaseAccountCreateUpdateParameters) (result documentdb.DatabaseAccountsCreateOrUpdateFuture, err error) {
						return documentdb.DatabaseAccountsCreateOrUpdateFuture{}, errBoom
					},
				},
			},
			args: args{
				mg: cosmosDBAccount(),
			},
			want: want{
				mg:  cosmosDBAccount(withConditions(xpv1.Creating())),
				err: errors.Wrap(errBoom, errCreateNoSQLAccount),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.e.Create(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Create(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("tc.e.Create(...): -want managed, +got managed:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.c, got); diff != "" {
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

	type want struct {
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"NotResourceGroup": {
			e: &external{},
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotNoSQLAccount),
			},
		},
		"DeleteError": {
			e: &external{
				client: &MockClient{
					MockDelete: func(_ context.Context, _ string, _ string) (result documentdb.DatabaseAccountsDeleteFuture, err error) {
						return documentdb.DatabaseAccountsDeleteFuture{}, errBoom
					},
				},
			},
			args: args{
				mg: cosmosDBAccount(),
			},
			want: want{
				mg:  cosmosDBAccount(withConditions(xpv1.Deleting())),
				err: errors.Wrap(errBoom, errDeleteNoSQLAccount),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.e.Delete(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}
