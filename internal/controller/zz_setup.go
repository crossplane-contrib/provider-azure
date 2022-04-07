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

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/terrajet/pkg/controller"

	resourcegrouppolicyassignment "github.com/crossplane-contrib/provider-jet-azure/internal/controller/authorization/resourcegrouppolicyassignment"
	resourcegroup "github.com/crossplane-contrib/provider-jet-azure/internal/controller/azure/resourcegroup"
	rediscache "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cache/rediscache"
	redisenterprisecluster "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cache/redisenterprisecluster"
	redisenterprisedatabase "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cache/redisenterprisedatabase"
	redisfirewallrule "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cache/redisfirewallrule"
	redislinkedserver "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cache/redislinkedserver"
	kubernetescluster "github.com/crossplane-contrib/provider-jet-azure/internal/controller/containerservice/kubernetescluster"
	kubernetesclusternodepool "github.com/crossplane-contrib/provider-jet-azure/internal/controller/containerservice/kubernetesclusternodepool"
	account "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/account"
	cassandrakeyspace "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/cassandrakeyspace"
	cassandratable "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/cassandratable"
	gremlindatabase "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/gremlindatabase"
	gremlingraph "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/gremlingraph"
	mongocollection "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/mongocollection"
	mongodatabase "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/mongodatabase"
	notebookworkspace "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/notebookworkspace"
	sqlcontainer "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/sqlcontainer"
	sqldatabase "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/sqldatabase"
	sqlfunction "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/sqlfunction"
	sqlstoredprocedure "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/sqlstoredprocedure"
	sqltrigger "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/sqltrigger"
	table "github.com/crossplane-contrib/provider-jet-azure/internal/controller/cosmosdb/table"
	activedirectoryadministrator "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/activedirectoryadministrator"
	configuration "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/configuration"
	database "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/database"
	firewallrule "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/firewallrule"
	flexibleserver "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/flexibleserver"
	flexibleserverconfiguration "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/flexibleserverconfiguration"
	flexibleserverdatabase "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/flexibleserverdatabase"
	flexibleserverfirewallrule "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/flexibleserverfirewallrule"
	server "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/server"
	serverkey "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/serverkey"
	virtualnetworkrule "github.com/crossplane-contrib/provider-jet-azure/internal/controller/dbforpostgresql/virtualnetworkrule"
	iothub "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothub"
	iothubconsumergroup "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubconsumergroup"
	iothubdps "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubdps"
	iothubdpscertificate "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubdpscertificate"
	iothubdpssharedaccesspolicy "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubdpssharedaccesspolicy"
	iothubendpointeventhub "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubendpointeventhub"
	iothubendpointservicebusqueue "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubendpointservicebusqueue"
	iothubendpointservicebustopic "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubendpointservicebustopic"
	iothubendpointstoragecontainer "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubendpointstoragecontainer"
	iothubenrichment "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubenrichment"
	iothubfallbackroute "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubfallbackroute"
	iothubroute "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubroute"
	iothubsharedaccesspolicy "github.com/crossplane-contrib/provider-jet-azure/internal/controller/devices/iothubsharedaccesspolicy"
	authorizationrule "github.com/crossplane-contrib/provider-jet-azure/internal/controller/eventhub/authorizationrule"
	consumergroup "github.com/crossplane-contrib/provider-jet-azure/internal/controller/eventhub/consumergroup"
	eventhub "github.com/crossplane-contrib/provider-jet-azure/internal/controller/eventhub/eventhub"
	eventhubnamespace "github.com/crossplane-contrib/provider-jet-azure/internal/controller/eventhub/eventhubnamespace"
	monitormetricalert "github.com/crossplane-contrib/provider-jet-azure/internal/controller/insights/monitormetricalert"
	accesspolicy "github.com/crossplane-contrib/provider-jet-azure/internal/controller/keyvault/accesspolicy"
	certificate "github.com/crossplane-contrib/provider-jet-azure/internal/controller/keyvault/certificate"
	certificateissuer "github.com/crossplane-contrib/provider-jet-azure/internal/controller/keyvault/certificateissuer"
	key "github.com/crossplane-contrib/provider-jet-azure/internal/controller/keyvault/key"
	managedhardwaresecuritymodule "github.com/crossplane-contrib/provider-jet-azure/internal/controller/keyvault/managedhardwaresecuritymodule"
	managedstorageaccount "github.com/crossplane-contrib/provider-jet-azure/internal/controller/keyvault/managedstorageaccount"
	managedstorageaccountsastokendefinition "github.com/crossplane-contrib/provider-jet-azure/internal/controller/keyvault/managedstorageaccountsastokendefinition"
	secret "github.com/crossplane-contrib/provider-jet-azure/internal/controller/keyvault/secret"
	vault "github.com/crossplane-contrib/provider-jet-azure/internal/controller/keyvault/vault"
	workspace "github.com/crossplane-contrib/provider-jet-azure/internal/controller/loganalytics/workspace"
	loadbalancer "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/loadbalancer"
	networkinterface "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/networkinterface"
	subnet "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/subnet"
	subnetnatgatewayassociation "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/subnetnatgatewayassociation"
	subnetnetworksecuritygroupassociation "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/subnetnetworksecuritygroupassociation"
	subnetroutetableassociation "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/subnetroutetableassociation"
	subnetserviceendpointstoragepolicy "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/subnetserviceendpointstoragepolicy"
	virtualnetwork "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/virtualnetwork"
	virtualnetworkgateway "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/virtualnetworkgateway"
	virtualnetworkgatewayconnection "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/virtualnetworkgatewayconnection"
	virtualnetworkpeering "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/virtualnetworkpeering"
	virtualwan "github.com/crossplane-contrib/provider-jet-azure/internal/controller/network/virtualwan"
	providerconfig "github.com/crossplane-contrib/provider-jet-azure/internal/controller/providerconfig"
	resourcegrouptemplatedeployment "github.com/crossplane-contrib/provider-jet-azure/internal/controller/resources/resourcegrouptemplatedeployment"
	mssqlserver "github.com/crossplane-contrib/provider-jet-azure/internal/controller/sql/mssqlserver"
	mssqlservertransparentdataencryption "github.com/crossplane-contrib/provider-jet-azure/internal/controller/sql/mssqlservertransparentdataencryption"
	serversql "github.com/crossplane-contrib/provider-jet-azure/internal/controller/sql/server"
	accountstorage "github.com/crossplane-contrib/provider-jet-azure/internal/controller/storage/account"
	blob "github.com/crossplane-contrib/provider-jet-azure/internal/controller/storage/blob"
	container "github.com/crossplane-contrib/provider-jet-azure/internal/controller/storage/container"
)

