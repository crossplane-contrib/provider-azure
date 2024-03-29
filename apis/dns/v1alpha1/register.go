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

package v1alpha1

import (
	"reflect"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// Package type metadata.
const (
	Group   = "dns.azure.crossplane.io"
	Version = "v1alpha1"
)

var (
	// SchemeGroupVersion is group version used to register these objects
	SchemeGroupVersion = schema.GroupVersion{Group: Group, Version: Version}

	// SchemeBuilder is used to add go types to the GroupVersionKind scheme
	SchemeBuilder = &scheme.Builder{GroupVersion: SchemeGroupVersion}
)

// Zone type metadata.
var (
	ZoneKind             = reflect.TypeOf(Zone{}).Name()
	ZoneGroupKind        = schema.GroupKind{Group: Group, Kind: ZoneKind}.String()
	ZoneKindAPIVersion   = ZoneKind + "." + SchemeGroupVersion.String()
	ZoneGroupVersionKind = SchemeGroupVersion.WithKind(ZoneKind)
)

// RecordSet type metadata.
var (
	RecordSetKind             = reflect.TypeOf(RecordSet{}).Name()
	RecordSetGroupKind        = schema.GroupKind{Group: Group, Kind: RecordSetKind}.String()
	RecordSetKindAPIVersion   = RecordSetKind + "." + SchemeGroupVersion.String()
	RecordSetGroupVersionKind = SchemeGroupVersion.WithKind(RecordSetKind)
)

func init() {
	SchemeBuilder.Register(&Zone{}, &ZoneList{})
	SchemeBuilder.Register(&RecordSet{}, &RecordSetList{})
}
