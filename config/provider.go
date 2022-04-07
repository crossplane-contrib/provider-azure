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

package config

import (
	// Note(ezgidemirel): we are importing this to embed provider schema document
	_ "embed"

	tjconfig "github.com/crossplane/terrajet/pkg/config"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/crossplane-contrib/provider-jet-azure/config/apimanagement"
	"github.com/crossplane-contrib/provider-jet-azure/config/base"
	"github.com/crossplane-contrib/provider-jet-azure/config/common"
	"github.com/crossplane-contrib/provider-jet-azure/config/cosmosdb"
	"github.com/crossplane-contrib/provider-jet-azure/config/datashare"
	"github.com/crossplane-contrib/provider-jet-azure/config/devices"
	"github.com/crossplane-contrib/provider-jet-azure/config/eventhub"
	"github.com/crossplane-contrib/provider-jet-azure/config/ip"
	"github.com/crossplane-contrib/provider-jet-azure/config/keyvault"
	"github.com/crossplane-contrib/provider-jet-azure/config/kubernetes"
	"github.com/crossplane-contrib/provider-jet-azure/config/loganalytics"
	"github.com/crossplane-contrib/provider-jet-azure/config/logic"
	"github.com/crossplane-contrib/provider-jet-azure/config/management"
	"github.com/crossplane-contrib/provider-jet-azure/config/monitor"
	"github.com/crossplane-contrib/provider-jet-azure/config/network"
	"github.com/crossplane-contrib/provider-jet-azure/config/notificationhubs"
	"github.com/crossplane-contrib/provider-jet-azure/config/postgresql"
	"github.com/crossplane-contrib/provider-jet-azure/config/redis"
	"github.com/crossplane-contrib/provider-jet-azure/config/resource"
	"github.com/crossplane-contrib/provider-jet-azure/config/security"
	"github.com/crossplane-contrib/provider-jet-azure/config/sql"
	"github.com/crossplane-contrib/provider-jet-azure/config/storage"
	"github.com/crossplane-contrib/provider-jet-azure/config/storagesync"
	"github.com/crossplane-contrib/provider-jet-azure/config/subnet"
)

const (
	resourcePrefix = "azure"
	modulePath     = "github.com/crossplane-contrib/provider-jet-azure"
)

//go:embed schema.json
var providerSchema string

var includedResources = []string{
	// "azurerm_.+",
	"azurerm_virtual_.+",
	"azurerm_kubernetes_.+",
	"azurerm_postgresql_.+",
	"azurerm_cosmosdb_.+",
	"azurerm_redis_.+",
	"azurerm_resource_group",
	"azurerm_subnet",
	"azurerm_storage_account$",
	"azurerm_storage_container$",
	"azurerm_storage_blob$",
	"azurerm_sql_server",
	"azurerm_mssql_server$",
	"azurerm_mssql_server_transparent_data_encryption$",
	"azurerm_lb$",
	"azurerm_log_analytics_workspace",
	"azurerm_iothub.*",
	"azurerm_monitor_metric_alert",
	"azurerm_key_vault.*",
	"azurerm_eventhub_namespace$",
	"azurerm_eventhub$",
	"azurerm_eventhub_consumer_group$",
	"azurerm_eventhub_authorization_rule$",
	"azurerm_network_interface$",
}

// These resources cannot be generated because of their suffixes colliding with
// kubebuilder-accepted type suffixes.
var skipList = []string{
	"azurerm_mssql_server_extended_auditing_policy",
	// group prefix collision
	"azurerm_api_management_group",
	"azurerm_api_management_product_group",
	"azurerm_dedicated_host_group",
	"azurerm_storage_sync_group",
	"azurerm_virtual_desktop_application_group",
	// associated with non-generated
	"azurerm_virtual_desktop_workspace_application_group_association",
	// generated name too long
	"azurerm_network_interface_application_gateway_backend_address_pool_association",
	"azurerm_sentinel_data_connector_microsoft_defender_advanced_threat_protection",
	// cannot generate a unique name
	"azurerm_route",
	"azurerm_network_security_rule",
	// deprecated
	"azurerm_virtual_machine_scale_set",
	"azurerm_virtual_machine_configuration_policy_assignment",
	"azurerm_virtual_machine",
	"azurerm_virtual_machine_extension",
	"azurerm_virtual_machine_data_disk_attachment",
	"azurerm_virtual_machine_scale_set_extension",
	// irrelevant
	"azurerm_virtual_desktop_application",
	"azurerm_virtual_desktop_host_pool",
	"azurerm_virtual_desktop_workspace",
	// other terrajet issues
	// ...
	// doc not found in Terraform Azurerm provider
	"azurerm_virtual_network_dns_servers",
	// unsupported sensitive field type
	"azurerm_security_center_automation",
	"azurerm_data_factory_trigger_tumbling_window",
	"azurerm_storage_share_file",
	"azurerm_sql_virtual_network_rule",
	"azurerm_virtual_desktop_workspace",
	"azurerm_data_lake_analytics_account",
	"azurerm_virtual_hub",
	"azurerm_log_analytics_storage_insights",
	"azurerm_virtual_hub_bgp_connection",
	"azurerm_automation_dsc_configuration",
	"azurerm_monitor_log_profile",
	"azurerm_machine_learning_inference_cluster",
	"azurerm_sql_failover_group",
	"azurerm_logic_app_integration_account_certificate",
	"azurerm_container_group",
}

// GetProvider returns provider configuration
func GetProvider() *tjconfig.Provider {
	pc := tjconfig.NewProviderWithSchema([]byte(providerSchema), resourcePrefix, modulePath,
		tjconfig.WithShortName("azurejet"),
		tjconfig.WithRootGroup("azure.jet.crossplane.io"),
		tjconfig.WithIncludeList(includedResources),
		tjconfig.WithSkipList(skipList),
		tjconfig.WithDefaultResourceFn(defaultResource(externalNameConfig(), groupOverrides())),
	)
	// external-name configuration for all resources
	for name := range pc.Resources {
		// external-name configuration
		pc.AddResourceConfigurator(name, func(r *tjconfig.Resource) {
			r.ExternalName = tjconfig.IdentifierFromProvider
		})
	}

	for _, configure := range []func(provider *tjconfig.Provider){
		// add custom config functions
		network.Configure,
		ip.Configure,
		management.Configure,
		redis.Configure,
		resource.Configure,
		kubernetes.Configure,
		postgresql.Configure,
		cosmosdb.Configure,
		sql.Configure,
		subnet.Configure,
		storage.Configure,
		loganalytics.Configure,
		devices.Configure,
		monitor.Configure,
		apimanagement.Configure,
		logic.Configure,
		security.Configure,
		base.Configure,
		datashare.Configure,
		notificationhubs.Configure,
		storagesync.Configure,
		keyvault.Configure,
		eventhub.Configure,
	} {
		configure(pc)
	}

	pc.ConfigureResources()

	// This function runs after the special configurations were applied. However, if some references were configured in
	// the configuration files, the reference generator does not override them.
	for _, r := range pc.Resources {
		if err := common.AddCommonReferences(r); err != nil {
			panic(err)
		}
	}

	return pc
}

func defaultResource(opts ...tjconfig.ResourceOption) tjconfig.DefaultResourceFn {
	return func(name string, terraformResource *schema.Resource, orgOpts ...tjconfig.ResourceOption) *tjconfig.Resource {
		return tjconfig.DefaultResource(name, terraformResource, append(orgOpts, opts...)...)
	}
}
