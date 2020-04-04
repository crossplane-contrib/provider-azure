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

package controller

import (
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane/crossplane-runtime/pkg/logging"

	"github.com/crossplane/provider-azure/pkg/controller/cache"
	"github.com/crossplane/provider-azure/pkg/controller/compute"
	"github.com/crossplane/provider-azure/pkg/controller/database/cosmosdb"
	"github.com/crossplane/provider-azure/pkg/controller/database/mysqlserver"
	"github.com/crossplane/provider-azure/pkg/controller/database/mysqlservervirtualnetworkrule"
	"github.com/crossplane/provider-azure/pkg/controller/database/postgresqlserver"
	"github.com/crossplane/provider-azure/pkg/controller/database/postgresqlservervirtualnetworkrule"
	"github.com/crossplane/provider-azure/pkg/controller/network/subnet"
	"github.com/crossplane/provider-azure/pkg/controller/network/virtualnetwork"
	"github.com/crossplane/provider-azure/pkg/controller/resourcegroup"
	"github.com/crossplane/provider-azure/pkg/controller/storage/account"
	"github.com/crossplane/provider-azure/pkg/controller/storage/container"
)

// Setup Azure controllers.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	for _, setup := range []func(ctrl.Manager, logging.Logger) error{
		cache.SetupRedisClaimScheduling,
		cache.SetupRedisClaimDefaulting,
		cache.SetupRedisClaimBinding,
		cache.SetupRedis,
		compute.SetupAKSClusterClaimScheduling,
		compute.SetupAKSClusterClaimDefaulting,
		compute.SetupAKSClusterClaimBinding,
		compute.SetupAKSClusterTarget,
		compute.SetupAKSCluster,
		mysqlserver.SetupClaimScheduling,
		mysqlserver.SetupClaimDefaulting,
		mysqlserver.SetupClaimBinding,
		mysqlserver.Setup,
		mysqlservervirtualnetworkrule.Setup,
		postgresqlserver.SetupClaimScheduling,
		postgresqlserver.SetupClaimDefaulting,
		postgresqlserver.SetupClaimBinding,
		postgresqlserver.Setup,
		postgresqlservervirtualnetworkrule.Setup,
		cosmosdb.Setup,
		virtualnetwork.Setup,
		subnet.Setup,
		resourcegroup.Setup,
		account.Setup,
		container.SetupClaimDefaulting,
		container.SetupClaimScheduling,
		container.SetupClaimBinding,
		container.Setup,
	} {
		if err := setup(mgr, l); err != nil {
			return err
		}
	}
	return nil
}
