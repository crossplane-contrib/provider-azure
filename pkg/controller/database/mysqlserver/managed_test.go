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

package mysqlserver

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/mysql/mgmt/mysql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	"github.com/crossplaneio/stack-azure/apis/database/v1alpha2"
	azurev1alpha2 "github.com/crossplaneio/stack-azure/apis/v1alpha2"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
)

var (
	_ resource.ExternalClient    = &external{}
	_ resource.ExternalConnecter = &connecter{}
)

type MockMySQLServerAPI struct {
	MockServerNameTaken func(ctx context.Context, s *v1alpha2.MysqlServer) (bool, error)
	MockGetServer       func(ctx context.Context, s *v1alpha2.MysqlServer) (mysql.Server, error)
	MockCreateServer    func(ctx context.Context, s *v1alpha2.MysqlServer, adminPassword string) error
	MockDeleteServer    func(ctx context.Context, s *v1alpha2.MysqlServer) error
}

func (m *MockMySQLServerAPI) ServerNameTaken(ctx context.Context, s *v1alpha2.MysqlServer) (bool, error) {
	return m.MockServerNameTaken(ctx, s)
}

func (m *MockMySQLServerAPI) GetServer(ctx context.Context, s *v1alpha2.MysqlServer) (mysql.Server, error) {
	return m.MockGetServer(ctx, s)
}

func (m *MockMySQLServerAPI) CreateServer(ctx context.Context, s *v1alpha2.MysqlServer, adminPassword string) error {
	return m.MockCreateServer(ctx, s, adminPassword)
}

func (m *MockMySQLServerAPI) DeleteServer(ctx context.Context, s *v1alpha2.MysqlServer) error {
	return m.MockDeleteServer(ctx, s)
}

type modifier func(*v1alpha2.MysqlServer)

func withName(name string) modifier {
	return func(p *v1alpha2.MysqlServer) {
		p.SetName(name)
	}
}

func withProviderRef(r *corev1.ObjectReference) modifier {
	return func(p *v1alpha2.MysqlServer) {
		p.Spec.ProviderReference = r
	}
}

func withAdminName(name string) modifier {
	return func(p *v1alpha2.MysqlServer) {
		p.Spec.AdminLoginName = name
	}
}

func mysqlserver(m ...modifier) *v1alpha2.MysqlServer {
	p := &v1alpha2.MysqlServer{}

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
		"ErrNotAMysqlServer": {
			ec: &connecter{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotMysqlServer),
		},
		"ErrGetProvider": {
			ec: &connecter{
				client: &test.MockClient{MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
					switch obj.(type) {
					case *azurev1alpha2.Provider:
						return errBoom
					default:
						return errors.New("unexpected type")
					}
				})},
				newClientFn: func(credentials []byte) (azure.MySQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(withProviderRef(&corev1.ObjectReference{})),
			},
			want: errors.Wrap(errBoom, errGetProvider),
		},
		"ErrGetProviderSecret": {
			ec: &connecter{
				client: &test.MockClient{MockGet: test.NewMockGetFn(nil, func(obj runtime.Object) error {
					switch obj.(type) {
					case *azurev1alpha2.Provider:
						return nil
					case *corev1.Secret:
						return errBoom
					default:
						return errors.New("unexpected type")
					}
				})},
				newClientFn: func(credentials []byte) (azure.MySQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(withProviderRef(&corev1.ObjectReference{})),
			},
			want: errors.Wrap(errBoom, errGetProviderSecret),
		},
		"Successful": {
			ec: &connecter{
				client:      &test.MockClient{MockGet: test.NewMockGetFn(nil)},
				newClientFn: func(credentials []byte) (azure.MySQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(withProviderRef(&corev1.ObjectReference{})),
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
		"ErrNotAMysqlServer": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotMysqlServer),
			},
		},
		"ErrGetServer": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1alpha2.MysqlServer) (mysql.Server, error) {
						return mysql.Server{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetMysqlServer),
			},
		},
		"ErrCheckServerName": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1alpha2.MysqlServer) (mysql.Server, error) {
						return mysql.Server{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockServerNameTaken: func(_ context.Context, _ *v1alpha2.MysqlServer) (bool, error) {
						return false, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCheckMysqlServerName),
			},
		},
		"ServerCreating": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1alpha2.MysqlServer) (mysql.Server, error) {
						return mysql.Server{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockServerNameTaken: func(_ context.Context, _ *v1alpha2.MysqlServer) (bool, error) {
						return true, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: want{
				eo: resource.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"ServerNotFound": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1alpha2.MysqlServer) (mysql.Server, error) {
						return mysql.Server{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockServerNameTaken: func(_ context.Context, _ *v1alpha2.MysqlServer) (bool, error) {
						return false, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: want{
				eo: resource.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"ServerAvailable": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1alpha2.MysqlServer) (mysql.Server, error) {
						return mysql.Server{ServerProperties: &mysql.ServerProperties{
							UserVisibleState:         mysql.ServerStateReady,
							FullyQualifiedDomainName: &endpoint,
						}}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg: mysqlserver(
					withName(name),
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
		"ErrNotAMysqlServer": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotMysqlServer),
			},
		},
		"ErrGeneratePassword": {
			e: &external{
				newPasswordFn: func(int) (string, error) { return "", errBoom },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGenPassword),
			},
		},
		"ErrCreateServer": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockCreateServer: func(_ context.Context, _ *v1alpha2.MysqlServer, _ string) error { return errBoom },
				},
				newPasswordFn: func(int) (string, error) { return password, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateMysqlServer),
			},
		},
		"Successful": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockCreateServer: func(_ context.Context, _ *v1alpha2.MysqlServer, _ string) error { return nil },
				},
				newPasswordFn: func(int) (string, error) { return password, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
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
		"ErrNotAMysqlServer": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotMysqlServer),
		},
		"ErrDeleteServer": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockDeleteServer: func(_ context.Context, _ *v1alpha2.MysqlServer) error { return errBoom },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: errors.Wrap(errBoom, errDeleteMysqlServer),
		},
		"Successful": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockDeleteServer: func(_ context.Context, _ *v1alpha2.MysqlServer) error { return nil },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
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