// Setup creates all controllers with the supplied logger and adds them to
// the supplied manager.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	for _, setup := range []func(ctrl.Manager, controller.Options) error{
		resourcegrouppolicyassignment.Setup,
		resourcegroup.Setup,
		rediscache.Setup,
		redisenterprisecluster.Setup,
		redisenterprisedatabase.Setup,
		redisfirewallrule.Setup,
		redislinkedserver.Setup,
		kubernetescluster.Setup,
		kubernetesclusternodepool.Setup,
		account.Setup,
		cassandrakeyspace.Setup,
		cassandratable.Setup,
		gremlindatabase.Setup,
		gremlingraph.Setup,
		mongocollection.Setup,
		mongodatabase.Setup,
		notebookworkspace.Setup,
		sqlcontainer.Setup,
		sqldatabase.Setup,
		sqlfunction.Setup,
		sqlstoredprocedure.Setup,
		sqltrigger.Setup,
		table.Setup,
		activedirectoryadministrator.Setup,
		configuration.Setup,
		database.Setup,
		firewallrule.Setup,
		flexibleserver.Setup,
		flexibleserverconfiguration.Setup,
		flexibleserverdatabase.Setup,
		flexibleserverfirewallrule.Setup,
		server.Setup,
		serverkey.Setup,
		virtualnetworkrule.Setup,
		iothub.Setup,
		iothubconsumergroup.Setup,
		iothubdps.Setup,
		iothubdpscertificate.Setup,
		iothubdpssharedaccesspolicy.Setup,
		iothubendpointeventhub.Setup,
		iothubendpointservicebusqueue.Setup,
		iothubendpointservicebustopic.Setup,
		iothubendpointstoragecontainer.Setup,
		iothubenrichment.Setup,
		iothubfallbackroute.Setup,
		iothubroute.Setup,
		iothubsharedaccesspolicy.Setup,
		authorizationrule.Setup,
		consumergroup.Setup,
		eventhub.Setup,
		eventhubnamespace.Setup,
		monitormetricalert.Setup,
		accesspolicy.Setup,
		certificate.Setup,
		certificateissuer.Setup,
		key.Setup,
		managedhardwaresecuritymodule.Setup,
		managedstorageaccount.Setup,
		managedstorageaccountsastokendefinition.Setup,
		secret.Setup,
		vault.Setup,
		workspace.Setup,
		loadbalancer.Setup,
		networkinterface.Setup,
		subnet.Setup,
		subnetnatgatewayassociation.Setup,
		subnetnetworksecuritygroupassociation.Setup,
		subnetroutetableassociation.Setup,
		subnetserviceendpointstoragepolicy.Setup,
		virtualnetwork.Setup,
		virtualnetworkgateway.Setup,
		virtualnetworkgatewayconnection.Setup,
		virtualnetworkpeering.Setup,
		virtualwan.Setup,
		providerconfig.Setup,
		resourcegrouptemplatedeployment.Setup,
		mssqlserver.Setup,
		mssqlservertransparentdataencryption.Setup,
		serversql.Setup,
		accountstorage.Setup,
		blob.Setup,
		container.Setup,
	} {
		if err := setup(mgr, o); err != nil {
			return err
		}
	}
	return nil
}
