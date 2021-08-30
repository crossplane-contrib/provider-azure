package secret

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/crossplane/provider-azure/apis/keyvault/v1alpha1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/google/go-cmp/cmp"
	"github.com/mitchellh/copystructure"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	errCheckUpToDate = "unable to determine if external resource is up to date"
)

// GenerateObservation produces a RedisObservation object from the redis.ResourceType
// received from Azure.
func GenerateObservation(az keyvault.SecretBundle) v1alpha1.SecretObservation {
	o := v1alpha1.SecretObservation{
		ID:          azure.ToString(az.ID),
		Value:       azure.ToString(az.Value),
		ContentType: azure.ToString(az.ContentType),
		Kid:         azure.ToString(az.Kid),
		Managed:     azure.ToBool(az.Managed),
	}
	o.Attributes = generateAttributes(az.Attributes)
	return o
}

func generateAttributes(az *keyvault.SecretAttributes) *v1alpha1.SecretAttributes {
	if az == nil {
		return nil
	}

	return &v1alpha1.SecretAttributes{
		RecoveryLevel: v1alpha1.DeletionRecoveryLevel(az.RecoveryLevel),
		Enabled:       az.Enabled,
		Created:       unixTimeToMetav1Time(az.Created),
		Updated:       unixTimeToMetav1Time(az.Updated),
		Expires:       unixTimeToMetav1Time(az.Expires),
		NotBefore:     unixTimeToMetav1Time(az.NotBefore),
	}
}

// LateInitialize fills the spec values that user did not fill with their
// corresponding value in the Azure, if there is any.
func LateInitialize(spec *v1alpha1.SecretParameters, az keyvault.SecretBundle) {
	spec.Tags = azure.LateInitializeStringMap(spec.Tags, az.Tags)
	spec.ContentType = azure.LateInitializeStringPtrFromPtr(spec.ContentType, az.ContentType)
	lateInitializeSecretAttributesFromAttributes(spec.SecretAttributes, az.Attributes)
}

func lateInitializeSecretAttributesFromAttributes(sa *v1alpha1.SecretAttributes, az *keyvault.SecretAttributes) *v1alpha1.SecretAttributes {
	sa.Created = lateInitializeTimePtrFromUnixTimePtr(sa.Created, az.Created)
	sa.Updated = lateInitializeTimePtrFromUnixTimePtr(sa.Updated, az.Updated)
	sa.Expires = lateInitializeTimePtrFromUnixTimePtr(sa.Expires, az.Expires)
	sa.NotBefore = lateInitializeTimePtrFromUnixTimePtr(sa.NotBefore, az.NotBefore)
	sa.Enabled = azure.LateInitializeBoolPtrFromPtr(sa.Enabled, az.Enabled)
	// TODO: update azure keyvault SDK package
	// sa.RecoverableDays = azure.LateInitializeIntPtrFromInt32Ptr(sa.RecoverableDays, az.RecoverableDays)
	sa.RecoveryLevel = v1alpha1.DeletionRecoveryLevel(*azure.LateInitializeStringPtrFromPtr((*string)(&sa.RecoveryLevel), (*string)(&az.RecoveryLevel)))
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

// GenerateSecretBundle takes a SecretParameters and returns *keyvault.SecretBundle.
// It assigns only the fields that are writable, i.e. not labelled as [READ-ONLY]
// in Azure's reference.
func GenerateSecretBundle(spec v1alpha1.SecretParameters, secret *keyvault.SecretBundle) keyvault.SecretBundle {
	secret.Value = azure.ToStringPtr(spec.Value)
	secret.ContentType = spec.ContentType
	secret.Tags = azure.ToStringPtrMap(spec.Tags)
	secret.Attributes = GenerateSecretAttributes(spec.SecretAttributes)
	return *secret
}

// GenerateSecretAttributes takes a *v1alpha1.SecretAttributes and returns *keyvault.SecretAttributes.
func GenerateSecretAttributes(spec *v1alpha1.SecretAttributes) *keyvault.SecretAttributes {
	if spec == nil {
		return nil
	}

	return &keyvault.SecretAttributes{
		Enabled:   spec.Enabled,
		Expires:   metav1TimeToUnixTime(spec.Expires),
		NotBefore: metav1TimeToUnixTime(spec.NotBefore),
	}
}

func IsUpToDate(spec v1alpha1.SecretParameters, observed keyvault.SecretBundle) (bool, error) {
	generated, err := copystructure.Copy(observed)
	if err != nil {
		return true, errors.Wrap(err, errCheckUpToDate)
	}
	clone, ok := generated.(*keyvault.SecretBundle)
	if !ok {
		return true, errors.New(errCheckUpToDate)
	}

	desired := GenerateSecretBundle(spec, clone)

	return cmp.Equal(
		desired,
		observed,
	), nil
}
