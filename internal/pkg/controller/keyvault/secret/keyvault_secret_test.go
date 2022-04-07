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

package secret

import (
	"context"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault/keyvaultapi"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	azure "github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients"
	"github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients/keyvault/secret/fake"

	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/keyvault/v1alpha1"
)

var (
	tags        = map[string]string{"created_by": "crossplane"}
	contentType = azure.ToStringPtr("text/plain")
	enabled     = azure.ToBoolPtr(true)
	expires     = &metav1.Time{Time: time.Now()}
	notBefore   = &metav1.Time{Time: time.Now()}
	ID          = "ID"
	secretValue = []byte("secret-value")
	value       = xpv1.SecretKeySelector{
		SecretReference: xpv1.SecretReference{
			Name:      "secret-name",
			Namespace: "secret-namespace",
		},
		Key: "secret-key",
	}
	vaultBaseURL     = "https://myvault.vault.azure.net"
	name             = "the-secret-name"
	unexpectedObject resource.Managed
)

var (
	errorBoom = errors.New("boom")
)

type keyvaulSecretResourceModifier func(*v1alpha1.KeyVaultSecret)

func withConditions(c ...xpv1.Condition) keyvaulSecretResourceModifier {
	return func(r *v1alpha1.KeyVaultSecret) { r.Status.ConditionedStatus.Conditions = c }
}

func withID(id string) keyvaulSecretResourceModifier {
	return func(r *v1alpha1.KeyVaultSecret) { r.Status.AtProvider.ID = id }
}

func withoutContentType() keyvaulSecretResourceModifier {
	return func(r *v1alpha1.KeyVaultSecret) { r.Spec.ForProvider.ContentType = nil }
}

func instance(rm ...keyvaulSecretResourceModifier) *v1alpha1.KeyVaultSecret {
	cr := &v1alpha1.KeyVaultSecret{
		Spec: v1alpha1.KeyVaultSecretSpec{
			ForProvider: v1alpha1.KeyVaultSecretParameters{
				VaultBaseURL: vaultBaseURL,
				Name:         name,
				Value:        value,
				ContentType:  contentType,
				Tags:         tags,
				SecretAttributes: &v1alpha1.KeyVaultSecretAttributesParameters{
					Enabled:        enabled,
					NotBeforeDate:  notBefore,
					ExpirationDate: expires,
				},
			},
		},
	}

	for _, m := range rm {
		m(cr)
	}

	return cr
}

