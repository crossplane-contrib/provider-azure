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
	"reflect"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/Azure/go-autorest/autorest/date"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/mitchellh/copystructure"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-azure/apis/keyvault/v1alpha1"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
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
func IsUpToDate(ctx context.Context, client client.Client, spec v1alpha1.KeyVaultSecretParameters, observed *keyvault.SecretBundle) (bool, error) {
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
	val, err := ExtractSecretValue(ctx, client, &spec.Value)
	if err != nil {
		return true, err
	}

	desired := overrideParameters(spec, *clone, val)

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
		NotBefore: metav1TimeToUnixTime(spec.NotBeforeDate),
		Expires:   metav1TimeToUnixTime(spec.ExpirationDate),
	}
}

// ExtractSecretValue extracts secret value from secretRef.
func ExtractSecretValue(ctx context.Context, client client.Client, value *xpv1.SecretKeySelector) (string, error) {
	ref := xpv1.CommonCredentialSelectors{SecretRef: value}
	val, err := resource.ExtractSecret(ctx, client, ref)

	return string(val), err
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
	sa.ExpirationDate = lateInitializeTimePtrFromUnixTimePtr(sa.ExpirationDate, az.Expires)
	sa.NotBeforeDate = lateInitializeTimePtrFromUnixTimePtr(sa.NotBeforeDate, az.NotBefore)
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

func overrideParameters(sp v1alpha1.KeyVaultSecretParameters, sb keyvault.SecretBundle, val string) keyvault.SecretBundle {
	sb.Value = azure.ToStringPtr(val)

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
		RecoveryLevel: string(az.RecoveryLevel),
		Created:       unixTimeToMetav1Time(az.Created),
		Updated:       unixTimeToMetav1Time(az.Updated),
	}
}
