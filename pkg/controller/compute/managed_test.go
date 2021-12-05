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
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	"github.com/crossplane/provider-azure/pkg/clients/compute/fake"
)

const (
	testPasswd         = "pass123"
	testExistingSecret = "existingSecret"
)

type modifier func(*v1alpha3.AKSCluster)

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

func withConnectionSecretRef(ref *xpv1.SecretReference) modifier {
	return func(c *v1alpha3.AKSCluster) {
		c.Spec.WriteConnectionSecretToReference = ref
	}
}

func aksCluster(m ...modifier) *v1alpha3.AKSCluster {
	ac := &v1alpha3.AKSCluster{}

	for _, mod := range m {
		mod(ac)
	}

	return ac
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
				ec: managed.ExternalCreation{
					ConnectionDetails: map[string][]byte{
						"password": {},
					},
				},
			},
		},
		"SuccessEnsureCluster": {
			e: &external{
				newPasswordFn: func() (string, error) { return testPasswd, nil },
				client: fake.AKSClient{
					MockEnsureManagedCluster: func(_ context.Context, _ *v1alpha3.AKSCluster, _ string) error {
						return nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  aksCluster(),
			},
			want: want{
				ec: managed.ExternalCreation{
					ConnectionDetails: map[string][]byte{
						"password": []byte(testPasswd),
					},
				},
			},
		},
		"SuccessExistingEmptyAppSecret": {
			e: &external{
				newPasswordFn: func() (string, error) { return testPasswd, nil },
				client: fake.AKSClient{
					MockEnsureManagedCluster: func(_ context.Context, _ *v1alpha3.AKSCluster, _ string) error {
						return nil
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, _ client.ObjectKey, o client.Object) error {
						s, ok := o.(*v1.Secret)
						if !ok {
							t.Fatalf("not a *v1.Secret")
						}
						s.Data = map[string][]byte{"password": {}}
						return nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg: aksCluster(withConnectionSecretRef(&xpv1.SecretReference{
					Name:      "test-secret",
					Namespace: "test-ns",
				})),
			},
			want: want{
				ec: managed.ExternalCreation{
					ConnectionDetails: map[string][]byte{
						"password": []byte(testPasswd),
					},
				},
			},
		},
		"SuccessExistingNonEmptyAppSecret": {
			e: &external{
				newPasswordFn: func() (string, error) { return testPasswd, nil },
				client: fake.AKSClient{
					MockEnsureManagedCluster: func(_ context.Context, _ *v1alpha3.AKSCluster, _ string) error {
						return nil
					},
				},
				kube: &test.MockClient{
					MockGet: func(_ context.Context, _ client.ObjectKey, o client.Object) error {
						s, ok := o.(*v1.Secret)
						if !ok {
							t.Fatalf("not a *v1.Secret")
						}
						s.Data = map[string][]byte{"password": []byte(testExistingSecret)}
						return nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg: aksCluster(withConnectionSecretRef(&xpv1.SecretReference{
					Name:      "test-secret",
					Namespace: "test-ns",
				})),
			},
			want: want{
				ec: managed.ExternalCreation{
					ConnectionDetails: map[string][]byte{
						"password": []byte(testExistingSecret),
					},
				},
			},
		},
		"ErrExistingAppSecret": {
			e: &external{
				newPasswordFn: func() (string, error) { return testPasswd, nil },
				client: fake.AKSClient{
					MockEnsureManagedCluster: func(_ context.Context, _ *v1alpha3.AKSCluster, _ string) error {
						return nil
					},
				},
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
			},
			args: args{
				ctx: context.Background(),
				mg: aksCluster(withConnectionSecretRef(&xpv1.SecretReference{
					Name:      "test-secret",
					Namespace: "test-ns",
				})),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetConnSecret),
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
