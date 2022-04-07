/*
Copyright 2022 The Crossplane Authors.

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

// Code generated by terrajet. DO NOT EDIT.

// Package apis contains Kubernetes API for the provider.
package apis

import (
	"k8s.io/apimachinery/pkg/runtime"

	v1alpha2 "github.com/crossplane/provider-azure/apis/authorization/v1alpha2"
	v1alpha2azure "github.com/crossplane/provider-azure/apis/azure/v1alpha2"
	v1alpha2cache "github.com/crossplane/provider-azure/apis/cache/v1alpha2"
	v1alpha2containerservice "github.com/crossplane/provider-azure/apis/containerservice/v1alpha2"
	v1alpha2cosmosdb "github.com/crossplane/provider-azure/apis/cosmosdb/v1alpha2"
	v1alpha2dbforpostgresql "github.com/crossplane/provider-azure/apis/dbforpostgresql/v1alpha2"
	v1alpha1 "github.com/crossplane/provider-azure/apis/devices/v1alpha1"
	v1alpha2devices "github.com/crossplane/provider-azure/apis/devices/v1alpha2"
	v1alpha2eventhub "github.com/crossplane/provider-azure/apis/eventhub/v1alpha2"
	v1alpha2insights "github.com/crossplane/provider-azure/apis/insights/v1alpha2"
	v1alpha2keyvault "github.com/crossplane/provider-azure/apis/keyvault/v1alpha2"
	v1alpha2loganalytics "github.com/crossplane/provider-azure/apis/loganalytics/v1alpha2"
	v1alpha2network "github.com/crossplane/provider-azure/apis/network/v1alpha2"
	v1alpha2resources "github.com/crossplane/provider-azure/apis/resources/v1alpha2"
	v1alpha2sql "github.com/crossplane/provider-azure/apis/sql/v1alpha2"
	v1alpha2storage "github.com/crossplane/provider-azure/apis/storage/v1alpha2"
	v1alpha1apis "github.com/crossplane/provider-azure/apis/v1alpha1"
)

func init() {
	// Register the types with the Scheme so the components can map objects to GroupVersionKinds and back
	AddToSchemes = append(AddToSchemes,
		v1alpha2.SchemeBuilder.AddToScheme,
		v1alpha2azure.SchemeBuilder.AddToScheme,
		v1alpha2cache.SchemeBuilder.AddToScheme,
		v1alpha2containerservice.SchemeBuilder.AddToScheme,
		v1alpha2cosmosdb.SchemeBuilder.AddToScheme,
		v1alpha2dbforpostgresql.SchemeBuilder.AddToScheme,
		v1alpha1.SchemeBuilder.AddToScheme,
		v1alpha2devices.SchemeBuilder.AddToScheme,
		v1alpha2eventhub.SchemeBuilder.AddToScheme,
		v1alpha2insights.SchemeBuilder.AddToScheme,
		v1alpha2keyvault.SchemeBuilder.AddToScheme,
		v1alpha2loganalytics.SchemeBuilder.AddToScheme,
		v1alpha2network.SchemeBuilder.AddToScheme,
		v1alpha2resources.SchemeBuilder.AddToScheme,
		v1alpha2sql.SchemeBuilder.AddToScheme,
		v1alpha2storage.SchemeBuilder.AddToScheme,
		v1alpha1apis.SchemeBuilder.AddToScheme,
	)
}

// AddToSchemes may be used to add all resources defined in the project to a Scheme
var AddToSchemes runtime.SchemeBuilder

// AddToScheme adds all Resources to the Scheme
func AddToScheme(s *runtime.Scheme) error {
	return AddToSchemes.AddToScheme(s)
}
