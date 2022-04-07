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

type AzureadAdministratorObservation struct {
}

type AzureadAdministratorParameters struct {

	// +kubebuilder:validation:Required
	LoginUsername *string `json:"loginUsername" tf:"login_username,omitempty"`

	// +kubebuilder:validation:Required
	ObjectID *string `json:"objectId" tf:"object_id,omitempty"`

	// +kubebuilder:validation:Optional
	TenantID *string `json:"tenantId,omitempty" tf:"tenant_id,omitempty"`
}

type ExtendedAuditingPolicyObservation struct {
}

type ExtendedAuditingPolicyParameters struct {

	// +kubebuilder:validation:Optional
	LogMonitoringEnabled *bool `json:"logMonitoringEnabled,omitempty" tf:"log_monitoring_enabled"`

	// +kubebuilder:validation:Optional
	RetentionInDays *float64 `json:"retentionInDays,omitempty" tf:"retention_in_days"`

	// +kubebuilder:validation:Optional
	StorageAccountAccessKeyIsSecondary *bool `json:"storageAccountAccessKeyIsSecondary,omitempty" tf:"storage_account_access_key_is_secondary"`

	// +kubebuilder:validation:Optional
	StorageAccountAccessKeySecretRef *v1.SecretKeySelector `json:"storageAccountAccessKeySecretRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	StorageEndpoint *string `json:"storageEndpoint,omitempty" tf:"storage_endpoint"`
}

type IdentityObservation struct {
	PrincipalID *string `json:"principalId,omitempty" tf:"principal_id,omitempty"`

	TenantID *string `json:"tenantId,omitempty" tf:"tenant_id,omitempty"`
}

type IdentityParameters struct {

	// +kubebuilder:validation:Required
	Type *string `json:"type" tf:"type,omitempty"`
}

type MSSQLServerObservation struct {
	FullyQualifiedDomainName *string `json:"fullyQualifiedDomainName,omitempty" tf:"fully_qualified_domain_name,omitempty"`

	ID *string `json:"id,omitempty" tf:"id,omitempty"`

	RestorableDroppedDatabaseIds []*string `json:"restorableDroppedDatabaseIds,omitempty" tf:"restorable_dropped_database_ids,omitempty"`
}

type MSSQLServerParameters struct {

	// +kubebuilder:validation:Required
	AdministratorLogin *string `json:"administratorLogin" tf:"administrator_login,omitempty"`

	// +kubebuilder:validation:Required
	AdministratorLoginPasswordSecretRef v1.SecretKeySelector `json:"administratorLoginPasswordSecretRef" tf:"-"`

	// +kubebuilder:validation:Optional
	AzureadAdministrator []AzureadAdministratorParameters `json:"azureadAdministrator,omitempty" tf:"azuread_administrator,omitempty"`

	// +kubebuilder:validation:Optional
	ConnectionPolicy *string `json:"connectionPolicy,omitempty" tf:"connection_policy,omitempty"`

	// +kubebuilder:validation:Optional
	ExtendedAuditingPolicy []ExtendedAuditingPolicyParameters `json:"extendedAuditingPolicy,omitempty" tf:"extended_auditing_policy,omitempty"`

	// +kubebuilder:validation:Optional
	Identity []IdentityParameters `json:"identity,omitempty" tf:"identity,omitempty"`

	// +kubebuilder:validation:Required
	Location *string `json:"location" tf:"location,omitempty"`

	// +kubebuilder:validation:Optional
	MinimumTLSVersion *string `json:"minimumTlsVersion,omitempty" tf:"minimum_tls_version,omitempty"`

	// +kubebuilder:validation:Optional
	PublicNetworkAccessEnabled *bool `json:"publicNetworkAccessEnabled,omitempty" tf:"public_network_access_enabled,omitempty"`

	// +crossplane:generate:reference:type=github.com/crossplane/provider-azure/apis/azure/v1alpha2.ResourceGroup
	// +kubebuilder:validation:Optional
	ResourceGroupName *string `json:"resourceGroupName,omitempty" tf:"resource_group_name,omitempty"`

	// +kubebuilder:validation:Optional
	ResourceGroupNameRef *v1.Reference `json:"resourceGroupNameRef,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	ResourceGroupNameSelector *v1.Selector `json:"resourceGroupNameSelector,omitempty" tf:"-"`

	// +kubebuilder:validation:Optional
	Tags map[string]*string `json:"tags,omitempty" tf:"tags,omitempty"`

	// +kubebuilder:validation:Required
	Version *string `json:"version" tf:"version,omitempty"`
}

// MSSQLServerSpec defines the desired state of MSSQLServer
type MSSQLServerSpec struct {
	v1.ResourceSpec `json:",inline"`
	ForProvider     MSSQLServerParameters `json:"forProvider"`
}

// MSSQLServerStatus defines the observed state of MSSQLServer.
type MSSQLServerStatus struct {
	v1.ResourceStatus `json:",inline"`
	AtProvider        MSSQLServerObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// MSSQLServer is the Schema for the MSSQLServers API
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azurejet}
type MSSQLServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              MSSQLServerSpec   `json:"spec"`
	Status            MSSQLServerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MSSQLServerList contains a list of MSSQLServers
type MSSQLServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MSSQLServer `json:"items"`
}

// Repository type metadata.
var (
	MSSQLServer_Kind             = "MSSQLServer"
	MSSQLServer_GroupKind        = schema.GroupKind{Group: CRDGroup, Kind: MSSQLServer_Kind}.String()
	MSSQLServer_KindAPIVersion   = MSSQLServer_Kind + "." + CRDGroupVersion.String()
	MSSQLServer_GroupVersionKind = CRDGroupVersion.WithKind(MSSQLServer_Kind)
)

func init() {
	SchemeBuilder.Register(&MSSQLServer{}, &MSSQLServerList{})
}
