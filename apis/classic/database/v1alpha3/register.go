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

package v1alpha3

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "database.azure.crossplane.io"
	Version = "v1alpha3"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// MySQLServerVirtualNetworkRule type metadata.
var (
	MySQLServerVirtualNetworkRuleKind             = reflect.TypeOf(MySQLServerVirtualNetworkRule{}).Name()
	MySQLServerVirtualNetworkRuleGroupKind        = schema.GroupKind{Group: Group, Kind: MySQLServerVirtualNetworkRuleKind}.String()
	MySQLServerVirtualNetworkRuleKindAPIVersion   = MySQLServerVirtualNetworkRuleKind + "." + SchemeGroupVersion.String()
	MySQLServerVirtualNetworkRuleGroupVersionKind = SchemeGroupVersion.WithKind(MySQLServerVirtualNetworkRuleKind)
)

// PostgreSQLServerVirtualNetworkRule type metadata.
var (
	PostgreSQLServerVirtualNetworkRuleKind             = reflect.TypeOf(PostgreSQLServerVirtualNetworkRule{}).Name()
	PostgreSQLServerVirtualNetworkRuleGroupKind        = schema.GroupKind{Group: Group, Kind: PostgreSQLServerVirtualNetworkRuleKind}.String()
	PostgreSQLServerVirtualNetworkRuleKindAPIVersion   = PostgreSQLServerVirtualNetworkRuleKind + "." + SchemeGroupVersion.String()
	PostgreSQLServerVirtualNetworkRuleGroupVersionKind = SchemeGroupVersion.WithKind(PostgreSQLServerVirtualNetworkRuleKind)
)

// MySQLServerFirewallRule type metadata.
var (
	MySQLServerFirewallRuleKind             = reflect.TypeOf(MySQLServerFirewallRule{}).Name()
	MySQLServerFirewallRuleGroupKind        = schema.GroupKind{Group: Group, Kind: MySQLServerFirewallRuleKind}.String()
	MySQLServerFirewallRuleKindAPIVersion   = MySQLServerFirewallRuleKind + "." + SchemeGroupVersion.String()
	MySQLServerFirewallRuleGroupVersionKind = SchemeGroupVersion.WithKind(MySQLServerFirewallRuleKind)
)

// PostgreSQLServerFirewallRule type metadata.
var (
	PostgreSQLServerFirewallRuleKind             = reflect.TypeOf(PostgreSQLServerFirewallRule{}).Name()
	PostgreSQLServerFirewallRuleGroupKind        = schema.GroupKind{Group: Group, Kind: PostgreSQLServerFirewallRuleKind}.String()
	PostgreSQLServerFirewallRuleKindAPIVersion   = PostgreSQLServerFirewallRuleKind + "." + SchemeGroupVersion.String()
	PostgreSQLServerFirewallRuleGroupVersionKind = SchemeGroupVersion.WithKind(PostgreSQLServerFirewallRuleKind)
)

// CosmosDBAccount type metadata.
var (
	CosmosDBAccountKind             = reflect.TypeOf(CosmosDBAccount{}).Name()
	CosmosDBAccountGroupKind        = schema.GroupKind{Group: Group, Kind: CosmosDBAccountKind}.String()
	CosmosDBAccountKindAPIVersion   = CosmosDBAccountKind + "." + SchemeGroupVersion.String()
	CosmosDBAccountGroupVersionKind = SchemeGroupVersion.WithKind(CosmosDBAccountKind)
)

func init() {
	SchemeBuilder.Register(&MySQLServerVirtualNetworkRule{}, &MySQLServerVirtualNetworkRuleList{})
	SchemeBuilder.Register(&PostgreSQLServerVirtualNetworkRule{}, &PostgreSQLServerVirtualNetworkRuleList{})
	SchemeBuilder.Register(&MySQLServerFirewallRule{}, &MySQLServerFirewallRuleList{})
	SchemeBuilder.Register(&PostgreSQLServerFirewallRule{}, &PostgreSQLServerFirewallRuleList{})
	SchemeBuilder.Register(&CosmosDBAccount{}, &CosmosDBAccountList{})
}
