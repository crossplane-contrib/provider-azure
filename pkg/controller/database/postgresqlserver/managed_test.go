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

	"github.com/Azure/azure-sdk-for-go/profiles/latest/postgresql/mgmt/postgresql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/database/v1beta1"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	"github.com/crossplane/provider-azure/pkg/clients/database"
)

var (
	_ managed.ExternalClient    = &external{}
	_ managed.ExternalConnecter = &connecter{}
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

func withProviderRef(r runtimev1alpha1.Reference) modifier {
	return func(p *v1beta1.PostgreSQLServer) {
		p.Spec.ProviderReference = r
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

var (
	namespace          = "coolNamespace"
	providerName       = "cool-aws"
	providerSecretName = "cool-aws-secret"
	providerSecretKey  = "credentials"
	providerSecretData = "definitelyini"

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
				mg:  postgresqlserver(withProviderRef(runtimev1alpha1.Reference{})),
			},
			want: errors.Wrap(errBoom, errGetProvider),
		},
		"GetProviderSecretFailed": {
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
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(withProviderRef(runtimev1alpha1.Reference{Name: providerName})),
			},
			want: errors.Wrapf(errBoom, errGetProviderSecret),
		},
		"GetProviderSecretNil": {
			ec: &connecter{
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
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(withProviderRef(runtimev1alpha1.Reference{Name: providerName})),
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
				newClientFn: func(credentials []byte) (database.PostgreSQLServerAPI, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  postgresqlserver(withProviderRef(runtimev1alpha1.Reference{Name: providerName})),
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
