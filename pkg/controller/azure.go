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

	computeclients "github.com/crossplaneio/stack-azure/pkg/clients/compute"
	"github.com/crossplaneio/stack-azure/pkg/controller/cache"
	"github.com/crossplaneio/stack-azure/pkg/controller/compute"
	"github.com/crossplaneio/stack-azure/pkg/controller/database/mysqlserver"
	"github.com/crossplaneio/stack-azure/pkg/controller/database/mysqlservervirtualnetworkrule"
	"github.com/crossplaneio/stack-azure/pkg/controller/database/postgresqlserver"
	"github.com/crossplaneio/stack-azure/pkg/controller/database/postgresqlservervirtualnetworkrule"
	"github.com/crossplaneio/stack-azure/pkg/controller/network/subnet"
	"github.com/crossplaneio/stack-azure/pkg/controller/network/virtualnetwork"
	"github.com/crossplaneio/stack-azure/pkg/controller/resourcegroup"
	"github.com/crossplaneio/stack-azure/pkg/controller/storage/account"
	"github.com/crossplaneio/stack-azure/pkg/controller/storage/container"
)

// Controllers passes down config and adds individual controllers to the manager.
type Controllers struct{}

// SetupWithManager adds all Azure controllers to the manager.
func (c *Controllers) SetupWithManager(mgr ctrl.Manager) error {
	aksReconciler := compute.NewAKSClusterReconciler(mgr, &computeclients.AKSSetupClientFactory{})

	controllers := []interface {
		SetupWithManager(ctrl.Manager) error
	}{
		&cache.RedisClaimSchedulingController{},
		&cache.RedisClaimDefaultingController{},
		&cache.RedisClaimController{},
		&cache.RedisController{},
		&compute.AKSClusterClaimSchedulingController{},
		&compute.AKSClusterClaimDefaultingController{},
		&compute.AKSClusterClaimController{},
		&compute.AKSClusterController{Reconciler: aksReconciler},
		&mysqlserver.ClaimSchedulingController{},
		&mysqlserver.ClaimDefaultingController{},
		&mysqlserver.ClaimController{},
		&mysqlserver.Controller{},
		&mysqlservervirtualnetworkrule.Controller{},
		&postgresqlserver.ClaimSchedulingController{},
		&postgresqlserver.ClaimDefaultingController{},
		&postgresqlserver.ClaimController{},
		&postgresqlserver.Controller{},
		&postgresqlservervirtualnetworkrule.Controller{},
		&virtualnetwork.Controller{},
		&subnet.Controller{},
		&resourcegroup.Controller{},
		&account.ClaimSchedulingController{},
		&account.ClaimDefaultingController{},
		&account.ClaimController{},
		&account.Controller{},
		&container.ClaimDefaultingController{},
		&container.ClaimSchedulingController{},
		&container.ClaimController{},
		&container.Controller{},
	}
	for _, c := range controllers {
		if err := c.SetupWithManager(mgr); err != nil {
			return err
		}
	}
	return nil
}
