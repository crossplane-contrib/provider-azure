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

package fake

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault/keyvaultapi"
)

var _ keyvaultapi.BaseClientAPI = &MockClient{}

// MockClient is a fake implementation of keyvaultapi.BaseClientAPI.
type MockClient struct {
	keyvaultapi.BaseClientAPI

	MockDeleteSecret func(ctx context.Context, vaultBaseURL string, secretName string) (result keyvault.DeletedSecretBundle, err error)
	MockGetSecret    func(ctx context.Context, vaultBaseURL string, secretName string, secretVersion string) (result keyvault.SecretBundle, err error)
	MockSetSecret    func(ctx context.Context, vaultBaseURL string, secretName string, parameters keyvault.SecretSetParameters) (result keyvault.SecretBundle, err error)
}

// DeleteSecret calls the MockClient's MockDeleteSecret method.
func (c *MockClient) DeleteSecret(ctx context.Context, vaultBaseURL string, secretName string) (result keyvault.DeletedSecretBundle, err error) {
	return c.MockDeleteSecret(ctx, vaultBaseURL, secretName)
}

// GetSecret calls the MockClient's MockGetSecret method.
func (c *MockClient) GetSecret(ctx context.Context, vaultBaseURL string, secretName string, secretVersion string) (result keyvault.SecretBundle, err error) {
	return c.MockGetSecret(ctx, vaultBaseURL, secretName, secretVersion)
}

// SetSecret calls the MockClient's MockSetSecret method.
func (c *MockClient) SetSecret(ctx context.Context, vaultBaseURL string, secretName string, parameters keyvault.SecretSetParameters) (result keyvault.SecretBundle, err error) {
	return c.MockSetSecret(ctx, vaultBaseURL, secretName, parameters)
}
