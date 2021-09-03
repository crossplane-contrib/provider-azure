package secret

import (
	"reflect"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/provider-azure/apis/keyvault/v1alpha1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

const (
	errCheckUpToDate = "unable to determine if external resource is up to date"
)

// GenerateObservation produces a KeyVaultSecretObservation object from the keyvault.SecretBundle
// received from Azure.
func GenerateObservation(az keyvault.SecretBundle) v1alpha1.KeyVaultSecretObservation {
	o := v1alpha1.KeyVaultSecretObservation{
		ID:      azure.ToString(az.ID),
		Kid:     az.Kid,
		Managed: az.Managed,
	}
	o.Attributes = generateKeyVaultSecretAttributes(az.Attributes)
	return o
}

func unixTimeComparer() cmp.Option {
	return cmp.Comparer(func(x, y *date.UnixTime) bool {
		if x == nil {
			return y == nil
		}
		if y == nil {
			return false
		}
		return time.Time(*x).Equal(time.Time(*y))
	})
}

// IsUpToDate checks whether SecretBundle is configured with given KeyVaultSecretParameters.
func IsUpToDate(spec v1alpha1.KeyVaultSecretParameters, observed *keyvault.SecretBundle) (bool, error) {
	// Add unixTimeCopier to copystructure to copy date.UnixTime correctly
	copystructure.Copiers[reflect.TypeOf(date.UnixTime{})] = unixTimeCopier

	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	clone, ok := generated.(*keyvault.SecretBundle)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}

	desired := overrideParameters(spec, *clone)

	return cmp.Equal(
		desired,
		*observed,
		cmpopts.IgnoreFields(keyvault.SecretBundle{}, "Response"),
		unixTimeComparer(),
	), nil
}

// GenerateAttributes creates *keyvault.KeyVaultSecretAttributesParameters from *v1alpha1.KeyVaultSecretAttributes.
func GenerateAttributes(spec *v1alpha1.KeyVaultSecretAttributesParameters) *keyvault.SecretAttributes {
	if spec == nil {
		return nil
	}

	return &keyvault.SecretAttributes{
		Enabled:   spec.Enabled,
		NotBefore: metav1TimeToUnixTime(spec.NotBefore),
		Expires:   metav1TimeToUnixTime(spec.Expires),
	}
}

// LateInitialize fills the spec values that user did not fill with their
// corresponding value in the Azure, if there is any.
func LateInitialize(spec *v1alpha1.KeyVaultSecretParameters, az keyvault.SecretBundle) {
	spec.Tags = azure.LateInitializeStringMap(spec.Tags, az.Tags)
	spec.ContentType = azure.LateInitializeStringPtrFromPtr(spec.ContentType, az.ContentType)
	spec.SecretAttributes = lateInitializeSecretAttributes(spec.SecretAttributes, az.Attributes)
}

func lateInitializeSecretAttributes(sa *v1alpha1.KeyVaultSecretAttributesParameters, az *keyvault.SecretAttributes) *v1alpha1.KeyVaultSecretAttributesParameters {
	if az == nil {
		return sa
	}
	if sa == nil {
		sa = &v1alpha1.KeyVaultSecretAttributesParameters{}
	}
	sa.Expires = lateInitializeTimePtrFromUnixTimePtr(sa.Expires, az.Expires)
	sa.NotBefore = lateInitializeTimePtrFromUnixTimePtr(sa.NotBefore, az.NotBefore)
	sa.Enabled = azure.LateInitializeBoolPtrFromPtr(sa.Enabled, az.Enabled)
	return sa
}

func lateInitializeTimePtrFromUnixTimePtr(mt *metav1.Time, ut *date.UnixTime) *metav1.Time {
	if mt != nil {
		return mt
	}
	return unixTimeToMetav1Time(ut)
}

func unixTimeToMetav1Time(t *date.UnixTime) *metav1.Time {
	if t == nil {
		return nil
	}
	return &metav1.Time{Time: time.Time(*t)}
}

func metav1TimeToUnixTime(t *metav1.Time) *date.UnixTime {
	if t == nil {
		return nil
	}

	return (*date.UnixTime)(&t.Time)
}

func overrideParameters(sp v1alpha1.KeyVaultSecretParameters, sb keyvault.SecretBundle) keyvault.SecretBundle {
	sb.Value = azure.ToStringPtr(sp.Value)

	if sp.ContentType != nil {
		sb.ContentType = sp.ContentType
	}

	if sp.Tags != nil {
		if sb.Tags == nil {
			sb.Tags = azure.ToStringPtrMap(sp.Tags)
		} else {
			for k, v := range sp.Tags {
				sb.Tags[k] = azure.ToStringPtr(v)
			}
		}
	}

	if sp.SecretAttributes != nil {
		attr := GenerateAttributes(sp.SecretAttributes)
		if sb.Attributes == nil {
			sb.Attributes = attr
		} else {
			sb.Attributes.Enabled = attr.Enabled
			sb.Attributes.NotBefore = attr.NotBefore
			sb.Attributes.Expires = attr.Expires
		}
	}

	return sb
}

func unixTimeCopier(v interface{}) (interface{}, error) {
	// Just... copy it.
	return v.(date.UnixTime), nil
}

func generateKeyVaultSecretAttributes(az *keyvault.SecretAttributes) *v1alpha1.KeyVaultSecretAttributesObservation {
	if az == nil {
		return nil
	}

	return &v1alpha1.KeyVaultSecretAttributesObservation{
		RecoveryLevel: v1alpha1.DeletionRecoveryLevel(az.RecoveryLevel),
		Created:       unixTimeToMetav1Time(az.Created),
		Updated:       unixTimeToMetav1Time(az.Updated),
	}
}
