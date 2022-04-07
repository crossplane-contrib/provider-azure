package dns

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	azure "github.com/crossplane/provider-azure/internal/pkg/clients"

	"github.com/crossplane/provider-azure/apis/classic/dns/v1alpha1"
)

const (
	// RecordSetSuccessfulState represents a RecordSet that is ready and has a successful last operation.
	RecordSetSuccessfulState = "Succeeded"
	// RecordSetDeletingState represents a RecordSet that is deleting.
	RecordSetDeletingState = "Deleting"
)

// ZoneAPI represents the API interface for a DNS Zone client
type ZoneAPI interface {
	Get(ctx context.Context, z *v1alpha1.Zone) (dns.Zone, error)
	CreateOrUpdate(ctx context.Context, z *v1alpha1.Zone) error
	Delete(ctx context.Context, z *v1alpha1.Zone) error
}

// ZoneClient is the concrete implementation of the ZoneAP interface for DNS Zone that calls Azure API.
type ZoneClient struct {
	dns.ZonesClient
}

// NewZoneClient creates and initializes a ZoneClient instance.
func NewZoneClient(cl dns.ZonesClient) *ZoneClient {
	return &ZoneClient{
		ZonesClient: cl,
	}
}

// Get retrieves the requested DNS Zone
func (c *ZoneClient) Get(ctx context.Context, z *v1alpha1.Zone) (dns.Zone, error) {
	return c.ZonesClient.Get(ctx, z.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(z))
}

// CreateOrUpdate creates or updates a DNS Zone
func (c *ZoneClient) CreateOrUpdate(ctx context.Context, z *v1alpha1.Zone) error {
	_, err := c.ZonesClient.CreateOrUpdate(ctx, z.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(z),
		NewZoneParameters(z), "", "")

	return err
}

// Delete deletes the given DNS Zone
func (c *ZoneClient) Delete(ctx context.Context, z *v1alpha1.Zone) error {
	_, err := c.ZonesClient.Delete(ctx, z.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(z), z.Status.AtProvider.Etag)

	return err
}

// UpdateZoneStatusFromAzure updates the status related to the external
// Azure DNS Zone in the ZoneStatus
func UpdateZoneStatusFromAzure(v *v1alpha1.Zone, az dns.Zone) {
	v.Status.AtProvider.ID = azure.ToString(az.ID)
	v.Status.AtProvider.Etag = azure.ToString(az.Etag)
	v.Status.AtProvider.Name = azure.ToString(az.Name)
	v.Status.AtProvider.Type = azure.ToString(az.Type)
	v.Status.AtProvider.MaxNumberOfRecordSets = azure.Int64ToInt(az.MaxNumberOfRecordSets)
	v.Status.AtProvider.NameServers = azure.ToStringArray(az.NameServers)
	v.Status.AtProvider.NumberOfRecordSets = azure.Int64ToInt(az.NumberOfRecordSets)
}

// NewZoneParameters returns an Azure DNS Zone object.
func NewZoneParameters(r *v1alpha1.Zone) dns.Zone {
	res := dns.Zone{
		Name:           azure.ToStringPtr(meta.GetExternalName(r)),
		ZoneProperties: &dns.ZoneProperties{},
	}

	if r.Spec.ForProvider.ZoneType != nil {
		res.ZoneType = dns.ZoneType(*r.Spec.ForProvider.ZoneType)
	}

	res.RegistrationVirtualNetworks = newSubResourceParameters(r.Spec.ForProvider.RegistrationVirtualNetworks)
	res.ResolutionVirtualNetworks = newSubResourceParameters(r.Spec.ForProvider.RegistrationVirtualNetworks)

	res.Location = &r.Spec.ForProvider.Location

	// This empty initialization necessary because, Azure SDK returns an empty struct when this field is set to nil.
	res.Tags = map[string]*string{}
	if r.Spec.ForProvider.Tags != nil {
		res.Tags = r.Spec.ForProvider.Tags
	}

	return res
}

func newSubResourceParameters(e []v1alpha1.SubResource) *[]dns.SubResource {
	if e != nil {
		endpoints := make([]dns.SubResource, len(e))

		for i, s := range e {
			endpoints[i] = dns.SubResource{
				ID: s.ID,
			}
		}

		return &endpoints
	}

	return nil
}

