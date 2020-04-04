/*
Copyright 2019 The Crossplane Authors.

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

package azure

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"

	"github.com/crossplane/provider-azure/apis/v1alpha3"
)

const (
	// UserAgent is the user agent addition that identifies the Crossplane Azure client
	UserAgent = "crossplane-azure-client"
	// AsyncOperationStatusInProgress is the status value for AsyncOperation type
	// that indicates the operation is still ongoing.
	AsyncOperationStatusInProgress = "InProgress"
	asyncOperationPollingMethod    = "AsyncOperation"
)

// A FieldOption determines how common Go types are translated to the types
// required by the Azure Go SDK.
type FieldOption int

// Field options.
const (
	// FieldRequired causes zero values to be converted to a pointer to the zero
	// value, rather than a nil pointer. Azure Go SDK types use pointer fields,
	// with a nil pointer indicating an unset field. Our ToPtr functions return
	// a nil pointer for a zero values, unless FieldRequired is set.
	FieldRequired FieldOption = iota
)

// Client struct that represents the information needed to connect to the Azure services as a client
type Client struct {
	autorest.Authorizer
	Credentials
}

// Credentials represents the contents of a JSON encoded Azure credentials file.
// It is a subset of the internal type used by the Azure auth library.
// https://github.com/Azure/go-autorest/blob/be17756/autorest/azure/auth/auth.go#L226
type Credentials struct {
	ClientID                       string `json:"clientId"`
	ClientSecret                   string `json:"clientSecret"`
	TenantID                       string `json:"tenantId"`
	SubscriptionID                 string `json:"subscriptionId"`
	ActiveDirectoryEndpointURL     string `json:"activeDirectoryEndpointUrl"`
	ResourceManagerEndpointURL     string `json:"resourceManagerEndpointUrl"`
	ActiveDirectoryGraphResourceID string `json:"activeDirectoryGraphResourceId"`
}

// NewClient returns a client that can be used to connect to Azure services
// using the supplied JSON credentials.
func NewClient(credentials []byte) (*Client, error) {
	creds := Credentials{}
	if err := json.Unmarshal(credentials, &creds); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal azure client secret data")
	}

	// create a config object from the loaded credentials data
	config := auth.NewClientCredentialsConfig(creds.ClientID, creds.ClientSecret, creds.TenantID)
	config.AADEndpoint = creds.ActiveDirectoryEndpointURL
	config.Resource = creds.ResourceManagerEndpointURL

	authorizer, err := config.Authorizer()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get authorizer from config")
	}

	return &Client{
		Authorizer: authorizer,
		Credentials: Credentials{
			SubscriptionID:                 creds.SubscriptionID,
			ClientID:                       creds.ClientID,
			ClientSecret:                   creds.ClientSecret,
			TenantID:                       creds.TenantID,
			ActiveDirectoryEndpointURL:     creds.ActiveDirectoryEndpointURL,
			ActiveDirectoryGraphResourceID: creds.ActiveDirectoryGraphResourceID,
		},
	}, nil
}

// ValidateClient verifies if the given client is valid by testing if it can make an Azure service API call
// TODO: is there a better way to validate the Azure client?
func ValidateClient(client *Client) error {
	groupsClient := resources.NewGroupsClient(client.SubscriptionID)
	groupsClient.Authorizer = client.Authorizer
	groupsClient.AddToUserAgent(UserAgent)

	_, err := groupsClient.ListComplete(context.TODO(), "", nil)
	return err
}

// FetchAsyncOperation updates the given operation object with the most up-to-date
// status retrieved from Azure API.
func FetchAsyncOperation(ctx context.Context, client autorest.Sender, as *v1alpha3.AsyncOperation) error {
	if as == nil || as.PollingURL == "" || as.Method == "" {
		return nil
	}
	// NOTE(muvaf):There is NewFutureFromResponse method to construct Future
	// object but that requires http.Request object. Even though we construct a
	// fake http.Request object, the poll operation makes decisions based on the
	// response status code and request headers. JSON marshal needs less
	// information and it's safer to cover all types of pollingTrackedBase objects.
	futureJSON, err := json.Marshal(map[string]string{
		"method":        as.Method,
		"pollingMethod": asyncOperationPollingMethod,
		"pollingURI":    as.PollingURL,
	})
	if err != nil {
		return err
	}
	op := &azure.Future{}
	if err := op.UnmarshalJSON(futureJSON); err != nil {
		return err
	}
	// NOTE(muvaf): FetchAsyncOperation is meant to fetch the operation status, meaning
	// it shouldn't fail if the operation reports error. It should fail if an
	// error appears during the HTTP calls that are made to fetch operation
	// status. But DoneWithContext returns uses the same error variable for both
	// cases, so, we make a compromise and not return the error even if it's
	// related to fetch call.
	_, err = op.DoneWithContext(ctx, client)
	as.Status = op.Status()
	if err != nil {
		as.ErrorMessage = err.Error()
	}
	return nil
}

// IsNotFound returns a value indicating whether the given error represents that the resource was not found.
func IsNotFound(err error) bool {
	detailedError, ok := err.(autorest.DetailedError)
	if !ok {
		return false
	}

	statusCode, ok := detailedError.StatusCode.(int)
	if !ok {
		return false
	}

	return statusCode == http.StatusNotFound
}

// ToStringPtr converts the supplied string for use with the Azure Go SDK.
func ToStringPtr(s string, o ...FieldOption) *string {
	for _, fo := range o {
		if fo == FieldRequired && s == "" {
			return to.StringPtr(s)
		}
	}

	if s == "" {
		return nil
	}

	return to.StringPtr(s)
}

// ToInt32Ptr converts the supplied int for use with the Azure Go SDK.
func ToInt32Ptr(i int, o ...FieldOption) *int32 {
	for _, fo := range o {
		if fo == FieldRequired && i == 0 {
			return to.Int32Ptr(int32(i))
		}
	}

	if i == 0 {
		return nil
	}
	return to.Int32Ptr(int32(i))
}

// ToInt32PtrFromIntPtr converts the supplied int pointer for use with the Azure Go SDK.
func ToInt32PtrFromIntPtr(i *int, o ...FieldOption) *int32 {
	if i == nil {
		return nil
	}
	return to.Int32Ptr(int32(*i))
}

// ToBoolPtr converts the supplied bool for use with the Azure Go SDK.
func ToBoolPtr(b bool, o ...FieldOption) *bool {
	for _, fo := range o {
		if fo == FieldRequired && !b {
			return to.BoolPtr(b)
		}
	}

	if !b {
		return nil
	}
	return to.BoolPtr(b)
}

// ToStringPtrMap converts the supplied map for use with the Azure Go SDK.
func ToStringPtrMap(m map[string]string) map[string]*string {
	if m == nil {
		return nil
	}
	return *(to.StringMapPtr(m))
}

// ToStringMap converts the supplied map from the Azure Go SDK to internal representation.
func ToStringMap(m map[string]*string) map[string]string {
	if m == nil {
		return nil
	}
	return to.StringMap(m)
}

// ToStringArrayPtr converts []string to *[]string which is expected by Azure API.
func ToStringArrayPtr(m []string) *[]string {
	if m == nil {
		return nil
	}
	return &m
}

// ToString converts the supplied pointer to string to a string, returning the
// empty string if the pointer is nil.
func ToString(s *string) string {
	return to.String(s)
}

// ToInt converts the supplied pointer to int32 to an int, returning zero if the
// pointer is nil,
func ToInt(i *int32) int {
	return int(to.Int32(i))
}

// ToInt32 converts the supplied *int to *int32, while returning nil if the
// supplied reference is nil.
func ToInt32(i *int) *int32 {
	if i == nil {
		return nil
	}
	return to.Int32Ptr(int32(*i))
}

// ToBool converts the supplied pointer to bool to a bool, returning the
// false if the pointer is nil.
func ToBool(b *bool) bool {
	return to.Bool(b)
}

// Late initialization is the concept of filling the empty fields in spec
// via the default ones provided by the system. See
// https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#late-initialization

// LateInitializeStringPtrFromPtr late-inits *string
func LateInitializeStringPtrFromPtr(in, from *string) *string {
	if in != nil {
		return in
	}
	return from
}

// LateInitializeStringPtrFromVal late-inits *string using string
func LateInitializeStringPtrFromVal(in *string, from string) *string {
	if in != nil {
		return in
	}
	return &from
}

// LateInitializeStringMap late-inits map[string]string
func LateInitializeStringMap(in map[string]string, from map[string]*string) map[string]string {
	if in != nil {
		return in
	}
	if from == nil {
		return nil
	}
	return to.StringMap(from)
}

// LateInitializeBoolPtrFromPtr late-inits *bool
func LateInitializeBoolPtrFromPtr(in, from *bool) *bool {
	if in != nil {
		return in
	}
	return from
}

// LateInitializeIntPtrFromInt32Ptr late-inits *int
func LateInitializeIntPtrFromInt32Ptr(in *int, from *int32) *int {
	if in != nil {
		return in
	}
	if from != nil {
		return to.IntPtr(int(*from))
	}
	return nil
}

// LateInitializeStringValArrFromArrPtr late-inits []string
func LateInitializeStringValArrFromArrPtr(in []string, from *[]string) []string {
	if in != nil {
		return in
	}
	return to.StringSlice(from)
}
