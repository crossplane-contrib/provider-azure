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
	"sigs.k8s.io/controller-runtime/pkg/runtime/scheme"
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

// MySQLServer type metadata.
var (
	MySQLServerKind             = reflect.TypeOf(MySQLServer{}).Name()
	MySQLServerKindAPIVersion   = MySQLServerKind + "." + SchemeGroupVersion.String()
	MySQLServerGroupVersionKind = SchemeGroupVersion.WithKind(MySQLServerKind)
)

// MySQLServerVirtualNetworkRule type metadata.
var (
	MySQLServerVirtualNetworkRuleKind             = reflect.TypeOf(MySQLServerVirtualNetworkRule{}).Name()
	MySQLServerVirtualNetworkRuleKindAPIVersion   = MySQLServerVirtualNetworkRuleKind + "." + SchemeGroupVersion.String()
	MySQLServerVirtualNetworkRuleGroupVersionKind = SchemeGroupVersion.WithKind(MySQLServerVirtualNetworkRuleKind)
)

// PostgreSQLServer type metadata.
var (
	PostgreSQLServerKind             = reflect.TypeOf(PostgreSQLServer{}).Name()
	PostgreSQLServerKindAPIVersion   = PostgreSQLServerKind + "." + SchemeGroupVersion.String()
	PostgreSQLServerGroupVersionKind = SchemeGroupVersion.WithKind(PostgreSQLServerKind)
)

// PostgreSQLServerVirtualNetworkRule type metadata.
var (
	PostgreSQLServerVirtualNetworkRuleKind             = reflect.TypeOf(PostgreSQLServerVirtualNetworkRule{}).Name()
	PostgreSQLServerVirtualNetworkRuleKindAPIVersion   = PostgreSQLServerVirtualNetworkRuleKind + "." + SchemeGroupVersion.String()
	PostgreSQLServerVirtualNetworkRuleGroupVersionKind = SchemeGroupVersion.WithKind(PostgreSQLServerVirtualNetworkRuleKind)
)

// SQLServerClass type metadata.
var (
	SQLServerClassKind             = reflect.TypeOf(SQLServerClass{}).Name()
	SQLServerClassKindAPIVersion   = SQLServerClassKind + "." + SchemeGroupVersion.String()
	SQLServerClassGroupVersionKind = SchemeGroupVersion.WithKind(SQLServerClassKind)
)

func init() {
	SchemeBuilder.Register(&MySQLServer{}, &MySQLServerList{})
	SchemeBuilder.Register(&MySQLServerVirtualNetworkRule{}, &MySQLServerVirtualNetworkRuleList{})
	SchemeBuilder.Register(&PostgreSQLServer{}, &PostgreSQLServerList{})
	SchemeBuilder.Register(&PostgreSQLServerVirtualNetworkRule{}, &PostgreSQLServerVirtualNetworkRuleList{})
	SchemeBuilder.Register(&SQLServerClass{}, &SQLServerClassList{})
}
