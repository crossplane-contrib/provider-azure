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
	"github.com/Azure/go-autorest/autorest/date"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-azure/apis/keyvault/v1alpha1"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
)

var (
	tags        = map[string]string{"created_by": "crossplane"}
	contentType = azure.ToStringPtr("text/plain")
	enabled     = azure.ToBoolPtr(true)
	expires     = &metav1.Time{Time: time.Now()}
	notBefore   = &metav1.Time{Time: time.Now()}
	created     = &metav1.Time{Time: time.Now()}
	updated     = &metav1.Time{Time: time.Now()}
	secretValue = []byte("secret-value")
	value       = xpv1.SecretKeySelector{
		SecretReference: xpv1.SecretReference{
			Name:      "secret-name",
			Namespace: "secret-namespace",
		},
		Key: "secret-key",
	}
	ID      = "ID"
	kid     = azure.ToStringPtr("Kid")
	managed = azure.ToBoolPtr(true)
	errBoom = errors.New("boom")
)

func toTimePtr(t time.Time) *time.Time {
	return &t
}

func TestGenerateObservation(t *testing.T) {
	cases := map[string]struct {
		arg  keyvault.SecretBundle
		want v1alpha1.KeyVaultSecretObservation
	}{
		"FullConversion": {
			arg: keyvault.SecretBundle{
				ContentType: contentType,
				ID:          azure.ToStringPtr(ID),
				Kid:         kid,
				Managed:     managed,
				Value:       azure.ToStringPtr(string(secretValue)),
				Attributes: &keyvault.SecretAttributes{
					Created:       metav1TimeToUnixTime(created),
					Enabled:       enabled,
					Expires:       metav1TimeToUnixTime(expires),
					NotBefore:     metav1TimeToUnixTime(notBefore),
					RecoveryLevel: keyvault.RecoverablePurgeable,
					Updated:       metav1TimeToUnixTime(updated),
				},
			},
			want: v1alpha1.KeyVaultSecretObservation{
				ID:      ID,
				Kid:     kid,
				Managed: managed,
				Attributes: &v1alpha1.KeyVaultSecretAttributesObservation{
					Created:       created,
					RecoveryLevel: string(keyvault.RecoverablePurgeable),
					Updated:       updated,
				},
			},
		},
		"RequiredConversion": {
			arg: keyvault.SecretBundle{
				ID: azure.ToStringPtr(ID),
			},
			want: v1alpha1.KeyVaultSecretObservation{
				ID: ID,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateObservation(tc.arg)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("GenerateObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateAttributes(t *testing.T) {
	cases := map[string]struct {
		arg  *v1alpha1.KeyVaultSecretAttributesParameters
		want *keyvault.SecretAttributes
	}{
		"FullConversion": {
			arg: &v1alpha1.KeyVaultSecretAttributesParameters{
				Enabled:        enabled,
				NotBeforeDate:  notBefore,
				ExpirationDate: expires,
			},
			want: &keyvault.SecretAttributes{
				Enabled:   enabled,
				NotBefore: metav1TimeToUnixTime(notBefore),
				Expires:   metav1TimeToUnixTime(expires),
			},
		},
		"NilConversion": {
			arg:  nil,
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateAttributes(tc.arg)
			if diff := cmp.Diff(tc.want, got, unixTimeComparer()); diff != "" {
				t.Errorf("GenerateAttributes(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	type args struct {
		az   keyvault.SecretBundle
		spec *v1alpha1.KeyVaultSecretParameters
	}
	cases := map[string]struct {
		args
		want *v1alpha1.KeyVaultSecretParameters
	}{
		"Must use template fields in initialization": {
			args: args{
				spec: &v1alpha1.KeyVaultSecretParameters{
					Tags:        tags,
					ContentType: contentType,
					SecretAttributes: &v1alpha1.KeyVaultSecretAttributesParameters{
						Enabled:        enabled,
						ExpirationDate: expires,
						NotBeforeDate:  notBefore,
					},
				},
				az: keyvault.SecretBundle{
					Tags:        azure.ToStringPtrMap(map[string]string{"key": "value", "created_by": "somebody"}),
					ContentType: azure.ToStringPtr("application/json"),
					Attributes: &keyvault.SecretAttributes{
						Enabled:   azure.ToBoolPtr(!(*enabled)),
						Expires:   (*date.UnixTime)(toTimePtr(time.Now())),
						NotBefore: (*date.UnixTime)(toTimePtr(time.Now())),
					},
				},
			},
			want: &v1alpha1.KeyVaultSecretParameters{
				Tags:        tags,
				ContentType: contentType,
				SecretAttributes: &v1alpha1.KeyVaultSecretAttributesParameters{
					Enabled:        enabled,
					ExpirationDate: expires,
					NotBeforeDate:  notBefore,
				},
			},
		},
		"Must initialize template spec field in initialization": {
			args: args{
				spec: &v1alpha1.KeyVaultSecretParameters{},
				az: keyvault.SecretBundle{
					ContentType: contentType,
					Tags:        azure.ToStringPtrMap(tags),
					Attributes: &keyvault.SecretAttributes{
						Enabled:   enabled,
						Expires:   metav1TimeToUnixTime(expires),
						NotBefore: metav1TimeToUnixTime(notBefore),
					},
				},
			},
			want: &v1alpha1.KeyVaultSecretParameters{
				ContentType: contentType,
				Tags:        tags,
				SecretAttributes: &v1alpha1.KeyVaultSecretAttributesParameters{
					Enabled:        enabled,
					ExpirationDate: expires,
					NotBeforeDate:  notBefore,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitialize(tc.args.spec, tc.args.az)
			if diff := cmp.Diff(tc.args.spec, tc.want); diff != "" {
				t.Errorf("LateInitialize(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsUpToDate(t *testing.T) {
	type args struct {
		client client.Client
		az     *keyvault.SecretBundle
		spec   v1alpha1.KeyVaultSecretParameters
	}
	cases := map[string]struct {
		args
		want bool
	}{
		"NotUpToDate": {
			args: args{
				spec: v1alpha1.KeyVaultSecretParameters{
					Value: value,
				},
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				az: &keyvault.SecretBundle{
					Value: azure.ToStringPtr("other value"),
				},
			},
			want: false,
		},
		"DiffTags": {
			args: args{
				spec: v1alpha1.KeyVaultSecretParameters{
					Value: value,
					Tags:  tags,
				},
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				az: &keyvault.SecretBundle{
					Value: azure.ToStringPtr(string(secretValue)),
					Tags:  azure.ToStringPtrMap(map[string]string{"created_by": "somebody"}),
				},
			},
			want: false,
		},
		"DiffAttributes": {
			args: args{
				spec: v1alpha1.KeyVaultSecretParameters{
					Value: value,
					SecretAttributes: &v1alpha1.KeyVaultSecretAttributesParameters{
						Enabled: enabled,
					},
				},
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				az: &keyvault.SecretBundle{
					Value: azure.ToStringPtr(string(secretValue)),
					Attributes: &keyvault.SecretAttributes{
						Enabled: azure.ToBoolPtr(!(*enabled)),
					},
				},
			},
			want: false,
		},
		"UpToDate": {
			args: args{
				spec: v1alpha1.KeyVaultSecretParameters{
					Value: value,
				},
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				az: &keyvault.SecretBundle{
					Value: azure.ToStringPtr(string(secretValue)),
				},
			},
			want: true,
		},
		"SameExpiresDates": {
			args: args{
				spec: v1alpha1.KeyVaultSecretParameters{
					Value: value,
					SecretAttributes: &v1alpha1.KeyVaultSecretAttributesParameters{
						ExpirationDate: expires,
					},
				},
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
				az: &keyvault.SecretBundle{
					Value: azure.ToStringPtr(string(secretValue)),
					Attributes: &keyvault.SecretAttributes{
						Expires: metav1TimeToUnixTime(expires),
					},
				},
			},
			want: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsUpToDate(context.Background(), tc.args.client, tc.args.spec, tc.args.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestExtractSecretValue(t *testing.T) {
	type args struct {
		client client.Client
		value  xpv1.SecretKeySelector
	}

	type want struct {
		value string
		err   error
	}
	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				value: value,
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(nil, func(obj client.Object) error {
						s, _ := obj.(*corev1.Secret)
						s.Data = map[string][]byte{
							"secret-key": secretValue,
						}

						return nil
					}),
				},
			},
			want: want{
				value: string(secretValue),
			},
		},
		"Failure": {
			args: args{
				value: value,
				client: &test.MockClient{
					MockGet: test.NewMockGetFn(errBoom),
				},
			},
			want: want{
				value: "",
				err:   errors.Wrap(errBoom, "cannot get credentials secret"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := ExtractSecretValue(context.Background(), tc.args.client, &tc.args.value)
			if diff := cmp.Diff(tc.want.value, got); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("IsUpToDate(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}
