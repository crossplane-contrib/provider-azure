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

package v1beta1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "database.azure.crossplane.io"
	Version = "v1beta1"
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
	MySQLServerGroupKind        = schema.GroupKind{Group: Group, Kind: MySQLServerKind}.String()
	MySQLServerKindAPIVersion   = MySQLServerKind + "." + SchemeGroupVersion.String()
	MySQLServerGroupVersionKind = SchemeGroupVersion.WithKind(MySQLServerKind)
)

// MySQLServerConfiguration type metadata.
var (
	MySQLServerConfigurationKind             = reflect.TypeOf(MySQLServerConfiguration{}).Name()
	MySQLServerConfigurationGroupKind        = schema.GroupKind{Group: Group, Kind: MySQLServerConfigurationKind}.String()
	MySQLServerConfigurationKindAPIVersion   = MySQLServerConfigurationKind + "." + SchemeGroupVersion.String()
	MySQLServerConfigurationGroupVersionKind = SchemeGroupVersion.WithKind(MySQLServerConfigurationKind)
)

// PostgreSQLServer type metadata.
var (
	PostgreSQLServerKind             = reflect.TypeOf(PostgreSQLServer{}).Name()
	PostgreSQLServerGroupKind        = schema.GroupKind{Group: Group, Kind: PostgreSQLServerKind}.String()
	PostgreSQLServerKindAPIVersion   = PostgreSQLServerKind + "." + SchemeGroupVersion.String()
	PostgreSQLServerGroupVersionKind = SchemeGroupVersion.WithKind(PostgreSQLServerKind)
)

// PostgreSQLServerConfiguration type metadata.
var (
	PostgreSQLServerConfigurationKind             = reflect.TypeOf(PostgreSQLServerConfiguration{}).Name()
	PostgreSQLServerConfigurationGroupKind        = schema.GroupKind{Group: Group, Kind: PostgreSQLServerConfigurationKind}.String()
	PostgreSQLServerConfigurationKindAPIVersion   = PostgreSQLServerConfigurationKind + "." + SchemeGroupVersion.String()
	PostgreSQLServerConfigurationGroupVersionKind = SchemeGroupVersion.WithKind(PostgreSQLServerConfigurationKind)
)

func init() {
	SchemeBuilder.Register(&MySQLServer{}, &MySQLServerList{})
	SchemeBuilder.Register(&MySQLServerConfiguration{}, &MySQLServerConfigurationList{})
	SchemeBuilder.Register(&PostgreSQLServer{}, &PostgreSQLServerList{})
	SchemeBuilder.Register(&PostgreSQLServerConfiguration{}, &PostgreSQLServerConfigurationList{})
}
