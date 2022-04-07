package clients

import (
	"context"
	"encoding/json"
	"fmt"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/terrajet/pkg/terraform"

	"github.com/crossplane/crossplane-runtime/pkg/resource"
	xpresource "github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-jet-azure/apis/v1alpha1"
)

const (
	// error messages
	errNoProviderConfig     = "no providerConfigRef provided"
	errGetProviderConfig    = "cannot get referenced ProviderConfig"
	errTrackUsage           = "cannot track ProviderConfig usage"
	errExtractCredentials   = "cannot extract credentials"
	errUnmarshalCredentials = "cannot unmarshal Azure credentials as JSON"
	errSubscriptionIDNotSet = "subscription ID must be set in ProviderConfig when credential source is InjectedIdentity"
	errTenantIDNotSet       = "tenant ID must be set in ProviderConfig when credential source is InjectedIdentity"
	// Azure service principal credentials file JSON keys
	keyAzureSubscriptionID = "subscriptionId"
	keyAzureClientID       = "clientId"
	keyAzureClientSecret   = "clientSecret"
	keyAzureTenantID       = "tenantId"
	// Terraform Provider configuration block keys
	keyTerraformFeatures        = "features"
	keySkipProviderRegistration = "skip_provider_registration"
	keyUseMSI                   = "use_msi"
	keyClientID                 = "client_id"
	keySubscriptionID           = "subscription_id"
	keyTenantID                 = "tenant_id"
	keyMSIEndpoint              = "msi_endpoint"
	// environment variable names for storing credentials
	envClientID       = "ARM_CLIENT_ID"
	envClientSecret   = "ARM_CLIENT_SECRET"
	envSubscriptionID = "ARM_SUBSCRIPTION_ID"
	envTenantID       = "ARM_TENANT_ID"

	fmtEnvVar = "%s=%s"
)

// TerraformSetupBuilder returns Terraform setup with provider specific
// configuration like provider credentials used to connect to cloud APIs in the
// expected form of a Terraform provider.
func TerraformSetupBuilder(version, providerSource, providerVersion string) terraform.SetupFn { //nolint:gocyclo
	return func(ctx context.Context, client client.Client, mg resource.Managed) (terraform.Setup, error) {
		ps := terraform.Setup{
			Version: version,
			Requirement: terraform.ProviderRequirement{
				Source:  providerSource,
				Version: providerVersion,
			},
		}

		configRef := mg.GetProviderConfigReference()
		if configRef == nil {
			return ps, errors.New(errNoProviderConfig)
		}
		pc := &v1alpha1.ProviderConfig{}
		if err := client.Get(ctx, types.NamespacedName{Name: configRef.Name}, pc); err != nil {
			return ps, errors.Wrap(err, errGetProviderConfig)
		}

		t := xpresource.NewProviderConfigUsageTracker(client, &v1alpha1.ProviderConfigUsage{})
		if err := t.Track(ctx, mg); err != nil {
			return ps, errors.Wrap(err, errTrackUsage)
		}

		ps.Configuration = map[string]interface{}{
			keyTerraformFeatures: struct{}{},
			// Terraform AzureRM provider tries to register all resource providers
			// in Azure just in case if the provider of the resource you're
			// trying to create is not registered and the returned error is
			// ambiguous. However, this requires service principal to have provider
			// registration permissions which are irrelevant in most contexts.
			// For details, see https://github.com/crossplane-contrib/provider-jet-azure/issues/104
			keySkipProviderRegistration: true,
		}

		var err error
		switch pc.Spec.Credentials.Source { //nolint:exhaustive
		case xpv1.CredentialsSourceInjectedIdentity:
			err = msiAuth(pc, &ps)
		default:
			err = spAuth(ctx, pc, &ps, client)
		}
		return ps, err
	}
}

func spAuth(ctx context.Context, pc *v1alpha1.ProviderConfig, ps *terraform.Setup, client client.Client) error {
	data, err := xpresource.CommonCredentialExtractor(ctx, pc.Spec.Credentials.Source, client, pc.Spec.Credentials.CommonCredentialSelectors)
	if err != nil {
		return errors.Wrap(err, errExtractCredentials)
	}
	azureCreds := map[string]string{}
	if err := json.Unmarshal(data, &azureCreds); err != nil {
		return errors.Wrap(err, errUnmarshalCredentials)
	}
	ps.Configuration[keySubscriptionID] = azureCreds[keyAzureSubscriptionID]
	// set credentials environment
	ps.Env = []string{
		fmt.Sprintf(fmtEnvVar, envSubscriptionID, azureCreds[keyAzureSubscriptionID]),
		fmt.Sprintf(fmtEnvVar, envTenantID, azureCreds[keyAzureTenantID]),
		fmt.Sprintf(fmtEnvVar, envClientID, azureCreds[keyAzureClientID]),
		fmt.Sprintf(fmtEnvVar, envClientSecret, azureCreds[keyAzureClientSecret]),
	}
	return nil
}

func msiAuth(pc *v1alpha1.ProviderConfig, ps *terraform.Setup) error {
	if pc.Spec.SubscriptionID == nil || len(*pc.Spec.SubscriptionID) == 0 {
		return errors.New(errSubscriptionIDNotSet)
	}
	if pc.Spec.TenantID == nil || len(*pc.Spec.TenantID) == 0 {
		return errors.New(errTenantIDNotSet)
	}
	ps.Configuration[keySubscriptionID] = *pc.Spec.SubscriptionID
	ps.Configuration[keyTenantID] = *pc.Spec.TenantID
	ps.Configuration[keyUseMSI] = "true"
	if pc.Spec.MSIEndpoint != nil && len(*pc.Spec.MSIEndpoint) != 0 {
		ps.Configuration[keyMSIEndpoint] = *pc.Spec.MSIEndpoint
	}
	if pc.Spec.ClientID != nil && len(*pc.Spec.ClientID) != 0 {
		ps.Configuration[keyClientID] = *pc.Spec.ClientID
	}
	return nil
}
