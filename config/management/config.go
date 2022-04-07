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

package management

import (
	"context"
	"fmt"

	"github.com/crossplane/terrajet/pkg/config"

	"github.com/crossplane-contrib/provider-jet-azure/config/common"
)

// Configure configures management group
func Configure(p *config.Provider) {
	p.AddResourceConfigurator("azurerm_management_group", func(r *config.Resource) {
		r.Version = common.VersionV1Alpha2
		r.Kind = "ManagementGroup"
		r.ExternalName = config.NameAsIdentifier
		r.ExternalName.GetExternalNameFn = common.GetNameFromFullyQualifiedID
		// /providers/Microsoft.Management/managementGroups/group1
		r.ExternalName.GetIDFn = func(_ context.Context, name string, parameters map[string]interface{}, providerConfig map[string]interface{}) (string, error) {
			return fmt.Sprintf("/providers/Microsoft.Management/managementGroups/%s", name), nil
		}
	})
}