// ZoneIsUpToDate decides if an upgrade is needed.
func ZoneIsUpToDate(r *v1alpha1.Zone, az dns.Zone) bool {
	up := NewZoneParameters(r)
	if !cmp.Equal(up.Tags, az.Tags) || !cmp.Equal(up.RegistrationVirtualNetworks, az.RegistrationVirtualNetworks) ||
		!cmp.Equal(up.ResolutionVirtualNetworks, az.ResolutionVirtualNetworks) {
		return false
	}

	return true
}

// RecordSetAPI represents the API interface for a DNS RecordSet client
type RecordSetAPI interface {
	Get(ctx context.Context, r *v1alpha1.RecordSet) (dns.RecordSet, error)
	CreateOrUpdate(ctx context.Context, r *v1alpha1.RecordSet) error
	Delete(ctx context.Context, r *v1alpha1.RecordSet) error
}

// RecordSetClient is the concrete implementation of the ZoneAP interface for RecordSet that calls Azure API.
type RecordSetClient struct {
	dns.RecordSetsClient
}

// NewRecordSetClient creates and initializes a RecordSet instance.
func NewRecordSetClient(cl dns.RecordSetsClient) *RecordSetClient {
	return &RecordSetClient{
		RecordSetsClient: cl,
	}
}

// Get retrieves the requested DNS RecordSet
func (c *RecordSetClient) Get(ctx context.Context, r *v1alpha1.RecordSet) (dns.RecordSet, error) {
	return c.RecordSetsClient.Get(ctx, r.Spec.ForProvider.ResourceGroupName, r.Spec.ForProvider.ZoneName, meta.GetExternalName(r),
		dns.RecordType(r.Spec.ForProvider.RecordType))
}

// CreateOrUpdate creates or updates a DNS RecordSet
func (c *RecordSetClient) CreateOrUpdate(ctx context.Context, r *v1alpha1.RecordSet) error {
	_, err := c.RecordSetsClient.CreateOrUpdate(ctx, r.Spec.ForProvider.ResourceGroupName, r.Spec.ForProvider.ZoneName,
		meta.GetExternalName(r), dns.RecordType(r.Spec.ForProvider.RecordType), NewRecordSetParameters(&r.Spec.ForProvider),
		"", "")

	return err
}

// Delete deletes the given DNS RecordSet
func (c *RecordSetClient) Delete(ctx context.Context, r *v1alpha1.RecordSet) error {
	_, err := c.RecordSetsClient.Delete(ctx, r.Spec.ForProvider.ResourceGroupName, r.Spec.ForProvider.ZoneName,
		meta.GetExternalName(r), dns.RecordType(r.Spec.ForProvider.RecordType), r.Status.AtProvider.Etag)

	return err
}

// UpdateRecordSetStatusFromAzure updates the status related to the external
// Azure DNS RecordSet in the RecordSetStatus
func UpdateRecordSetStatusFromAzure(v *v1alpha1.RecordSet, az dns.RecordSet) {
	v.Status.AtProvider.ID = azure.ToString(az.ID)
	v.Status.AtProvider.Etag = azure.ToString(az.Etag)
	v.Status.AtProvider.Name = azure.ToString(az.Name)
	v.Status.AtProvider.Type = azure.ToString(az.Type)
	v.Status.AtProvider.ProvisioningState = azure.ToString(az.ProvisioningState)
	v.Status.AtProvider.FQDN = azure.ToString(az.Fqdn)
}

// NewRecordSetParameters returns an Azure DNS RecordSet object.
func NewRecordSetParameters(r *v1alpha1.RecordSetParameters) dns.RecordSet {
	res := dns.RecordSet{
		RecordSetProperties: &dns.RecordSetProperties{
			Metadata: azure.ToStringPtrMap(r.Metadata),
			TTL:      azure.ToInt64(&r.TTL),
			// This empty initialization necessary because, Azure SDK returns an empty struct when this field is set to nil.
			TargetResource: &dns.SubResource{},
		},
	}

	if r.TargetResource != nil {
		res.TargetResource = &dns.SubResource{
			ID: r.TargetResource.ID,
		}
	}

	newRecordParameters(r, &res)

	return res
}

