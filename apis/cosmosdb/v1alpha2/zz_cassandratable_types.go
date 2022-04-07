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

type CassandraTableAutoscaleSettingsObservation struct {
}

type CassandraTableAutoscaleSettingsParameters struct {

	// +kubebuilder:validation:Optional
	MaxThroughput *float64 `json:"maxThroughput,omitempty" tf:"max_throughput,omitempty"`
}

type CassandraTableObservation struct {
	ID *string `json:"id,omitempty" tf:"id,omitempty"`
}

type CassandraTableParameters struct {

	// +kubebuilder:validation:Optional
	AnalyticalStorageTTL *float64 `json:"analyticalStorageTtl,omitempty" tf:"analytical_storage_ttl,omitempty"`

	// +kubebuilder:validation:Optional
	AutoscaleSettings []CassandraTableAutoscaleSettingsParameters `json:"autoscaleSettings,omitempty" tf:"autoscale_settings,omitempty"`

	// +crossplane:generate:reference:type=CassandraKeySpace
	// +crossplane:generate:reference:extractor=github.com/crossplane/provider-azure/apis/rconfig.ExtractResourceID()
	// +kubebuilder:validation:Optional
	CassandraKeySpaceID *string `json:"cassandraKeyspaceId,omitempty" tf:"cassandra_keyspace_id,omitempty"`

	// +kubebuilder:validation:Optional
	CassandraKeySpaceIDRef *v1.Reference `json:"cassandraKeySpaceIdRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	CassandraKeySpaceIDSelector *v1.Selector `json:"cassandraKeySpaceIdSelector,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	DefaultTTL *float64 `json:"defaultTtl,omitempty" tf:"default_ttl,omitempty"`

	// +kubebuilder:validation:Required
	Schema []SchemaParameters `json:"schema" tf:"schema,omitempty"`

	// +kubebuilder:validation:Optional
	Throughput *float64 `json:"throughput,omitempty" tf:"throughput,omitempty"`
}

type ClusterKeyObservation struct {
}

type ClusterKeyParameters struct {

	// +kubebuilder:validation:Required
	Name *string `json:"name" tf:"name,omitempty"`

	// +kubebuilder:validation:Required
	OrderBy *string `json:"orderBy" tf:"order_by,omitempty"`
}

type ColumnObservation struct {
}

type ColumnParameters struct {

	// +kubebuilder:validation:Required
	Name *string `json:"name" tf:"name,omitempty"`

	// +kubebuilder:validation:Required
	Type *string `json:"type" tf:"type,omitempty"`
}

type PartitionKeyObservation struct {
}

type PartitionKeyParameters struct {

	// +kubebuilder:validation:Required
	Name *string `json:"name" tf:"name,omitempty"`
}

type SchemaObservation struct {
}

type SchemaParameters struct {

	// +kubebuilder:validation:Optional
	ClusterKey []ClusterKeyParameters `json:"clusterKey,omitempty" tf:"cluster_key,omitempty"`

	// +kubebuilder:validation:Required
	Column []ColumnParameters `json:"column" tf:"column,omitempty"`

	// +kubebuilder:validation:Required
	PartitionKey []PartitionKeyParameters `json:"partitionKey" tf:"partition_key,omitempty"`
}

// CassandraTableSpec defines the desired state of CassandraTable
type CassandraTableSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     CassandraTableParameters `json:"forProvider"`
}

// CassandraTableStatus defines the observed state of CassandraTable.
type CassandraTableStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        CassandraTableObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// CassandraTable is the Schema for the CassandraTables API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azurejet}
type CassandraTable struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CassandraTableSpec   `json:"spec"`
	Status            CassandraTableStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CassandraTableList contains a list of CassandraTables
type CassandraTableList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CassandraTable `json:"items"`
}

// Repository type metadata.
var (
	CassandraTable_Kind             = "CassandraTable"
	CassandraTable_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: CassandraTable_Kind}.String()
	CassandraTable_KindAPIVersion   = CassandraTable_Kind + "." + CRDGroupVersion.String()
	CassandraTable_GroupVersionKind = CRDGroupVersion.WithKind(CassandraTable_Kind)
)

func init() {
	SchemeBuilder.Register(&CassandraTable{}, &CassandraTableList{})
}