func TestObserve(t *testing.T) {
	type args struct {
		cr   resource.Managed
		kv   keyvaultapi.BaseClientAPI
		kube client.Client
	}
	type want struct {
		cr  resource.Managed
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ResourceIsNotKeyVaultSecret": {
			args: args{
				cr: unexpectedObject,
			},
			want: want{
				o:   managed.ExternalObservation{},
				err: errors.New(errNotSecret),
			},
		},
		"Successful": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				kv: &fake.MockClient{
					MockGetSecret: func(ctx context.Context, vaultBaseURL string, secretName string, secretVersion string) (keyvault.SecretBundle, error) {
						return keyvault.SecretBundle{
							ID:    azure.ToStringPtr(ID),
							Value: azure.ToStringPtr(string(secretValue)),
						}, nil
					},
				},
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Available()),
					withID(ID),
				),
				o:   managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: false},
				err: nil,
			},
		},
		"GetFailed": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				kv: &fake.MockClient{
					MockGetSecret: func(ctx context.Context, vaultBaseURL string, secretName string, secretVersion string) (keyvault.SecretBundle, error) {
						return keyvault.SecretBundle{}, errorBoom
					},
				},
			},
			want: want{
				cr:  instance(),
				o:   managed.ExternalObservation{ResourceExists: false},
				err: errors.Wrap(errorBoom, errGetFailed),
			},
		},
		"LateInitialized": {
			args: args{
				cr: instance(
					withoutContentType(),
				),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				kv: &fake.MockClient{
					MockGetSecret: func(_ context.Context, _ string, _ string, _ string) (keyvault.SecretBundle, error) {
						return keyvault.SecretBundle{
							ID:          azure.ToStringPtr(ID),
							ContentType: contentType,
						}, nil
					},
				},
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Available()),
					withID(ID),
				),
				o: managed.ExternalObservation{
					ResourceExists:          true,
					ResourceLateInitialized: true,
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{
				kube:   tc.kube,
				client: tc.kv,
			}
			o, err := e.Observe(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Observe(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, o); diff != "" {
				t.Errorf("Observe(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		cr   resource.Managed
		kube client.Client
		kv   keyvaultapi.BaseClientAPI
	}
	type want struct {
		cr  resource.Managed
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ResourceIsNotKeyVaultSecret": {
			args: args{
				cr: unexpectedObject,
			},
			want: want{
				o:   managed.ExternalCreation{},
				err: errors.New(errNotSecret),
			},
		},
		"Successful": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				kv: &fake.MockClient{
					MockSetSecret: func(ctx context.Context, vaultBaseURL, secretName string, parameters keyvault.SecretSetParameters) (keyvault.SecretBundle, error) {
						return keyvault.SecretBundle{}, nil
					},
				},
			},
			want: want{
				o:  managed.ExternalCreation{},
				cr: instance(),
			},
		},
		"Failed": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				kv: &fake.MockClient{
					MockSetSecret: func(ctx context.Context, vaultBaseURL, secretName string, parameters keyvault.SecretSetParameters) (keyvault.SecretBundle, error) {
						return keyvault.SecretBundle{}, errorBoom
					},
				},
			},
			want: want{
				cr:  instance(),
				o:   managed.ExternalCreation{},
				err: errors.Wrap(errorBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.args.kv, kube: tc.args.kube}

			c, err := e.Create(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Create(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, c); diff != "" {
				t.Errorf("Create(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		cr   resource.Managed
		kube client.Client
		kv   keyvaultapi.BaseClientAPI
	}
	type want struct {
		cr  resource.Managed
		o   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ResourceIsNotKeyVaultSecret": {
			args: args{
				cr: unexpectedObject,
			},
			want: want{
				o:   managed.ExternalUpdate{},
				err: errors.New(errNotSecret),
			},
		},
		"Successful": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				kv: &fake.MockClient{
					MockSetSecret: func(ctx context.Context, vaultBaseURL, secretName string, parameters keyvault.SecretSetParameters) (keyvault.SecretBundle, error) {
						return keyvault.SecretBundle{}, nil
					},
				},
			},
			want: want{
				o:  managed.ExternalUpdate{},
				cr: instance(),
			},
		},
		"Failed": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				kv: &fake.MockClient{
					MockSetSecret: func(ctx context.Context, vaultBaseURL, secretName string, parameters keyvault.SecretSetParameters) (keyvault.SecretBundle, error) {
						return keyvault.SecretBundle{}, errorBoom
					},
				},
			},
			want: want{
				cr:  instance(),
				o:   managed.ExternalUpdate{},
				err: errors.Wrap(errorBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.args.kv, kube: tc.args.kube}

			c, err := e.Update(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, c); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		cr resource.Managed
		kv keyvaultapi.BaseClientAPI
	}
	type want struct {
		cr  resource.Managed
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"ResourceIsNotKeyVaultSecret": {
			args: args{
				cr: unexpectedObject,
			},
			want: want{
				err: errors.New(errNotSecret),
			},
		},
		"Successful": {
			args: args{
				cr: instance(),
				kv: &fake.MockClient{
					MockDeleteSecret: func(ctx context.Context, vaultBaseURL, secretName string) (keyvault.DeletedSecretBundle, error) {
						return keyvault.DeletedSecretBundle{}, nil
					},
				},
			},
			want: want{
				cr: instance(),
			},
		},
		"Failed": {
			args: args{
				cr: instance(),
				kv: &fake.MockClient{
					MockDeleteSecret: func(ctx context.Context, vaultBaseURL, secretName string) (keyvault.DeletedSecretBundle, error) {
						return keyvault.DeletedSecretBundle{}, errorBoom
					},
				},
			},
			want: want{
				cr:  instance(),
				err: errors.Wrap(errorBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.args.kv}

			err := e.Delete(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Delete(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Delete(...): -want, +got\n%s", diff)
			}
		})
	}
}
