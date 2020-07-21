/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    htcp://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package compute

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/containerservice/mgmt/2018-03-31/containerservice"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	"github.com/crossplane/provider-azure/pkg/clients/compute"
	"github.com/crossplane/provider-azure/pkg/clients/compute/fake"
)

type modifier func(*v1alpha3.AKSCluster)

func withProviderRef(r runtimev1alpha1.Reference) modifier {
	return func(c *v1alpha3.AKSCluster) {
		c.Spec.ProviderReference = r
	}
}

func withState(state string) modifier {
	return func(c *v1alpha3.AKSCluster) {
		c.Status.State = state
	}
}

func withProviderID(id string) modifier {
	return func(c *v1alpha3.AKSCluster) {
		c.Status.ProviderID = id
	}
}

func withEndpoint(ep string) modifier {
	return func(c *v1alpha3.AKSCluster) {
		c.Status.Endpoint = ep
	}
}

func aksCluster(m ...modifier) *v1alpha3.AKSCluster {
	ac := &v1alpha3.AKSCluster{}

	for _, mod := range m {
		mod(ac)
	}

	return ac
}

func TestConnect(t *testing.T) {
	errBoom := errors.New("boom")
	providerName := "cool-azure"
	providerSecretName := "cool-azure-secret"
	providerSecretKey := "credentials"
	providerSecretData := "definitelyjson"
	namespace := "cool-namespace"

	provider := azurev1alpha3.Provider{
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

	providerSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data:       map[string][]byte{providerSecretKey: []byte(providerSecretData)},
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	cases := map[string]struct {
		ec   managed.ExternalConnecter
		args args
		want error
	}{
		"ErrNotAKSCluster": {
			ec: &connecter{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotAKSCluster),
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
				newClientFn: func(credentials []byte) (compute.AKSClient, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(withProviderRef(runtimev1alpha1.Reference{})),
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
				newClientFn: func(credentials []byte) (compute.AKSClient, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(withProviderRef(runtimev1alpha1.Reference{Name: providerName})),
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
				newClientFn: func(credentials []byte) (compute.AKSClient, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(withProviderRef(runtimev1alpha1.Reference{})),
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
				newClientFn: func(credentials []byte) (compute.AKSClient, error) { return nil, nil },
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(withProviderRef(runtimev1alpha1.Reference{Name: providerName})),
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
	id := "koolAD"
	stateSucceeded := "Succeeded"
	stateWat := "Wat"
	endpoint := "http://wat.example.org"

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		eo  managed.ExternalObservation
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"ErrNotAKSCluster": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotAKSCluster),
			},
		},
		"ErrClusterNotFound": {
			e: &external{
				client: fake.AKSClient{
					MockGetManagedCluster: func(_ context.Context, _ *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error) {
						return containerservice.ManagedCluster{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(),
			},
			want: want{
				eo: managed.ExternalObservation{ResourceExists: false},
				mg: aksCluster(),
			},
		},
		"ErrGetCluster": {
			e: &external{
				client: fake.AKSClient{
					MockGetManagedCluster: func(_ context.Context, _ *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error) {
						return containerservice.ManagedCluster{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetAKSCluster),
				mg:  aksCluster(),
			},
		},
		"NotReady": {
			e: &external{
				client: fake.AKSClient{
					MockGetManagedCluster: func(_ context.Context, _ *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error) {
						return containerservice.ManagedCluster{
							ID: to.StringPtr(id),
							ManagedClusterProperties: &containerservice.ManagedClusterProperties{
								ProvisioningState: to.StringPtr(stateWat),
								Fqdn:              to.StringPtr(endpoint),
							},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(),
			},
			want: want{
				eo: managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true},
				mg: aksCluster(
					withProviderID(id),
					withState(stateWat),
					withEndpoint(endpoint),
				),
			},
		},
		"ErrGetKubeConfig": {
			e: &external{
				client: fake.AKSClient{
					MockGetManagedCluster: func(_ context.Context, _ *v1alpha3.AKSCluster) (containerservice.ManagedCluster, error) {
						return containerservice.ManagedCluster{ManagedClusterProperties: &containerservice.ManagedClusterProperties{
							ProvisioningState: to.StringPtr(stateSucceeded),
						}}, nil
					},
					MockGetKubeConfig: func(_ context.Context, _ *v1alpha3.AKSCluster) ([]byte, error) {
						return nil, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(),
			},
			want: want{
				mg: aksCluster(
					withState(stateSucceeded),
				),
				err: errors.Wrap(errBoom, errGetKubeConfig),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eo, err := tc.e.Observe(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("tc.e.Observe(...): -want managed, +got managed:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.eo, eo); diff != "" {
				t.Errorf("tc.e.Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	errBoom := errors.New("boom")
	// password := "verysecure"

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
		"ErrNotAKSCluster": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotAKSCluster),
			},
		},
		"ErrGeneratePassword": {
			e: &external{
				newPasswordFn: func() (string, error) { return "", errBoom },
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGenPassword),
			},
		},
		"ErrEnsureCluster": {
			e: &external{
				newPasswordFn: func() (string, error) { return "", nil },
				client: fake.AKSClient{
					MockEnsureManagedCluster: func(_ context.Context, _ *v1alpha3.AKSCluster, _ string) error {
						return errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateAKSCluster),
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
		"ErrNotAKSCluster": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotAKSCluster),
		},
		"ErrDeleteCluster": {
			e: &external{
				newPasswordFn: func() (string, error) { return "", nil },
				client: fake.AKSClient{
					MockDeleteManagedCluster: func(_ context.Context, _ *v1alpha3.AKSCluster) error {
						return errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(),
			},
			want: errors.Wrap(errBoom, errDeleteAKSCluster),
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
