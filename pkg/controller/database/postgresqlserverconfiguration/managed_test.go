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

package postgresqlserverconfiguration

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-azure/apis/database/v1beta1"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
)

const (
	inProgressResponse = `{"status": "InProgress"}`
	subscriptID        = "subscription-id"
)

type MockPostgreSQLConfigurationAPI struct {
	MockGet            func(ctx context.Context, s *v1beta1.PostgreSQLServerConfiguration) (postgresql.Configuration, error)
	MockCreateOrUpdate func(ctx context.Context, s *v1beta1.PostgreSQLServerConfiguration) error
	MockDelete         func(ctx context.Context, s *v1beta1.PostgreSQLServerConfiguration) error
	MockGetRESTClient  func() autorest.Sender
}

func (m *MockPostgreSQLConfigurationAPI) Get(ctx context.Context, s *v1beta1.PostgreSQLServerConfiguration) (postgresql.Configuration, error) {
	return m.MockGet(ctx, s)
}

func (m *MockPostgreSQLConfigurationAPI) CreateOrUpdate(ctx context.Context, s *v1beta1.PostgreSQLServerConfiguration) error {
	return m.MockCreateOrUpdate(ctx, s)
}

func (m *MockPostgreSQLConfigurationAPI) Delete(ctx context.Context, s *v1beta1.PostgreSQLServerConfiguration) error {
	return m.MockDelete(ctx, s)
}

func (m *MockPostgreSQLConfigurationAPI) GetRESTClient() autorest.Sender {
	return m.MockGetRESTClient()
}

type modifier func(configuration *v1beta1.PostgreSQLServerConfiguration)

func withLastOperation(op azurev1alpha3.AsyncOperation) modifier {
	return func(p *v1beta1.PostgreSQLServerConfiguration) {
		p.Status.AtProvider.LastOperation = op
	}
}

func withExternalName(name string) modifier {
	return func(p *v1beta1.PostgreSQLServerConfiguration) {
		meta.SetExternalName(p, name)
	}
}

func postgresqlserverconfiguration(m ...modifier) *v1beta1.PostgreSQLServerConfiguration {
	p := &v1beta1.PostgreSQLServerConfiguration{}

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
		"ErrNotAPostgreSQLServerConfiguration": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotPostgreSQLServerConfig),
			},
		},
		"ErrGetServer": {
			e: &external{
				client: &MockPostgreSQLConfigurationAPI{
					MockGet: func(_ context.Context, _ *v1beta1.PostgreSQLServerConfiguration) (postgresql.Configuration, error) {
						return postgresql.Configuration{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserverconfiguration(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetPostgreSQLServerConfig),
			},
		},
		"ServerCreating": {
			e: &external{
				client: &MockPostgreSQLConfigurationAPI{
					MockGet: func(_ context.Context, _ *v1beta1.PostgreSQLServerConfiguration) (postgresql.Configuration, error) {
						return postgresql.Configuration{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockGetRESTClient: func() autorest.Sender {
						return autorest.SenderFunc(func(req *http.Request) (*http.Response, error) {
							return &http.Response{
								Request:       req,
								StatusCode:    http.StatusAccepted,
								Body:          ioutil.NopCloser(strings.NewReader(inProgressResponse)),
								ContentLength: int64(len([]byte(inProgressResponse))),
							}, nil
						})
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserverconfiguration(withLastOperation(azurev1alpha3.AsyncOperation{Method: http.MethodPut, PollingURL: "crossplane.io"})),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"ServerNotFound": {
			e: &external{
				client: &MockPostgreSQLConfigurationAPI{
					MockGet: func(_ context.Context, _ *v1beta1.PostgreSQLServerConfiguration) (postgresql.Configuration, error) {
						return postgresql.Configuration{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockGetRESTClient: func() autorest.Sender {
						return nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserverconfiguration(),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"ServerAvailable": {
			e: &external{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				client: &MockPostgreSQLConfigurationAPI{
					MockGet: func(_ context.Context, _ *v1beta1.PostgreSQLServerConfiguration) (postgresql.Configuration, error) {
						return postgresql.Configuration{
							ConfigurationProperties: &postgresql.ConfigurationProperties{},
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
				mg: postgresqlserverconfiguration(
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
		"ErrNotAPostgreSQLServerConfiguration": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotPostgreSQLServerConfig),
			},
		},
		"ErrCreateServer": {
			e: &external{
				client: &MockPostgreSQLConfigurationAPI{
					MockCreateOrUpdate: func(_ context.Context, _ *v1beta1.PostgreSQLServerConfiguration) error { return errBoom },
				},
				subscriptionID: subscriptID,
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserverconfiguration(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreatePostgreSQLServerConfig),
			},
		},
		"Successful": {
			e: &external{
				client: &MockPostgreSQLConfigurationAPI{
					MockCreateOrUpdate: func(_ context.Context, _ *v1beta1.PostgreSQLServerConfiguration) error { return nil },
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
				mg:  postgresqlserverconfiguration(),
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
		"ErrNotAPostgreSQLServerConfiguration": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotPostgreSQLServerConfig),
		},
		"ErrDeleteServer": {
			e: &external{
				client: &MockPostgreSQLConfigurationAPI{
					MockDelete: func(_ context.Context, _ *v1beta1.PostgreSQLServerConfiguration) error { return errBoom },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserverconfiguration(),
			},
			want: errors.Wrap(errBoom, errDeletePostgreSQLServerConfig),
		},
		"Successful": {
			e: &external{
				client: &MockPostgreSQLConfigurationAPI{
					MockDelete: func(_ context.Context, _ *v1beta1.PostgreSQLServerConfiguration) error { return nil },
					MockGetRESTClient: func() autorest.Sender {
						return autorest.SenderFunc(func(*http.Request) (*http.Response, error) {
							return nil, nil
						})
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserverconfiguration(),
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
