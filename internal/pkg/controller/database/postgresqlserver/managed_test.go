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

package postgresqlserver

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/postgresql/mgmt/2017-12-01/postgresql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients/database"

	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/database/v1beta1"
	azurev1alpha3 "github.com/crossplane-contrib/provider-jet-azure/apis/classic/v1alpha3"
)

var (
	_ managed.ExternalClient       = &external{}
	_ managed.ExternalConnecter    = &connecter{}
	_ database.PostgreSQLServerAPI = &MockPostgreSQLServerAPI{}
)

type MockPostgreSQLServerAPI struct {
	MockGetServer     func(ctx context.Context, s *v1beta1.PostgreSQLServer) (postgresql.Server, error)
	MockCreateServer  func(ctx context.Context, s *v1beta1.PostgreSQLServer, adminPassword string) error
	MockDeleteServer  func(ctx context.Context, s *v1beta1.PostgreSQLServer) error
	MockUpdateServer  func(ctx context.Context, s *v1beta1.PostgreSQLServer) error
	MockGetRESTClient func() autorest.Sender
}

func (m *MockPostgreSQLServerAPI) GetRESTClient() autorest.Sender {
	return m.MockGetRESTClient()
}

func (m *MockPostgreSQLServerAPI) GetServer(ctx context.Context, s *v1beta1.PostgreSQLServer) (postgresql.Server, error) {
	return m.MockGetServer(ctx, s)
}

func (m *MockPostgreSQLServerAPI) CreateServer(ctx context.Context, s *v1beta1.PostgreSQLServer, adminPassword string) error {
	return m.MockCreateServer(ctx, s, adminPassword)
}

func (m *MockPostgreSQLServerAPI) UpdateServer(ctx context.Context, s *v1beta1.PostgreSQLServer) error {
	return m.MockUpdateServer(ctx, s)
}

func (m *MockPostgreSQLServerAPI) DeleteServer(ctx context.Context, s *v1beta1.PostgreSQLServer) error {
	return m.MockDeleteServer(ctx, s)
}

type modifier func(*v1beta1.PostgreSQLServer)

func withExternalName(name string) modifier {
	return func(p *v1beta1.PostgreSQLServer) {
		meta.SetExternalName(p, name)
	}
}

func withAdminName(name string) modifier {
	return func(p *v1beta1.PostgreSQLServer) {
		p.Spec.ForProvider.AdministratorLogin = name
	}
}

func withLastOperation(op azurev1alpha3.AsyncOperation) modifier {
	return func(p *v1beta1.PostgreSQLServer) {
		p.Status.AtProvider.LastOperation = op
	}
}

func postgresqlserver(m ...modifier) *v1beta1.PostgreSQLServer {
	p := &v1beta1.PostgreSQLServer{}

	for _, mod := range m {
		mod(p)
	}
	return p
}

const (
	inProgressResponse = `{"status": "InProgress"}`
)

func TestObserve(t *testing.T) {
	errBoom := errors.New("boom")
	name := "coolserver"
	endpoint := "coolazure.example.prg"
	admin := "cooladmin"

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
		"ErrNotAPostgreSQLServer": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotPostgreSQLServer),
			},
		},
		"ErrGetServer": {
			e: &external{
				client: &MockPostgreSQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.PostgreSQLServer) (postgresql.Server, error) {
						return postgresql.Server{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetPostgreSQLServer),
			},
		},
		"ServerCreating": {
			e: &external{
				client: &MockPostgreSQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.PostgreSQLServer) (postgresql.Server, error) {
						return postgresql.Server{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
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
				mg:  postgresqlserver(withLastOperation(azurev1alpha3.AsyncOperation{Method: http.MethodPut, PollingURL: "crossplane.io"})),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"ServerNotFound": {
			e: &external{
				client: &MockPostgreSQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.PostgreSQLServer) (postgresql.Server, error) {
						return postgresql.Server{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockGetRESTClient: func() autorest.Sender {
						return nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
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
				client: &MockPostgreSQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.PostgreSQLServer) (postgresql.Server, error) {
						return postgresql.Server{
							Sku: &postgresql.Sku{},
							ServerProperties: &postgresql.ServerProperties{
								UserVisibleState:         postgresql.ServerStateReady,
								FullyQualifiedDomainName: &endpoint,
								StorageProfile:           &postgresql.StorageProfile{},
							}}, nil
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
				mg: postgresqlserver(
					withExternalName(name),
					withAdminName(admin),
				),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretEndpointKey: []byte(endpoint),
						xpv1.ResourceCredentialsSecretUserKey:     []byte(fmt.Sprintf("%s@%s", admin, name)),
						xpv1.ResourceCredentialsSecretPortKey:     []byte(v1beta1.PostgreSQLServerPort),
					},
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
	password := "verysecure"

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
		"ErrNotAPostgreSQLServer": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotPostgreSQLServer),
			},
		},
		"ErrGeneratePassword": {
			e: &external{
				newPasswordFn: func() (string, error) { return "", errBoom },
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGenPassword),
			},
		},
		"ErrCreateServer": {
			e: &external{
				client: &MockPostgreSQLServerAPI{
					MockCreateServer: func(_ context.Context, _ *v1beta1.PostgreSQLServer, _ string) error { return errBoom },
				},
				newPasswordFn: func() (string, error) { return password, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreatePostgreSQLServer),
			},
		},
		"Successful": {
			e: &external{
				client: &MockPostgreSQLServerAPI{
					MockCreateServer: func(_ context.Context, _ *v1beta1.PostgreSQLServer, _ string) error { return nil },
					MockGetRESTClient: func() autorest.Sender {
						return autorest.SenderFunc(func(*http.Request) (*http.Response, error) {
							return nil, nil
						})
					},
				},
				newPasswordFn: func() (string, error) { return password, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
			},
			want: want{
				ec: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{xpv1.ResourceCredentialsSecretPasswordKey: []byte(password)},
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
		"ErrNotAPostgreSQLServer": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotPostgreSQLServer),
		},
		"ErrDeleteServer": {
			e: &external{
				client: &MockPostgreSQLServerAPI{
					MockDeleteServer: func(_ context.Context, _ *v1beta1.PostgreSQLServer) error { return errBoom },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
			},
			want: errors.Wrap(errBoom, errDeletePostgreSQLServer),
		},
		"Successful": {
			e: &external{
				client: &MockPostgreSQLServerAPI{
					MockDeleteServer: func(_ context.Context, _ *v1beta1.PostgreSQLServer) error { return nil },
					MockGetRESTClient: func() autorest.Sender {
						return autorest.SenderFunc(func(*http.Request) (*http.Response, error) {
							return nil, nil
						})
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
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
