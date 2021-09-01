package secret

import (
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/crossplane/provider-azure/apis/keyvault/v1alpha1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	tags        = map[string]string{"created_by": "crossplane"}
	contentType = azure.ToStringPtr("text/plain")
	enabled     = azure.ToBoolPtr(true)
	expires     = &metav1.Time{Time: time.Now()}
	notBefore   = &metav1.Time{Time: time.Now()}
	created     = &metav1.Time{Time: time.Now()}
	updated     = &metav1.Time{Time: time.Now()}
	value       = "the-secret-value"
	ID          = "ID"
	kid         = azure.ToStringPtr("Kid")
	managed     = azure.ToBoolPtr(true)
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
				Value:       azure.ToStringPtr(value),
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
				ContentType: contentType,
				ID:          ID,
				Kid:         kid,
				Managed:     managed,
				Value:       value,
				Attributes: &v1alpha1.KeyVaultSecretAttributes{
					Created:       created,
					Enabled:       enabled,
					Expires:       expires,
					NotBefore:     notBefore,
					RecoveryLevel: v1alpha1.RecoverablePurgeable,
					Updated:       updated,
				},
			},
		},
		"RequiredConversion": {
			arg: keyvault.SecretBundle{
				ID:    azure.ToStringPtr(ID),
				Value: azure.ToStringPtr(value),
			},
			want: v1alpha1.KeyVaultSecretObservation{
				ID:    ID,
				Value: value,
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
		arg  *v1alpha1.KeyVaultSecretAttributes
		want *keyvault.SecretAttributes
	}{
		"FullConversion": {
			arg: &v1alpha1.KeyVaultSecretAttributes{
				Enabled:   enabled,
				NotBefore: notBefore,
				Expires:   expires,
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
					SecretAttributes: &v1alpha1.KeyVaultSecretAttributes{
						Enabled:   enabled,
						Expires:   expires,
						NotBefore: notBefore,
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
				SecretAttributes: &v1alpha1.KeyVaultSecretAttributes{
					Enabled:   enabled,
					Expires:   expires,
					NotBefore: notBefore,
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
				SecretAttributes: &v1alpha1.KeyVaultSecretAttributes{
					Enabled:   enabled,
					Expires:   expires,
					NotBefore: notBefore,
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
		az   *keyvault.SecretBundle
		spec v1alpha1.KeyVaultSecretParameters
	}
	cases := map[string]struct {
		args
		want bool
		err  error
	}{
		"NotUpToDate": {
			args: args{
				spec: v1alpha1.KeyVaultSecretParameters{
					Value: value,
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
				az: &keyvault.SecretBundle{
					Value: azure.ToStringPtr(value),
					Tags:  azure.ToStringPtrMap(map[string]string{"created_by": "somebody"}),
				},
			},
			want: false,
		},
		"DiffAttributes": {
			args: args{
				spec: v1alpha1.KeyVaultSecretParameters{
					Value: value,
					SecretAttributes: &v1alpha1.KeyVaultSecretAttributes{
						Enabled: enabled,
					},
				},
				az: &keyvault.SecretBundle{
					Value: azure.ToStringPtr(value),
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
				az: &keyvault.SecretBundle{
					Value: azure.ToStringPtr(value),
				},
			},
			want: true,
		},
		"Same expires dates": {
			args: args{
				spec: v1alpha1.KeyVaultSecretParameters{
					Value: value,
					SecretAttributes: &v1alpha1.KeyVaultSecretAttributes{
						Expires: expires,
					},
				},
				az: &keyvault.SecretBundle{
					Value: azure.ToStringPtr(value),
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
			got, _ := IsUpToDate(tc.args.spec, tc.args.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("IsUpToDate(...): -want, +got:\n%s", diff)
			}
		})
	}
}
