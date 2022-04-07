/*
Copyright 2021 The Crossplane Authors.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ZoneParameters define the desired state of an Azure DNS Zone.
type ZoneParameters struct {
	// ResourceGroupName specifies the name of the resource group that should
	// contain this DNS Zone.
	// +immutable
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	// +immutable
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - A selector for a ResourceGroup object to
	// retrieve its name
	// +immutable
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// Location is the Azure location that the DNS Zone will be created in
	// +kubebuilder:validation:Required
	// +immutable
	Location string `json:"location"`

	// ZoneType - Type of DNS zone to create.
	// Allowed values: Private, Public
	// Default: Public
	// +kubebuilder:validation:Enum=Public;Private
	// +kubebuilder:default=Public
	// +optional
	// +immutable
	ZoneType *string `json:"zoneType,omitempty"`

	// RegistrationVirtualNetworks - A list of references to virtual networks that register hostnames in this
	// DNS zone. This is an only when ZoneType is Private.
	// +optional
	RegistrationVirtualNetworks []SubResource `json:"registrationVirtualNetworks,omitempty"`

	// ResolutionVirtualNetworks - A list of references to virtual networks that resolve records in this DNS zone.
	// This is an only when ZoneType is Private.
	// +optional
	ResolutionVirtualNetworks []SubResource `json:"resolutionVirtualNetworks,omitempty"`

	// Tags - Resource tags.
	// +optional
	Tags map[string]*string `json:"tags,omitempty"`
}

// ZoneObservation define the actual state of an Azure DNS Zone.
type ZoneObservation struct {
	// ID - Resource ID
	ID string `json:"id,omitempty"`

	// Etag - The etag of the zone.
	Etag string `json:"etag,omitempty"`

	// Name - The name of the zone.
	Name string `json:"name,omitempty"`

	// MaxNumberOfRecordSets - The maximum number of record sets that can be created in this DNS zone.
	// This is a read-only property and any attempt to set this value will be ignored.
	MaxNumberOfRecordSets int `json:"maxNumberOfRecordSets,omitempty"`

	// NumberOfRecordSets - The current number of record sets in this DNS zone.
	// This is a read-only property and any attempt to set this value will be ignored.
	NumberOfRecordSets int `json:"numberOfRecordSets,omitempty"`

	// NameServers - The name servers for this DNS zone. This is a read-only property and any attempt
	// to set this value will be ignored.
	NameServers []string `json:"nameServers,omitempty"`

	// Type - Resource type.
	Type string `json:"type,omitempty"`
}

// SubResource a reference to another resource
type SubResource struct {
	// ID - Resource id.
	ID *string `json:"id,omitempty"`
}

// A ZoneSpec defines the desired state of a Zone.
type ZoneSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ZoneParameters `json:"forProvider"`
}

// A ZoneStatus represents the observed state of a Zone.
type ZoneStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          ZoneObservation `json:"atProvider,omitempty"`
}

// RecordType enumerates the values for record type.
type RecordType string

// The reasoning behind capitalizing DNS record type names comes from the:
// https://www.rfc-editor.org/rfc/rfc1035.html
const (
	// A type DNS record
	A RecordType = "A"
	// AAAA type DNS record
	AAAA RecordType = "AAAA"
	// CAA type DNS record
	CAA RecordType = "CAA"
	// CNAME type DNS record
	CNAME RecordType = "CNAME"
	// MX type DNS record
	MX RecordType = "MX"
	// NS type DNS record
	NS RecordType = "NS"
	// PTR type DNS record
	PTR RecordType = "PTR"
	// SOA type DNS record
	SOA RecordType = "SOA"
	// SRV type DNS record
	SRV RecordType = "SRV"
	// TXT type DNS record
	TXT RecordType = "TXT"
)

// RecordSetParameters define the desired state of an Azure DNS RecordSet.
type RecordSetParameters struct {
	// ResourceGroupName specifies the name of the resource group that should
	// contain this DNS Zone.
	// +immutable
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup object to retrieve
	// its name
	// +immutable
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - A selector for a ResourceGroup object to
	// retrieve its name
	// +immutable
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// ZoneName specifies the name of the Zone that should
	// contain this DNS RecordSet.
	// +immutable
	ZoneName string `json:"zoneName,omitempty"`

	// ZoneNameRef - A reference to a Zone object to retrieve
	// its name
	// +immutable
	ZoneNameRef *xpv1.Reference `json:"zoneNameRef,omitempty"`

	// ZoneNameSelector - A selector for a Zone object to
	// retrieve its name
	// +immutable
	ZoneNameSelector *xpv1.Selector `json:"zoneNameSelector,omitempty"`

	// Metadata - The metadata attached to the record set
	// +optional
	Metadata map[string]string `json:"metadata,omitempty"`

	// TTL - The TTL (time-to-live) of the records in the record set.
	// +kubebuilder:validation:Required
	TTL int `json:"ttl"`

	// TargetResource - A reference to an azure resource from where the dns resource value is taken.
	// +optional
	TargetResource *SubResource `json:"targetResource,omitempty"`

	// RecordType enumerates the values for record type.
	// +kubebuilder:validation:Required
	RecordType RecordType `json:"recordType"`

	// ARecords - The list of A records in the record set.
	// +optional
	ARecords []ARecord `json:"aRecords,omitempty"`

	// AAAARecords - The list of AAAA records in the record set.
	// +optional
	AAAARecords []AAAARecord `json:"aaaaRecords,omitempty"`

	// MXRecords - The list of MX records in the record set.
	// +optional
	MXRecords []MXRecord `json:"mxRecords,omitempty"`

	// NSRecords - The list of NS records in the record set.
	// +optional
	NSRecords []NSRecord `json:"nsRecords,omitempty"`

	// PTRRecords - The list of PTR records in the record set.
	// +optional
	PTRRecords []PTRRecord `json:"ptrRecords,omitempty"`

	// SRVRecords - The list of SRV records in the record set.
	// +optional
	SRVRecords []SRVRecord `json:"srvRecords,omitempty"`

	// TXTRecords - The list of TXT records in the record set.
	// +optional
	TXTRecords []TXTRecord `json:"txtRecords,omitempty"`

	// CNAMERecord - The CNAME record in the  record set.
	// +optional
	CNAMERecord CNAMERecord `json:"cnameRecord,omitempty"`

	// SOARecord - The SOA record in the record set.
	// +optional
	SOARecord SOARecord `json:"soaRecord,omitempty"`

	// CAARecords - The list of CAA records in the record set.
	// +optional
	CAARecords []CAARecord `json:"caaRecords,omitempty"`
}

// RecordSetObservation define the actual state of an Azure DNS RecordSet.
type RecordSetObservation struct {
	// ID - Resource ID
	ID string `json:"id,omitempty"`

	// Etag - The etag of the zone.
	Etag string `json:"etag,omitempty"`

	// Name - The name of the zone.
	Name string `json:"name,omitempty"`

	// Type - Resource type.
	Type string `json:"type,omitempty"`

	// FQDN - Fully qualified domain name of the record set.
	FQDN string `json:"fqdn,omitempty"`

	// ProvisioningState -provisioning State of the record set.
	ProvisioningState string `json:"provisioningState,omitempty"`
}

// ARecord an A record.
type ARecord struct {
	// IPV4Address - The IPv4 address of this A record.
	// +optional
	IPV4Address *string `json:"ipV4Address,omitempty"`
}

// AAAARecord an AAAA record.
type AAAARecord struct {
	// IPV6Address - The IPv6 address of this AAAA record.
	// +optional
	IPV6Address *string `json:"ipV6Address,omitempty"`
}

// CAARecord a CAA record.
type CAARecord struct {
	// Flags - The flags for this CAA record as an integer between 0 and 255.
	// +optional
	Flags *int `json:"flags,omitempty"`
	// Tag - The tag for this CAA record.
	// +optional
	Tag *string `json:"tag,omitempty"`
	// Value - The value for this CAA record.
	// +optional
	Value *string `json:"value,omitempty"`
}

// CNAMERecord a CNAME record.
type CNAMERecord struct {
	// CNAME - The canonical name for this CNAME record.
	// +optional
	CNAME *string `json:"cname,omitempty"`
}

// MXRecord an MX record.
type MXRecord struct {
	// Preference - The preference value for this MX record.
	// +optional
	Preference *int `json:"preference,omitempty"`
	// Exchange - The domain name of the mail host for this MX record.
	// +optional
	Exchange *string `json:"exchange,omitempty"`
}

// NSRecord an NS record.
type NSRecord struct {
	// NSDName - The name server name for this NS record.
	// +optional
	NSDName *string `json:"nsDName,omitempty"`
}

// PTRRecord a PTR record.
type PTRRecord struct {
	// PTRDName - The PTR target domain name for this PTR record.
	// +optional
	PTRDName *string `json:"ptrDName,omitempty"`
}

// SOARecord an SOA record.
type SOARecord struct {
	// Host - The domain name of the authoritative name server for this SOA record.
	// +optional
	Host *string `json:"host,omitempty"`
	// Email - The email contact for this SOA record.
	// +optional
	Email *string `json:"email,omitempty"`
	// SerialNumber - The serial number for this SOA record.
	// +optional
	SerialNumber *int `json:"serialNumber,omitempty"`
	// RefreshTime - The refresh value for this SOA record.
	// +optional
	RefreshTime *int `json:"refreshTime,omitempty"`
	// RetryTime - The retry time for this SOA record.
	// +optional
	RetryTime *int `json:"retryTime,omitempty"`
	// ExpireTime - The expire time for this SOA record.
	// +optional
	ExpireTime *int `json:"expireTime,omitempty"`
	// MinimumTTL - The minimum value for this SOA record. By convention this is used to determine the negative caching duration.
	// +optional
	MinimumTTL *int `json:"minimumTTL,omitempty"`
}

// SRVRecord an SRV record.
type SRVRecord struct {
	// Priority - The priority value for this SRV record.
	// +optional
	Priority *int `json:"priority,omitempty"`
	// Weight - The weight value for this SRV record.
	// +optional
	Weight *int `json:"weight,omitempty"`
	// Port - The port value for this SRV record.
	// +optional
	Port *int `json:"port,omitempty"`
	// Target - The target domain name for this SRV record.
	// +optional
	Target *string `json:"target,omitempty"`
}

// TXTRecord a TXT record.
type TXTRecord struct {
	// Value - The text value of this TXT record.
	// +optional
	Value []string `json:"value,omitempty"`
}

// A RecordSetSpec defines the desired state of a RecordSet.
type RecordSetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       RecordSetParameters `json:"forProvider"`
}

// A RecordSetStatus represents the observed state of a RecordSet.
type RecordSetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          RecordSetObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Zone is a managed resource that represents an Azure DNS Zone
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.version"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type Zone struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZoneSpec   `json:"spec"`
	Status ZoneStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ZoneList contains a list of Zone.
type ZoneList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zone `json:"items"`
}

// +kubebuilder:object:root=true

// A RecordSet is a managed resource that represents an Azure DNS RecordSet
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="VERSION",type="string",JSONPath=".spec.forProvider.version"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
type RecordSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RecordSetSpec   `json:"spec"`
	Status RecordSetStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RecordSetList contains a list of RecordSet.
type RecordSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RecordSet `json:"items"`
}
