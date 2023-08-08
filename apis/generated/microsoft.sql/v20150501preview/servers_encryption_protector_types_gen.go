// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
// Code generated by k8s-infra-gen. DO NOT EDIT.
package v20150501preview

import (
	"github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +kubebuilder:object:root=true
//Generated from: https://schema.management.azure.com/schemas/2015-05-01-preview/Microsoft.Sql.json#/resourceDefinitions/servers_encryptionProtector
type ServersEncryptionProtector struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ServersEncryptionProtector_Spec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true
//Generated from: https://schema.management.azure.com/schemas/2015-05-01-preview/Microsoft.Sql.json#/resourceDefinitions/servers_encryptionProtector
type ServersEncryptionProtectorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ServersEncryptionProtector `json:"items"`
}

type ServersEncryptionProtector_Spec struct {
	v1alpha1.ResourceSpec `json:",inline"`
	ForProvider           ServersEncryptionProtectorParameters `json:"forProvider"`
}

type ServersEncryptionProtectorParameters struct {

	// +kubebuilder:validation:Required
	//ApiVersion: API Version of the resource type, optional when apiProfile is used
	//on the template
	ApiVersion ServersEncryptionProtectorSpecApiVersion `json:"apiVersion"`

	//Location: Location to deploy resource to
	Location *string `json:"location,omitempty"`

	// +kubebuilder:validation:Required
	//Name: Name of the resource
	Name string `json:"name"`

	// +kubebuilder:validation:Required
	//Properties: Properties for an encryption protector execution.
	Properties                EncryptionProtectorProperties `json:"properties"`
	ResourceGroupName         string                        `json:"resourceGroupName"`
	ResourceGroupNameRef      *v1alpha1.Reference           `json:"resourceGroupNameRef,omitempty"`
	ResourceGroupNameSelector *v1alpha1.Selector            `json:"resourceGroupNameSelector,omitempty"`
	ServersName               string                        `json:"serversName"`
	ServersNameRef            *v1alpha1.Reference           `json:"serversNameRef,omitempty"`
	ServersNameSelector       *v1alpha1.Selector            `json:"serversNameSelector,omitempty"`

	//Tags: Name-value pairs to add to the resource
	Tags map[string]string `json:"tags,omitempty"`

	// +kubebuilder:validation:Required
	//Type: Resource type
	Type ServersEncryptionProtectorSpecType `json:"type"`
}

//Generated from: https://schema.management.azure.com/schemas/2015-05-01-preview/Microsoft.Sql.json#/definitions/EncryptionProtectorProperties
type EncryptionProtectorProperties struct {

	//ServerKeyName: The name of the server key.
	ServerKeyName *string `json:"serverKeyName,omitempty"`

	// +kubebuilder:validation:Required
	//ServerKeyType: The encryption protector type like 'ServiceManaged',
	//'AzureKeyVault'.
	ServerKeyType EncryptionProtectorPropertiesServerKeyType `json:"serverKeyType"`
}

// +kubebuilder:validation:Enum={"2015-05-01-preview"}
type ServersEncryptionProtectorSpecApiVersion string

const ServersEncryptionProtectorSpecApiVersion20150501Preview = ServersEncryptionProtectorSpecApiVersion("2015-05-01-preview")

// +kubebuilder:validation:Enum={"Microsoft.Sql/servers/encryptionProtector"}
type ServersEncryptionProtectorSpecType string

const ServersEncryptionProtectorSpecTypeMicrosoftSqlServersEncryptionProtector = ServersEncryptionProtectorSpecType("Microsoft.Sql/servers/encryptionProtector")

// +kubebuilder:validation:Enum={"AzureKeyVault","ServiceManaged"}
type EncryptionProtectorPropertiesServerKeyType string

const (
	EncryptionProtectorPropertiesServerKeyTypeAzureKeyVault  = EncryptionProtectorPropertiesServerKeyType("AzureKeyVault")
	EncryptionProtectorPropertiesServerKeyTypeServiceManaged = EncryptionProtectorPropertiesServerKeyType("ServiceManaged")
)

func init() {
	SchemeBuilder.Register(&ServersEncryptionProtector{}, &ServersEncryptionProtectorList{})
}
