/*
Copyright 2020 The Crossplane Authors.

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

package cosmosdb

import (
	"encoding/json"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb"
	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb/documentdbapi"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

// A AccountClient handles CRUD operations for Azure CosmosDB Accounts.
type AccountClient documentdbapi.DatabaseAccountsClientAPI

// NewDatabaseAccountClient create Azure DatabaseAccountsClient using provided
// credentials data
func NewDatabaseAccountClient(credentials []byte) (AccountClient, error) {
	creds := &azure.Credentials{}
	if err := json.Unmarshal(credentials, creds); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal Azure client secret data")
	}

	config := auth.NewClientCredentialsConfig(creds.ClientID, creds.ClientSecret, creds.TenantID)
	config.AADEndpoint = creds.ActiveDirectoryEndpointURL
	config.Resource = creds.ResourceManagerEndpointURL

	authorizer, err := config.Authorizer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get authorizer from config")
	}

	client := documentdb.NewDatabaseAccountsClient(creds.SubscriptionID)
	client.Authorizer = authorizer

	if err := client.AddToUserAgent(azure.UserAgent); err != nil {
		return nil, errors.Wrap(err, "cannot add to Azure client user agent")
	}

	return client, nil
}

// ToDatabaseAccountCreateOrUpdate from CosmosDBAccountSpec
func ToDatabaseAccountCreateOrUpdate(s *v1alpha3.CosmosDBAccountSpec) documentdb.DatabaseAccountCreateUpdateParameters {
	if s == nil {
		return documentdb.DatabaseAccountCreateUpdateParameters{}
	}

	return documentdb.DatabaseAccountCreateUpdateParameters{
		Kind:                                  s.ForProvider.Kind,
		Location:                              azure.ToStringPtr(s.ForProvider.Location),
		Tags:                                  azure.ToStringPtrMap(s.ForProvider.Tags),
		DatabaseAccountCreateUpdateProperties: toDatabaseProperties(&s.ForProvider.Properties),
	}
}

// UpdateCosmosDBAccountObservation produces SQLServerObservation from
// documentdb.CosmosDBAccountStatus.
func UpdateCosmosDBAccountObservation(o *v1alpha3.CosmosDBAccountStatus, in documentdb.DatabaseAccount) {
	o.AtProvider = &v1alpha3.CosmosDBAccountObservation{
		ID:    azure.ToString(in.ID),
		State: azure.ToString(in.DatabaseAccountProperties.ProvisioningState),
	}
}

func toDatabaseProperties(a *v1alpha3.CosmosDBAccountProperties) *documentdb.DatabaseAccountCreateUpdateProperties {
	if a == nil {
		return nil
	}

	return &documentdb.DatabaseAccountCreateUpdateProperties{
		ConsistencyPolicy:            toDatabaseConsistencyPolicy(a.ConsistencyPolicy),
		Locations:                    toDatabaseLocations(a.Locations),
		DatabaseAccountOfferType:     azure.ToStringPtr(a.DatabaseAccountOfferType),
		EnableAutomaticFailover:      a.EnableAutomaticFailover,
		EnableCassandraConnector:     a.EnableCassandraConnector,
		EnableMultipleWriteLocations: a.EnableAutomaticFailover,
	}
}

func fromDatabaseProperties(a *documentdb.DatabaseAccountProperties) v1alpha3.CosmosDBAccountProperties {
	if a == nil {
		return v1alpha3.CosmosDBAccountProperties{}
	}

	// TODO(asouza): figure out how to handle WriteLocations since Create
	// request do not have R/W Locations, only Locations.
	return v1alpha3.CosmosDBAccountProperties{
		ConsistencyPolicy:            fromDatabaseConsistencyPolicy(a.ConsistencyPolicy),
		Locations:                    fromDatabaseLocations(a.ReadLocations),
		DatabaseAccountOfferType:     string(a.DatabaseAccountOfferType),
		EnableAutomaticFailover:      a.EnableAutomaticFailover,
		EnableCassandraConnector:     a.EnableCassandraConnector,
		EnableMultipleWriteLocations: a.EnableMultipleWriteLocations,
	}
}

// CheckEqualDatabaseProperties compares the observed state with the desired
// spec.
func CheckEqualDatabaseProperties(p v1alpha3.CosmosDBAccountProperties, a documentdb.DatabaseAccount) bool {
	o := fromDatabaseProperties(a.DatabaseAccountProperties)

	// asouza: only keep attributes that can be modified in the comparison.
	return (equalConsistencyPolicyIfNotNull(p.ConsistencyPolicy, o.ConsistencyPolicy) &&
		checkEqualLocations(p.Locations, o.Locations) &&
		equalBoolIfNotNull(p.EnableAutomaticFailover, o.EnableAutomaticFailover) &&
		equalBoolIfNotNull(p.EnableMultipleWriteLocations, o.EnableMultipleWriteLocations))
}

func equalConsistencyPolicyIfNotNull(spec, current *v1alpha3.CosmosDBAccountConsistencyPolicy) bool {
	if spec != nil {
		return (spec == current) || (*spec == *current)
	}

	return true
}

func equalBoolIfNotNull(spec, current *bool) bool {
	return azure.ToBool(spec) == azure.ToBool(current)
}

func toDatabaseConsistencyPolicy(a *v1alpha3.CosmosDBAccountConsistencyPolicy) *documentdb.ConsistencyPolicy {
	if a == nil {
		return nil
	}

	return &documentdb.ConsistencyPolicy{
		DefaultConsistencyLevel: documentdb.DefaultConsistencyLevel(a.DefaultConsistencyLevel),
		MaxStalenessPrefix:      a.MaxStalenessPrefix,
		MaxIntervalInSeconds:    a.MaxIntervalInSeconds,
	}
}

func fromDatabaseConsistencyPolicy(a *documentdb.ConsistencyPolicy) *v1alpha3.CosmosDBAccountConsistencyPolicy {
	if a == nil {
		return nil
	}

	return &v1alpha3.CosmosDBAccountConsistencyPolicy{
		DefaultConsistencyLevel: string(a.DefaultConsistencyLevel),
		MaxStalenessPrefix:      a.MaxStalenessPrefix,
		MaxIntervalInSeconds:    a.MaxIntervalInSeconds,
	}
}

func toDatabaseLocations(a []v1alpha3.CosmosDBAccountLocation) *[]documentdb.Location {
	if a == nil {
		return &[]documentdb.Location{}
	}

	s := make([]documentdb.Location, len(a))
	for i := range a {
		s[i] = documentdb.Location{
			LocationName:     &a[i].LocationName,
			FailoverPriority: &a[i].FailoverPriority,
			IsZoneRedundant:  &a[i].IsZoneRedundant,
		}
	}

	return &s
}

func fromDatabaseLocations(a *[]documentdb.Location) []v1alpha3.CosmosDBAccountLocation {
	lenA := 0
	if a != nil {
		lenA = len(*a)
	}
	s := make([]v1alpha3.CosmosDBAccountLocation, lenA)
	if lenA > 0 {
		for i, location := range *a {
			s[i] = v1alpha3.CosmosDBAccountLocation{
				LocationName:     azure.ToString(location.LocationName),
				FailoverPriority: int32(azure.ToInt(location.FailoverPriority)),
				IsZoneRedundant:  azure.ToBool(location.IsZoneRedundant),
			}
		}
	}
	return s
}

func checkEqualLocations(a, b []v1alpha3.CosmosDBAccountLocation) bool {
	// Consider zero length and nil as equal.
	lenA := 0
	if a != nil {
		lenA = len(a)
	}

	lenB := 0
	if b != nil {
		lenB = len(b)
	}

	if (lenA == 0) && (lenB == 0) {
		return true
	}

	return cmp.Equal(a, b, cmpopts.SortSlices(func(i, j v1alpha3.CosmosDBAccountLocation) bool { return i.LocationName < j.LocationName }))
}
