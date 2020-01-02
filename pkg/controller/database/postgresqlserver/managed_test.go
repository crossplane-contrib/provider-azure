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

package postgresqlserver

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/crossplaneio/stack-azure/pkg/clients/database"

	"github.com/crossplaneio/crossplane-runtime/pkg/meta"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/postgresql/mgmt/postgresql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	"github.com/crossplaneio/stack-azure/apis/database/v1beta1"
	azurev1alpha3 "github.com/crossplaneio/stack-azure/apis/v1alpha3"
)

var (
	_ resource.ExternalClient    = &external{}
	_ resource.ExternalConnecter = &connecter{}
)

type MockPostgreSQLServerAPI struct {
	MockServerNameTaken func(ctx context.Context, s *v1beta1.PostgreSQLServer) (bool, error)
	MockGetServer       func(ctx context.Context, s *v1beta1.PostgreSQLServer) (postgresql.Server, error)
	MockCreateServer    func(ctx context.Context, s *v1beta1.PostgreSQLServer, adminPassword string) error
	MockDeleteServer    func(ctx context.Context, s *v1beta1.PostgreSQLServer) error
	MockUpdateServer    func(ctx context.Context, s *v1beta1.PostgreSQLServer) error
	MockGetRESTClient   func() autorest.Sender
}

func (m *MockPostgreSQLServerAPI) GetRESTClient() autorest.Sender {
	return m.MockGetRESTClient()
}

func (m *MockPostgreSQLServerAPI) ServerNameTaken(ctx context.Context, s *v1beta1.PostgreSQLServer) (bool, error) {
	return m.MockServerNameTaken(ctx, s)
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

func withProviderRef(r *corev1.ObjectReference) modifier {
	return func(p *v1beta1.PostgreSQLServer) {
		p.Spec.ProviderReference = r
	}
}

func withAdminName(name string) modifier {
	return func(p *v1beta1.PostgreSQLServer) {
		p.Spec.ForProvider.AdministratorLogin = name
	}
}

func postgresqlserver(m ...modifier) *v1beta1.PostgreSQLServer {
	p := &v1beta1.PostgreSQLServer{}

	for _, mod := range m {
		mod(p)
	}
	return p
}

func TestConnect(t *testing.T) {
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	cases := map[string]struct {
		ec   resource.ExternalConnecter
		args args
		want error
	}{
		"ErrNotAPostgreSQLServer": {
			ec: &connecter{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotPostgreSQLServer),
		},
		"ErrGetProvider": {
			ec: &connecter{
				client: &test.MockClient{MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
					switch obj.(type) {
					case *azurev1alpha3.Provider:
						return errBoom
					default:
						return errors.New("unexpected type")
					}
				})},
				newClientFn: func(credentials []byte) (database.PostgreSQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(withProviderRef(&corev1.ObjectReference{})),
			},
			want: errors.Wrap(errBoom, errGetProvider),
		},
		"ErrGetProviderSecret": {
			ec: &connecter{
				client: &test.MockClient{MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
					switch obj.(type) {
					case *azurev1alpha3.Provider:
						return nil
					case *corev1.Secret:
						return errBoom
					default:
						return errors.New("unexpected type")
					}
				})},
				newClientFn: func(credentials []byte) (database.PostgreSQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(withProviderRef(&corev1.ObjectReference{})),
			},
			want: errors.Wrap(errBoom, errGetProviderSecret),
		},
		"Successful": {
			ec: &connecter{
				client:      &test.MockClient{MockGet: test.NewMockGetFn(nil)},
				newClientFn: func(credentials []byte) (database.PostgreSQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(withProviderRef(&corev1.ObjectReference{})),
			},
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, got := tc.ec.Connect(tc.args.ctx, tc.args.mg)
			if diff := cmp.Diff(tc.want, got, test.EquateErrors()); diff != "" {
				t.Errorf("-want error, +got error:\n%s", diff)
			}
		})
	}
}

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
		eo  resource.ExternalObservation
		err error
	}

	cases := map[string]struct {
		e    resource.ExternalClient
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
		"ErrCheckServerName": {
			e: &external{
				client: &MockPostgreSQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.PostgreSQLServer) (postgresql.Server, error) {
						return postgresql.Server{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockServerNameTaken: func(_ context.Context, _ *v1beta1.PostgreSQLServer) (bool, error) {
						return false, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCheckPostgreSQLServerName),
			},
		},
		"ServerCreating": {
			e: &external{
				client: &MockPostgreSQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.PostgreSQLServer) (postgresql.Server, error) {
						return postgresql.Server{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockServerNameTaken: func(_ context.Context, _ *v1beta1.PostgreSQLServer) (bool, error) {
						return true, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
			},
			want: want{
				eo: resource.ExternalObservation{
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
					MockServerNameTaken: func(_ context.Context, _ *v1beta1.PostgreSQLServer) (bool, error) {
						return false, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
			},
			want: want{
				eo: resource.ExternalObservation{
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
				eo: resource.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: resource.ConnectionDetails{
						runtimev1alpha1.ResourceCredentialsSecretEndpointKey: []byte(endpoint),
						runtimev1alpha1.ResourceCredentialsSecretUserKey:     []byte(fmt.Sprintf("%s@%s", admin, name)),
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
		ec  resource.ExternalCreation
		err error
	}

	cases := map[string]struct {
		e    resource.ExternalClient
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
				newPasswordFn: func(int) (string, error) { return "", errBoom },
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
				newPasswordFn: func(int) (string, error) { return password, nil },
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
				newPasswordFn: func(int) (string, error) { return password, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(),
			},
			want: want{
				ec: resource.ExternalCreation{
					ConnectionDetails: resource.ConnectionDetails{runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(password)},
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
		e    resource.ExternalClient
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