// nolint: gocyclo
func newRecordParameters(r *v1alpha1.RecordSetParameters, res *dns.RecordSet) {
	if len(r.ARecords) > 0 {
		var recordsA []dns.ARecord
		for _, rec := range r.ARecords {
			recordsA = append(recordsA, dns.ARecord{
				Ipv4Address: rec.IPV4Address,
			})
		}
		res.ARecords = &recordsA
	}

	if len(r.AAAARecords) > 0 {
		var recordsAAAA []dns.AaaaRecord
		for _, rec := range r.AAAARecords {
			recordsAAAA = append(recordsAAAA, dns.AaaaRecord{
				Ipv6Address: rec.IPV6Address,
			})
		}
		res.AaaaRecords = &recordsAAAA
	}

	if len(r.CAARecords) > 0 {
		var recordsCAA []dns.CaaRecord
		for _, rec := range r.CAARecords {
			recordsCAA = append(recordsCAA, dns.CaaRecord{
				Flags: azure.ToInt32PtrFromIntPtr(rec.Flags),
				Tag:   rec.Tag,
				Value: rec.Value,
			})
		}
		res.CaaRecords = &recordsCAA
	}

	if len(r.MXRecords) > 0 {
		var recordsMX []dns.MxRecord
		for _, rec := range r.MXRecords {
			recordsMX = append(recordsMX, dns.MxRecord{
				Preference: azure.ToInt32PtrFromIntPtr(rec.Preference),
				Exchange:   rec.Exchange,
			})
		}
		res.MxRecords = &recordsMX
	}

	if len(r.NSRecords) > 0 {
		var recordsNS []dns.NsRecord
		for _, rec := range r.NSRecords {
			recordsNS = append(recordsNS, dns.NsRecord{
				Nsdname: rec.NSDName,
			})
		}
		res.NsRecords = &recordsNS
	}

	if len(r.PTRRecords) > 0 {
		var recordsPTR []dns.PtrRecord
		for _, rec := range r.PTRRecords {
			recordsPTR = append(recordsPTR, dns.PtrRecord{
				Ptrdname: rec.PTRDName,
			})
		}
		res.PtrRecords = &recordsPTR
	}

	if len(r.SRVRecords) > 0 {
		var recordsSRV []dns.SrvRecord
		for _, rec := range r.SRVRecords {
			recordsSRV = append(recordsSRV, dns.SrvRecord{
				Priority: azure.ToInt32PtrFromIntPtr(rec.Priority),
				Weight:   azure.ToInt32PtrFromIntPtr(rec.Weight),
				Port:     azure.ToInt32PtrFromIntPtr(rec.Port),
				Target:   rec.Target,
			})
		}
		res.SrvRecords = &recordsSRV
	}

	if len(r.TXTRecords) > 0 {
		var recordsTXT []dns.TxtRecord
		for _, rec := range r.TXTRecords {
			recordsTXT = append(recordsTXT, dns.TxtRecord{
				Value: to.StringSlicePtr(rec.Value),
			})
		}
		res.TxtRecords = &recordsTXT
	}

	if r.CNAMERecord.CNAME != nil {
		res.CnameRecord = &dns.CnameRecord{
			Cname: r.CNAMERecord.CNAME,
		}
	}

	if r.SOARecord.Host != nil {
		res.SoaRecord = &dns.SoaRecord{
			Host:         r.SOARecord.Host,
			Email:        r.SOARecord.Email,
			SerialNumber: azure.ToInt64(r.SOARecord.SerialNumber),
			RefreshTime:  azure.ToInt64(r.SOARecord.RefreshTime),
			RetryTime:    azure.ToInt64(r.SOARecord.RetryTime),
			ExpireTime:   azure.ToInt64(r.SOARecord.ExpireTime),
			MinimumTTL:   azure.ToInt64(r.SOARecord.MinimumTTL),
		}
	}
}

// RecordSetIsUpToDate decides if an upgrade is needed.
func RecordSetIsUpToDate(r *v1alpha1.RecordSetParameters, az *dns.RecordSetProperties) bool {
	up := NewRecordSetParameters(r)

	return cmp.Equal(up.RecordSetProperties, az,
		cmpopts.IgnoreFields(dns.RecordSetProperties{}, "Fqdn", "ProvisioningState"))
}
