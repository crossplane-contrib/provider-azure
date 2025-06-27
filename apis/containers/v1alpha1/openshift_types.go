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

package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// OpenshiftParameters are the configurable fields of a Openshift.
type OpenshiftParameters struct {
	ResourceGroupName         string            `json:"resourceGroupName,omitempty"`
	ResourceGroupNameRef      *xpv1.Reference   `json:"resourceGroupNameRef,omitempty"`
	ResourceGroupNameSelector *xpv1.Selector    `json:"resourceGroupNameSelector,omitempty"`
	Location                  string            `json:"location,omitempty"`
	Tags                      map[string]string `json:"tags,omitempty"`
	ClusterProfile            `json:"clusterProfile,omitempty"`
	ServicePrincipalProfile   `json:"servicePrincipalProfile,omitempty"`
	NetworkProfile            `json:"networkProfile,omitempty"`
	WorkerProfile             `json:"workerProfile,omitempty"`
	MasterProfile             `json:"masterProfile,omitempty"`
}

type ClusterProfile struct {
	PullSecretRef   *xpv1.SecretKeySelector `json:"pullSecretRef,omitempty"`
	PullSecret      string                  `json:"pullSecret,omitempty"`
	Domain          string                  `json:"domain,omitempty"`
	Version         string                  `json:"version,omitempty"`
	ResourceGroupID string                  `json:"resourceGroupId,omitempty"`
}

type ServicePrincipalProfile struct {
	AzureRedHatOpenShiftRPPrincipalIDRef *xpv1.SecretKeySelector `json:"azureRedHatOpenShiftRPPrincipalIDRef,omitempty"`
	AzureRedHatOpenShiftRPPrincipalID    string                  `json:"azureRedHatOpenShiftRPPrincipalID,omitempty"`
	PrincipalID                          string                  `json:"servicePrincipalId,omitempty"`
	PrincipalIDRef                       *xpv1.SecretKeySelector `json:"servicePrincipalIdRef,omitempty"`
	ClientIDRef                          *xpv1.SecretKeySelector `json:"clientIdRef,omitempty"`
	ClientSecretRef                      *xpv1.SecretKeySelector `json:"clientSecretRef,omitempty"`
	ClientID                             string                  `json:"clientId,omitempty"`
	ClientSecret                         string                  `json:"clientSecret,omitempty"`
}

type NetworkProfile struct {
	PodCidr     string `json:"podCidr,omitempty"`
	ServiceCidr string `json:"serviceCidr,omitempty"`
}

type MasterProfile struct {
	VMSize           string          `json:"vmSize,omitempty"`
	SubnetIDRef      *xpv1.Reference `json:"subnetIDRef,omitempty"`
	SubnetIDSelector *xpv1.Selector  `json:"subnetIDSelector,omitempty"`
	SubnetID         string          `json:"subnetID,omitempty"`
}

type WorkerProfile struct {
	VMSize           string          `json:"vmSize,omitempty"`
	DiskSizeGB       int             `json:"diskSizeGB,omitempty"`
	SubnetIDRef      *xpv1.Reference `json:"subnetIDRef,omitempty"`
	SubnetIDSelector *xpv1.Selector  `json:"subnetIDSelector,omitempty"`
	SubnetID         string          `json:"subnetID,omitempty"`
	Count            int             `json:"count,omitempty"`
}

// OpenshiftObservation are the observable fields of a Openshift.
type OpenshiftObservation struct {
	ProvisioningState string `json:"provisioningState,omitempty"`
	ID                string `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
}

// A OpenshiftSpec defines the desired state of a Openshift.
type OpenshiftSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       OpenshiftParameters `json:"forProvider"`
}

// A OpenshiftStatus represents the observed state of a Openshift.
type OpenshiftStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          OpenshiftObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Openshift is an example API type.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,openshift}
type Openshift struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OpenshiftSpec   `json:"spec"`
	Status OpenshiftStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OpenshiftList contains a list of Openshift
type OpenshiftList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Openshift `json:"items"`
}

// Openshift type metadata.
var (
	OpenshiftKind             = reflect.TypeOf(Openshift{}).Name()
	OpenshiftGroupKind        = schema.GroupKind{Group: Group, Kind: OpenshiftKind}.String()
	OpenshiftKindAPIVersion   = OpenshiftKind + "." + SchemeGroupVersion.String()
	OpenshiftGroupVersionKind = SchemeGroupVersion.WithKind(OpenshiftKind)
)

func init() {
	SchemeBuilder.Register(&Openshift{}, &OpenshiftList{})
}
