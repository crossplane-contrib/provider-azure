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

package mysqlserverconfiguration

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/database/v1beta1"
	azurev1alpha3 "github.com/crossplane-contrib/provider-jet-azure/apis/classic/v1alpha3"
)

const (
	inProgress  = "InProgress"
	subscriptID = "subscription-id"
)

type MockMySQLConfigurationAPI struct {
	MockGet            func(ctx context.Context, s *v1beta1.MySQLServerConfiguration) (mysql.Configuration, error)
	MockCreateOrUpdate func(ctx context.Context, s *v1beta1.MySQLServerConfiguration) error
	MockDelete         func(ctx context.Context, s *v1beta1.MySQLServerConfiguration) error
	MockGetRESTClient  func() autorest.Sender
}

func (m *MockMySQLConfigurationAPI) Get(ctx context.Context, s *v1beta1.MySQLServerConfiguration) (mysql.Configuration, error) {
	return m.MockGet(ctx, s)
}

func (m *MockMySQLConfigurationAPI) CreateOrUpdate(ctx context.Context, s *v1beta1.MySQLServerConfiguration) error {
	return m.MockCreateOrUpdate(ctx, s)
}

func (m *MockMySQLConfigurationAPI) Delete(ctx context.Context, s *v1beta1.MySQLServerConfiguration) error {
	return m.MockDelete(ctx, s)
}

func (m *MockMySQLConfigurationAPI) GetRESTClient() autorest.Sender {
	return m.MockGetRESTClient()
}

type modifier func(configuration *v1beta1.MySQLServerConfiguration)

func withLastOperation(op azurev1alpha3.AsyncOperation) modifier {
	return func(p *v1beta1.MySQLServerConfiguration) {
		p.Status.AtProvider.LastOperation = op
	}
}

func withExternalName(name string) modifier {
	return func(p *v1beta1.MySQLServerConfiguration) {
		meta.SetExternalName(p, name)
	}
}

func mysqlserverconfiguration(m ...modifier) *v1beta1.MySQLServerConfiguration {
	p := &v1beta1.MySQLServerConfiguration{}

	for _, mod := range m {
		mod(p)
	}
	return p
}

