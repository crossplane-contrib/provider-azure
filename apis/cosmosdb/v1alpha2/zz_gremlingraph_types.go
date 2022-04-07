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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	v1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

type CompositeIndexObservation struct {
}

type CompositeIndexParameters struct {

	// +kubebuilder:validation:Required
	Index []IndexParameters `json:"index" tf:"index,omitempty"`
}

type ConflictResolutionPolicyObservation struct {
}

type ConflictResolutionPolicyParameters struct {

	// +kubebuilder:validation:Optional
	ConflictResolutionPath *string `json:"conflictResolutionPath,omitempty" tf:"conflict_resolution_path,omitempty"`

	// +kubebuilder:validation:Optional
	ConflictResolutionProcedure *string `json:"conflictResolutionProcedure,omitempty" tf:"conflict_resolution_procedure,omitempty"`

	// +kubebuilder:validation:Required
	Mode *string `json:"mode" tf:"mode,omitempty"`
}

type GremlinGraphAutoscaleSettingsObservation struct {
}

type GremlinGraphAutoscaleSettingsParameters struct {

	// +kubebuilder:validation:Optional
	MaxThroughput *float64 `json:"maxThroughput,omitempty" tf:"max_throughput,omitempty"`
}

type GremlinGraphObservation struct {
	ID *string `json:"id,omitempty" tf:"id,omitempty"`
}

type GremlinGraphParameters struct {

	// +crossplane:generate:reference:type=Account
	// +kubebuilder:validation:Optional
	AccountName *string `json:"accountName,omitempty" tf:"account_name,omitempty"`

	// +kubebuilder:validation:Optional
	AccountNameRef *v1.Reference `json:"accountNameRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	AccountNameSelector *v1.Selector `json:"accountNameSelector,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	AutoscaleSettings []GremlinGraphAutoscaleSettingsParameters `json:"autoscaleSettings,omitempty" tf:"autoscale_settings,omitempty"`

	// +kubebuilder:validation:Optional
	ConflictResolutionPolicy []ConflictResolutionPolicyParameters `json:"conflictResolutionPolicy,omitempty" tf:"conflict_resolution_policy,omitempty"`

	// +crossplane:generate:reference:type=GremlinDatabase
	// +kubebuilder:validation:Optional
	DatabaseName *string `json:"databaseName,omitempty" tf:"database_name,omitempty"`

	// +kubebuilder:validation:Optional
	DatabaseNameRef *v1.Reference `json:"databaseNameRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	DatabaseNameSelector *v1.Selector `json:"databaseNameSelector,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	DefaultTTL *float64 `json:"defaultTtl,omitempty" tf:"default_ttl,omitempty"`

	// +kubebuilder:validation:Optional
	IndexPolicy []IndexPolicyParameters `json:"indexPolicy,omitempty" tf:"index_policy,omitempty"`

	// +kubebuilder:validation:Required
	PartitionKeyPath *string `json:"partitionKeyPath" tf:"partition_key_path,omitempty"`

	// +kubebuilder:validation:Optional
	PartitionKeyVersion *float64 `json:"partitionKeyVersion,omitempty" tf:"partition_key_version,omitempty"`

	// +crossplane:generate:reference:type=github.com/crossplane/provider-azure/apis/azure/v1alpha2.ResourceGroup
	// +kubebuilder:validation:Optional
	ResourceGroupName *string `json:"resourceGroupName,omitempty" tf:"resource_group_name,omitempty"`

	// +kubebuilder:validation:Optional
	ResourceGroupNameRef *v1.Reference `json:"resourceGroupNameRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	ResourceGroupNameSelector *v1.Selector `json:"resourceGroupNameSelector,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	Throughput *float64 `json:"throughput,omitempty" tf:"throughput,omitempty"`

	// +kubebuilder:validation:Optional
	UniqueKey []UniqueKeyParameters `json:"uniqueKey,omitempty" tf:"unique_key,omitempty"`
}

type IndexObservation struct {
}

type IndexParameters struct {

	// +kubebuilder:validation:Required
	Order *string `json:"order" tf:"order,omitempty"`

	// +kubebuilder:validation:Required
	Path *string `json:"path" tf:"path,omitempty"`
}

type IndexPolicyObservation struct {
}

type IndexPolicyParameters struct {

	// +kubebuilder:validation:Optional
	Automatic *bool `json:"automatic,omitempty" tf:"automatic,omitempty"`

	// +kubebuilder:validation:Optional
	CompositeIndex []CompositeIndexParameters `json:"compositeIndex,omitempty" tf:"composite_index,omitempty"`

	// +kubebuilder:validation:Optional
	ExcludedPaths []*string `json:"excludedPaths,omitempty" tf:"excluded_paths,omitempty"`

	// +kubebuilder:validation:Optional
	IncludedPaths []*string `json:"includedPaths,omitempty" tf:"included_paths,omitempty"`

	// +kubebuilder:validation:Required
	IndexingMode *string `json:"indexingMode" tf:"indexing_mode,omitempty"`

	// +kubebuilder:validation:Optional
	SpatialIndex []SpatialIndexParameters `json:"spatialIndex,omitempty" tf:"spatial_index,omitempty"`
}

type SpatialIndexObservation struct {
	Types []*string `json:"types,omitempty" tf:"types,omitempty"`
}

type SpatialIndexParameters struct {

	// +kubebuilder:validation:Required
	Path *string `json:"path" tf:"path,omitempty"`
}

type UniqueKeyObservation struct {
}

type UniqueKeyParameters struct {

	// +kubebuilder:validation:Required
	Paths []*string `json:"paths" tf:"paths,omitempty"`
}

// GremlinGraphSpec defines the desired state of GremlinGraph
type GremlinGraphSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     GremlinGraphParameters `json:"forProvider"`
}

// GremlinGraphStatus defines the observed state of GremlinGraph.
type GremlinGraphStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        GremlinGraphObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// GremlinGraph is the Schema for the GremlinGraphs API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azurejet}
type GremlinGraph struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              GremlinGraphSpec   `json:"spec"`
	Status            GremlinGraphStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GremlinGraphList contains a list of GremlinGraphs
type GremlinGraphList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GremlinGraph `json:"items"`
}

// Repository type metadata.
var (
	GremlinGraph_Kind             = "GremlinGraph"
	GremlinGraph_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: GremlinGraph_Kind}.String()
	GremlinGraph_KindAPIVersion   = GremlinGraph_Kind + "." + CRDGroupVersion.String()
	GremlinGraph_GroupVersionKind = CRDGroupVersion.WithKind(GremlinGraph_Kind)
)

func init() {
	SchemeBuilder.Register(&GremlinGraph{}, &GremlinGraphList{})
}
