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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/provider-azure/apis/database/v1beta1"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	"github.com/crossplane/provider-azure/pkg/clients/database"
)

var (
	_ managed.ExternalClient    = &external{}
	_ managed.ExternalConnecter = &connecter{}
)

type MockMySQLServerAPI struct {
	MockServerNameTaken func(ctx context.Context, s *v1beta1.MySQLServer) (bool, error)
	MockGetServer       func(ctx context.Context, s *v1beta1.MySQLServer) (mysql.Server, error)
	MockCreateServer    func(ctx context.Context, s *v1beta1.MySQLServer, adminPassword string) error
	MockUpdateServer    func(ctx context.Context, s *v1beta1.MySQLServer) error
	MockDeleteServer    func(ctx context.Context, s *v1beta1.MySQLServer) error
	MockGetRESTClient   func() autorest.Sender
}

func (m *MockMySQLServerAPI) GetRESTClient() autorest.Sender {
	return m.MockGetRESTClient()
}

func (m *MockMySQLServerAPI) ServerNameTaken(ctx context.Context, s *v1beta1.MySQLServer) (bool, error) {
	return m.MockServerNameTaken(ctx, s)
}

func (m *MockMySQLServerAPI) GetServer(ctx context.Context, s *v1beta1.MySQLServer) (mysql.Server, error) {
	return m.MockGetServer(ctx, s)
}

func (m *MockMySQLServerAPI) CreateServer(ctx context.Context, s *v1beta1.MySQLServer, adminPassword string) error {
	return m.MockCreateServer(ctx, s, adminPassword)
}

func (m *MockMySQLServerAPI) UpdateServer(ctx context.Context, s *v1beta1.MySQLServer) error {
	return m.MockUpdateServer(ctx, s)
}

func (m *MockMySQLServerAPI) DeleteServer(ctx context.Context, s *v1beta1.MySQLServer) error {
	return m.MockDeleteServer(ctx, s)
}

type modifier func(*v1beta1.MySQLServer)

func withExternalName(name string) modifier {
	return func(p *v1beta1.MySQLServer) {
		meta.SetExternalName(p, name)
	}
}

func withProviderRef(r *corev1.ObjectReference) modifier {
	return func(p *v1beta1.MySQLServer) {
		p.Spec.ProviderReference = r
	}
}

func withAdminName(name string) modifier {
	return func(p *v1beta1.MySQLServer) {
		p.Spec.ForProvider.AdministratorLogin = name
	}
}

func mysqlserver(m ...modifier) *v1beta1.MySQLServer {
	p := &v1beta1.MySQLServer{}

	for _, mod := range m {
		mod(p)
	}
	return p
}

var (
	providerName       = "cool-azure"
	providerSecretName = "cool-azure-secret"
	providerSecretKey  = "credentials"
	providerSecretData = "definitelyjson"
	namespace          = "cool-namespace"

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

func TestConnect(t *testing.T) {
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	cases := map[string]struct {
		ec   managed.ExternalConnecter
		args args
		want error
	}{
		"ErrNotAMySQLServer": {
			ec: &connecter{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotMySQLServer),
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
				newClientFn: func(credentials []byte) (database.MySQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(withProviderRef(&corev1.ObjectReference{})),
			},
			want: errors.Wrap(errBoom, errGetProvider),
		},
		"ErrGetProviderSecret": {
			ec: &connecter{
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
				newClientFn: func(credentials []byte) (database.MySQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(withProviderRef(&corev1.ObjectReference{Name: providerName})),
			},
			want: errors.Wrap(errBoom, errGetProviderSecret),
		},
		"ErrProviderSecretNil": {
			ec: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							nilSecretProvider := provider
							nilSecretProvider.SetCredentialsSecretReference(nil)
							*obj.(*azurev1alpha3.Provider) = nilSecretProvider
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							*obj.(*corev1.Secret) = providerSecret
						}
						return nil
					},
				},
				newClientFn: func(credentials []byte) (database.MySQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(withProviderRef(&corev1.ObjectReference{})),
			},
			want: errors.New(errProviderSecretNil),
		},
		"Successful": {
			ec: &connecter{
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
				newClientFn: func(credentials []byte) (database.MySQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(withProviderRef(&corev1.ObjectReference{Name: providerName})),
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
		eo  managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"ErrNotAMySQLServer": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotMySQLServer),
			},
		},
		"ErrGetServer": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.MySQLServer) (mysql.Server, error) {
						return mysql.Server{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetMySQLServer),
			},
		},
		"ErrCheckServerName": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.MySQLServer) (mysql.Server, error) {
						return mysql.Server{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockServerNameTaken: func(_ context.Context, _ *v1beta1.MySQLServer) (bool, error) {
						return false, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCheckMySQLServerName),
			},
		},
		"ServerCreating": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.MySQLServer) (mysql.Server, error) {
						return mysql.Server{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockServerNameTaken: func(_ context.Context, _ *v1beta1.MySQLServer) (bool, error) {
						return true, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists: true,
				},
			},
		},
		"ServerNotFound": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.MySQLServer) (mysql.Server, error) {
						return mysql.Server{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
					MockServerNameTaken: func(_ context.Context, _ *v1beta1.MySQLServer) (bool, error) {
						return false, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
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
				client: &MockMySQLServerAPI{
					MockGetServer: func(_ context.Context, _ *v1beta1.MySQLServer) (mysql.Server, error) {
						return mysql.Server{
							Sku: &mysql.Sku{},
							ServerProperties: &mysql.ServerProperties{
								UserVisibleState:         mysql.ServerStateReady,
								FullyQualifiedDomainName: &endpoint,
								StorageProfile:           &mysql.StorageProfile{},
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
				mg: mysqlserver(
					withExternalName(name),
					withAdminName(admin),
				),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: managed.ConnectionDetails{
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
		ec  managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"ErrNotAMySQLServer": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotMySQLServer),
			},
		},
		"ErrGeneratePassword": {
			e: &external{
				newPasswordFn: func() (string, error) { return "", errBoom },
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
					MockCreateServer: func(_ context.Context, _ *v1beta1.MySQLServer, _ string) error { return errBoom },
				},
				newPasswordFn: func() (string, error) { return password, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateMySQLServer),
			},
		},
		"Successful": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockCreateServer: func(_ context.Context, _ *v1beta1.MySQLServer, _ string) error { return nil },
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
				mg:  mysqlserver(),
			},
			want: want{
				ec: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{runtimev1alpha1.ResourceCredentialsSecretPasswordKey: []byte(password)},
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
		"ErrNotAMySQLServer": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotMySQLServer),
		},
		"ErrDeleteServer": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockDeleteServer: func(_ context.Context, _ *v1beta1.MySQLServer) error { return errBoom },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  mysqlserver(),
			},
			want: errors.Wrap(errBoom, errDeleteMySQLServer),
		},
		"Successful": {
			e: &external{
				client: &MockMySQLServerAPI{
					MockDeleteServer: func(_ context.Context, _ *v1beta1.MySQLServer) error { return nil },
					MockGetRESTClient: func() autorest.Sender {
						return autorest.SenderFunc(func(*http.Request) (*http.Response, error) {
							return nil, nil
						})
					},
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