func TestObserve(t *testing.T) {
	errBoom := errors.New("boom")
	name := "coolserver"

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		eo  managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"ErrNotAMySQLServerConfiguration": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotMySQLServerConfig),
			},
		},
		"ErrGetServer": {
			e: &external{
				client: &MockMySQLConfigurationAPI{
					MockGet: func(_ context.Context, _ *v1beta1.MySQLServerConfiguration) (mysql.Configuration, error) {
						return mysql.Configuration{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserverconfiguration(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetMySQLServerConfig),
			},
		},
		"ServerCreating": {
			e: &external{
				client: &MockMySQLConfigurationAPI{
					MockGet: func(_ context.Context, _ *v1beta1.MySQLServerConfiguration) (mysql.Configuration, error) {
						return mysql.Configuration{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserverconfiguration(withLastOperation(azurev1alpha3.AsyncOperation{Method: http.MethodPut, PollingURL: "crossplane.io"})),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: errors.Wrap(autorest.DetailedError{StatusCode: http.StatusNotFound}, errNotFoundMySQLServerConfig),
			},
		},
		"ServerNotFound": {
			e: &external{
				client: &MockMySQLConfigurationAPI{
					MockGet: func(_ context.Context, _ *v1beta1.MySQLServerConfiguration) (mysql.Configuration, error) {
						return mysql.Configuration{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockGetRESTClient: func() autorest.Sender {
						return nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserverconfiguration(),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists: false,
				},
				err: errors.Wrap(autorest.DetailedError{StatusCode: http.StatusNotFound}, errNotFoundMySQLServerConfig),
			},
		},
		"ServerAvailable": {
			e: &external{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				client: &MockMySQLConfigurationAPI{
					MockGet: func(_ context.Context, _ *v1beta1.MySQLServerConfiguration) (mysql.Configuration, error) {
						return mysql.Configuration{
							ConfigurationProperties: &mysql.ConfigurationProperties{},
						}, nil
					},
					MockGetRESTClient: func() autorest.Sender {
						return autorest.SenderFunc(func(*http.Request) (*http.Response, error) {
							return nil, nil
						})
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg: mysqlserverconfiguration(
					withExternalName(name),
				),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eo, err := tc.e.Observe(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): -want error, +got error:\n%s", diff)
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
		"ErrNotAMySQLServerConfiguration": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotMySQLServerConfig),
			},
		},
		"ErrCreateServer": {
			e: &external{
				client: &MockMySQLConfigurationAPI{
					MockCreateOrUpdate: func(_ context.Context, _ *v1beta1.MySQLServerConfiguration) error { return errBoom },
				},
				subscriptionID: subscriptID,
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserverconfiguration(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateMySQLServerConfig),
			},
		},
		"Successful": {
			e: &external{
				client: &MockMySQLConfigurationAPI{
					MockCreateOrUpdate: func(_ context.Context, _ *v1beta1.MySQLServerConfiguration) error { return nil },
					MockGetRESTClient: func() autorest.Sender {
						return autorest.SenderFunc(func(*http.Request) (*http.Response, error) {
							return nil, nil
						})
					},
				},
				subscriptionID: subscriptID,
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserverconfiguration(),
			},
			want: want{
				ec: managed.ExternalCreation{
					ExternalNameAssigned: true,
				},
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
		"ErrNotAMySQLServerConfiguration": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotMySQLServerConfig),
		},
		"ErrDeleteServer": {
			e: &external{
				client: &MockMySQLConfigurationAPI{
					MockDelete: func(_ context.Context, _ *v1beta1.MySQLServerConfiguration) error { return errBoom },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserverconfiguration(),
			},
			want: errors.Wrap(errBoom, errDeleteMySQLServerConfig),
		},
		"Successful": {
			e: &external{
				client: &MockMySQLConfigurationAPI{
					MockDelete: func(_ context.Context, _ *v1beta1.MySQLServerConfiguration) error { return nil },
					MockGetRESTClient: func() autorest.Sender {
						return autorest.SenderFunc(func(*http.Request) (*http.Response, error) {
							return nil, nil
						})
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserverconfiguration(),
			},
			want: nil,
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

func TestUpdate(t *testing.T) {
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		eu  managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"ErrNotAMySQLServerConfiguration": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotMySQLServerConfig),
			},
		},
		"ServerUpdating": {
			e: &external{},
			args: args{
				ctx: context.Background(),
				mg: mysqlserverconfiguration(withLastOperation(azurev1alpha3.AsyncOperation{
					Method: http.MethodPatch, PollingURL: "crossplane.io", Status: inProgress})),
			},
			want: want{
				err: nil,
			},
		},
		"ErrUpdateServer": {
			e: &external{
				client: &MockMySQLConfigurationAPI{
					MockCreateOrUpdate: func(_ context.Context, _ *v1beta1.MySQLServerConfiguration) error { return errBoom },
				},
				subscriptionID: subscriptID,
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserverconfiguration(),
			},
			want: want{
				err: errors.Wrap(errBoom, errUpdateMySQLServerConfig),
			},
		},
		"Successful": {
			e: &external{
				client: &MockMySQLConfigurationAPI{
					MockCreateOrUpdate: func(_ context.Context, _ *v1beta1.MySQLServerConfiguration) error { return nil },
					MockGetRESTClient: func() autorest.Sender {
						return autorest.SenderFunc(func(*http.Request) (*http.Response, error) {
							return nil, nil
						})
					},
				},
				subscriptionID: subscriptID,
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserverconfiguration(),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eu, err := tc.e.Update(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Update(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.eu, eu); diff != "" {
				t.Errorf("tc.e.Update(...): -want, +got:\n%s", diff)
			}
		})
	}
}
