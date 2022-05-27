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

package dns

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/dns/mgmt/dns"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/crossplane-contrib/provider-azure/apis/dns/v1alpha1"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
)

var (
	id       = "a-very-cool-id"
	etag     = "a-very-cool-etag"
	uid      = types.UID("definitely-a-uuid")
	location = "cool-location"
	tags     = map[string]string{"one": "test", "two": "test"}

	name                  = "verycool.online"
	maxNumberOfRecordSets = 10000
	numberOfRecordSets    = 5
	servers               = []string{"a", "b", "c"}
	typeName              = "a-very-cool-type"
	publicZone            = "Public"

	ttl            = 3600
	fqdn           = "a-very-cool-fqdn"
	state          = "Succeed"
	ip             = "1.1.1.1"
	recordStr      = "a-very-cool-rec-data"
	recordStrSlice = []string{"1", "2", "3"}
	recordInt      = 1
	recordInt64    = 1
)

func TestUpdateZoneStatusFromAzure(t *testing.T) {
	cases := []struct {
		name string
		r    dns.Zone
		want v1alpha1.ZoneObservation
	}{
		{
			name: "SuccessfulFull",
			r: dns.Zone{
				ID:   azure.ToStringPtr(id),
				Etag: azure.ToStringPtr(etag),
				Name: azure.ToStringPtr(name),
				Type: azure.ToStringPtr(typeName),
				ZoneProperties: &dns.ZoneProperties{
					MaxNumberOfRecordSets: azure.ToInt64(&maxNumberOfRecordSets),
					NumberOfRecordSets:    azure.ToInt64(&numberOfRecordSets),
					NameServers:           azure.ToStringArrayPtr(servers),
				},
			},
			want: v1alpha1.ZoneObservation{
				ID:                    id,
				Etag:                  etag,
				Name:                  name,
				Type:                  typeName,
				MaxNumberOfRecordSets: maxNumberOfRecordSets,
				NumberOfRecordSets:    numberOfRecordSets,
				NameServers:           servers,
			},
		},
		{
			name: "SuccessfulPartial",
			r: dns.Zone{
				ID: azure.ToStringPtr(id),
				ZoneProperties: &dns.ZoneProperties{
					MaxNumberOfRecordSets: azure.ToInt64(&maxNumberOfRecordSets),
					NumberOfRecordSets:    azure.ToInt64(&numberOfRecordSets),
					NameServers:           azure.ToStringArrayPtr(servers),
				},
			},
			want: v1alpha1.ZoneObservation{
				ID:                    id,
				MaxNumberOfRecordSets: maxNumberOfRecordSets,
				NumberOfRecordSets:    numberOfRecordSets,
				NameServers:           servers,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := &v1alpha1.Zone{}

			UpdateZoneStatusFromAzure(v, tc.r)

			if diff := cmp.Diff(tc.want, v.Status.AtProvider); diff != "" {
				t.Errorf("UpdateZoneStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestNewZoneParameters(t *testing.T) {
	cases := []struct {
		name string
		r    *v1alpha1.Zone
		want dns.Zone
	}{
		{
			name: "Successful",
			r: &v1alpha1.Zone{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha1.ZoneSpec{
					ForProvider: v1alpha1.ZoneParameters{
						ZoneType: &publicZone,
						Location: location,
						Tags:     azure.ToStringPtrMap(tags),
						RegistrationVirtualNetworks: []v1alpha1.SubResource{
							{ID: azure.ToStringPtr(id)},
						},
						ResolutionVirtualNetworks: []v1alpha1.SubResource{
							{ID: azure.ToStringPtr(id)},
						},
					},
				},
			},
			want: dns.Zone{
				Location: azure.ToStringPtr(location),
				Tags:     azure.ToStringPtrMap(tags),
				ZoneProperties: &dns.ZoneProperties{
					ZoneType: dns.ZoneType(publicZone),
					RegistrationVirtualNetworks: &[]dns.SubResource{
						{ID: azure.ToStringPtr(id)},
					},
					ResolutionVirtualNetworks: &[]dns.SubResource{
						{ID: azure.ToStringPtr(id)},
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewZoneParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewZoneParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestZoneIsUpToDate(t *testing.T) {
	cases := []struct {
		name string
		kube *v1alpha1.Zone
		az   dns.Zone
		want bool
	}{
		{
			name: "NotUpToDate",
			kube: &v1alpha1.Zone{
				Spec: v1alpha1.ZoneSpec{
					ForProvider: v1alpha1.ZoneParameters{
						Tags: azure.ToStringPtrMap(tags),
						RegistrationVirtualNetworks: []v1alpha1.SubResource{
							{ID: azure.ToStringPtr(id)},
						},
						ResolutionVirtualNetworks: []v1alpha1.SubResource{
							{ID: azure.ToStringPtr(id)},
						},
					},
				},
			},
			az: dns.Zone{
				Tags: nil,
				ZoneProperties: &dns.ZoneProperties{
					ResolutionVirtualNetworks: &[]dns.SubResource{
						{ID: azure.ToStringPtr(id)},
					},
				},
			},
			want: false,
		},
		{
			name: "UpToDate",
			kube: &v1alpha1.Zone{
				Spec: v1alpha1.ZoneSpec{
					ForProvider: v1alpha1.ZoneParameters{
						Tags: azure.ToStringPtrMap(tags),
						RegistrationVirtualNetworks: []v1alpha1.SubResource{
							{ID: azure.ToStringPtr(id)},
						},
						ResolutionVirtualNetworks: []v1alpha1.SubResource{
							{ID: azure.ToStringPtr(id)},
						},
					},
				},
			},
			az: dns.Zone{
				Tags: azure.ToStringPtrMap(tags),
				ZoneProperties: &dns.ZoneProperties{
					RegistrationVirtualNetworks: &[]dns.SubResource{
						{ID: azure.ToStringPtr(id)},
					},
					ResolutionVirtualNetworks: &[]dns.SubResource{
						{ID: azure.ToStringPtr(id)},
					},
				},
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ZoneIsUpToDate(tc.kube, tc.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("SubnetNeedsUpdate(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestUpdateRecordSetStatusFromAzure(t *testing.T) {
	cases := []struct {
		name string
		r    dns.RecordSet
		want v1alpha1.RecordSetObservation
	}{
		{
			name: "SuccessfulFull",
			r: dns.RecordSet{
				ID:   azure.ToStringPtr(id),
				Etag: azure.ToStringPtr(etag),
				Name: azure.ToStringPtr(name),
				Type: azure.ToStringPtr(typeName),
				RecordSetProperties: &dns.RecordSetProperties{
					ProvisioningState: azure.ToStringPtr(state),
					Fqdn:              azure.ToStringPtr(fqdn),
					Metadata:          azure.ToStringPtrMap(tags),
					TTL:               azure.ToInt64(&ttl),
					TargetResource: &dns.SubResource{
						ID: azure.ToStringPtr(id),
					},
				},
			},
			want: v1alpha1.RecordSetObservation{
				ID:                id,
				Etag:              etag,
				Name:              name,
				Type:              typeName,
				FQDN:              fqdn,
				ProvisioningState: state,
			},
		},
		{
			name: "SuccessfulPartial",
			r: dns.RecordSet{
				ID: azure.ToStringPtr(id),
				RecordSetProperties: &dns.RecordSetProperties{
					ProvisioningState: azure.ToStringPtr(state),
				},
			},
			want: v1alpha1.RecordSetObservation{
				ID:                id,
				ProvisioningState: state,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := &v1alpha1.RecordSet{}

			UpdateRecordSetStatusFromAzure(v, tc.r)

			if diff := cmp.Diff(tc.want, v.Status.AtProvider); diff != "" {
				t.Errorf("UpdateZoneStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func fillRecordSetParametersFields(r *v1alpha1.RecordSetParameters) {
	r.TTL = ttl
	r.Metadata = tags
	r.TargetResource = &v1alpha1.SubResource{
		ID: azure.ToStringPtr(id),
	}

}

func fillRecordSetFields(r *dns.RecordSet) {
	r.TTL = azure.ToInt64(&ttl)
	r.Metadata = azure.ToStringPtrMap(tags)
	r.TargetResource = &dns.SubResource{
		ID: azure.ToStringPtr(id),
	}
}

func TestNewRecordSetParameters(t *testing.T) {
	cases := []struct {
		name string
		r    *v1alpha1.RecordSetParameters
		want dns.RecordSet
	}{
		{
			name: "ARecord",
			r: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.A,
				ARecords: []v1alpha1.ARecord{
					{IPV4Address: &ip},
				},
			},

			want: dns.RecordSet{
				RecordSetProperties: &dns.RecordSetProperties{
					ARecords: &[]dns.ARecord{
						{Ipv4Address: azure.ToStringPtr(ip)},
					},
				},
			},
		},
		{
			name: "AAAARecord",
			r: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.AAAA,
				AAAARecords: []v1alpha1.AAAARecord{
					{IPV6Address: &ip},
				},
			},
			want: dns.RecordSet{
				RecordSetProperties: &dns.RecordSetProperties{
					AaaaRecords: &[]dns.AaaaRecord{
						{Ipv6Address: azure.ToStringPtr(ip)},
					},
				},
			},
		},
		{
			name: "MXRecord",
			r: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.MX,
				MXRecords: []v1alpha1.MXRecord{
					{
						Exchange:   &recordStr,
						Preference: &recordInt,
					},
				},
			},
			want: dns.RecordSet{
				RecordSetProperties: &dns.RecordSetProperties{
					MxRecords: &[]dns.MxRecord{
						{
							Exchange:   azure.ToStringPtr(recordStr),
							Preference: azure.ToInt32Ptr(recordInt),
						},
					},
				},
			},
		},
		{
			name: "NSRecord",
			r: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.NS,
				NSRecords: []v1alpha1.NSRecord{
					{NSDName: &recordStr},
				},
			},
			want: dns.RecordSet{
				RecordSetProperties: &dns.RecordSetProperties{
					NsRecords: &[]dns.NsRecord{
						{Nsdname: azure.ToStringPtr(recordStr)},
					},
				},
			},
		},
		{
			name: "PTRRecord",
			r: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.PTR,
				PTRRecords: []v1alpha1.PTRRecord{
					{PTRDName: &recordStr},
				},
			},
			want: dns.RecordSet{
				RecordSetProperties: &dns.RecordSetProperties{
					PtrRecords: &[]dns.PtrRecord{
						{Ptrdname: azure.ToStringPtr(recordStr)},
					},
				},
			},
		},
		{
			name: "SRVRecord",
			r: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.SRV,
				SRVRecords: []v1alpha1.SRVRecord{
					{
						Priority: &recordInt,
						Weight:   &recordInt,
						Port:     &recordInt,
						Target:   &recordStr,
					},
				},
			},
			want: dns.RecordSet{
				RecordSetProperties: &dns.RecordSetProperties{
					SrvRecords: &[]dns.SrvRecord{
						{
							Priority: azure.ToInt32Ptr(recordInt),
							Weight:   azure.ToInt32Ptr(recordInt),
							Port:     azure.ToInt32Ptr(recordInt),
							Target:   azure.ToStringPtr(recordStr),
						},
					},
				},
			},
		},
		{
			name: "TXTRecord",
			r: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.TXT,
				TXTRecords: []v1alpha1.TXTRecord{
					{Value: recordStrSlice},
				},
			},
			want: dns.RecordSet{
				RecordSetProperties: &dns.RecordSetProperties{
					TxtRecords: &[]dns.TxtRecord{
						{Value: azure.ToStringArrayPtr(recordStrSlice)},
					},
				},
			},
		},
		{
			name: "CNameRecord",
			r: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.CNAME,
				CNAMERecord: v1alpha1.CNAMERecord{
					CNAME: &recordStr,
				},
			},
			want: dns.RecordSet{
				RecordSetProperties: &dns.RecordSetProperties{
					CnameRecord: &dns.CnameRecord{
						Cname: azure.ToStringPtr(recordStr),
					},
				},
			},
		},
		{
			name: "CAARecord",
			r: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.CAA,
				CAARecords: []v1alpha1.CAARecord{
					{
						Flags: &recordInt,
						Tag:   &recordStr,
						Value: &recordStr,
					},
				},
			},
			want: dns.RecordSet{
				RecordSetProperties: &dns.RecordSetProperties{
					CaaRecords: &[]dns.CaaRecord{
						{
							Flags: azure.ToInt32Ptr(recordInt),
							Tag:   azure.ToStringPtr(recordStr),
							Value: azure.ToStringPtr(recordStr),
						},
					},
				},
			},
		},
		{
			name: "SOARecord",
			r: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.SOA,
				SOARecord: v1alpha1.SOARecord{
					Host:         &recordStr,
					Email:        &recordStr,
					SerialNumber: &recordInt64,
					RefreshTime:  &recordInt64,
					RetryTime:    &recordInt64,
					ExpireTime:   &recordInt64,
					MinimumTTL:   &recordInt64,
				},
			},
			want: dns.RecordSet{
				RecordSetProperties: &dns.RecordSetProperties{
					SoaRecord: &dns.SoaRecord{
						Host:         azure.ToStringPtr(recordStr),
						Email:        &recordStr,
						SerialNumber: azure.ToInt64(&recordInt),
						RefreshTime:  azure.ToInt64(&recordInt),
						RetryTime:    azure.ToInt64(&recordInt),
						ExpireTime:   azure.ToInt64(&recordInt),
						MinimumTTL:   azure.ToInt64(&recordInt),
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fillRecordSetParametersFields(tc.r)
			fillRecordSetFields(&tc.want)

			got := NewRecordSetParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewZoneParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestRecordSetIsUpToDate(t *testing.T) {
	cases := []struct {
		name string
		kube *v1alpha1.RecordSetParameters
		az   *dns.RecordSetProperties
		want bool
	}{
		{
			name: "NotUpToDate",
			kube: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.A,
				ARecords: []v1alpha1.ARecord{
					{IPV4Address: &ip},
				},
			},
			az: &dns.RecordSetProperties{
				ARecords: &[]dns.ARecord{},
			},
			want: false,
		},
		{
			name: "UpToDate",
			kube: &v1alpha1.RecordSetParameters{
				RecordType: v1alpha1.A,
				ARecords: []v1alpha1.ARecord{
					{IPV4Address: &ip},
				},
			},
			az: &dns.RecordSetProperties{
				Metadata: azure.ToStringPtrMap(tags),
				TTL:      azure.ToInt64(&ttl),
				TargetResource: &dns.SubResource{
					ID: azure.ToStringPtr(id),
				},
				ARecords: &[]dns.ARecord{
					{Ipv4Address: azure.ToStringPtr(ip)},
				},
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fillRecordSetParametersFields(tc.kube)

			got := RecordSetIsUpToDate(tc.kube, tc.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("SubnetNeedsUpdate(...): -want, +got\n%s", diff)
			}
		})
	}
}
