//go:build !ignore_autogenerated
// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	"github.com/crossplane/crossplane-runtime/apis/common/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *AAAARecord) DeepCopyInto(out *AAAARecord) {
	*out = *in
	if in.IPV6Address != nil {
		in, out := &in.IPV6Address, &out.IPV6Address
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new AAAARecord.
func (in *AAAARecord) DeepCopy() *AAAARecord {
	if in == nil {
		return nil
	}
	out := new(AAAARecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ARecord) DeepCopyInto(out *ARecord) {
	*out = *in
	if in.IPV4Address != nil {
		in, out := &in.IPV4Address, &out.IPV4Address
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ARecord.
func (in *ARecord) DeepCopy() *ARecord {
	if in == nil {
		return nil
	}
	out := new(ARecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CAARecord) DeepCopyInto(out *CAARecord) {
	*out = *in
	if in.Flags != nil {
		in, out := &in.Flags, &out.Flags
		*out = new(int)
		**out = **in
	}
	if in.Tag != nil {
		in, out := &in.Tag, &out.Tag
		*out = new(string)
		**out = **in
	}
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CAARecord.
func (in *CAARecord) DeepCopy() *CAARecord {
	if in == nil {
		return nil
	}
	out := new(CAARecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CNAMERecord) DeepCopyInto(out *CNAMERecord) {
	*out = *in
	if in.CNAME != nil {
		in, out := &in.CNAME, &out.CNAME
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CNAMERecord.
func (in *CNAMERecord) DeepCopy() *CNAMERecord {
	if in == nil {
		return nil
	}
	out := new(CNAMERecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MXRecord) DeepCopyInto(out *MXRecord) {
	*out = *in
	if in.Preference != nil {
		in, out := &in.Preference, &out.Preference
		*out = new(int)
		**out = **in
	}
	if in.Exchange != nil {
		in, out := &in.Exchange, &out.Exchange
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MXRecord.
func (in *MXRecord) DeepCopy() *MXRecord {
	if in == nil {
		return nil
	}
	out := new(MXRecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *NSRecord) DeepCopyInto(out *NSRecord) {
	*out = *in
	if in.NSDName != nil {
		in, out := &in.NSDName, &out.NSDName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new NSRecord.
func (in *NSRecord) DeepCopy() *NSRecord {
	if in == nil {
		return nil
	}
	out := new(NSRecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PTRRecord) DeepCopyInto(out *PTRRecord) {
	*out = *in
	if in.PTRDName != nil {
		in, out := &in.PTRDName, &out.PTRDName
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PTRRecord.
func (in *PTRRecord) DeepCopy() *PTRRecord {
	if in == nil {
		return nil
	}
	out := new(PTRRecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecordSet) DeepCopyInto(out *RecordSet) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecordSet.
func (in *RecordSet) DeepCopy() *RecordSet {
	if in == nil {
		return nil
	}
	out := new(RecordSet)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RecordSet) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecordSetList) DeepCopyInto(out *RecordSetList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]RecordSet, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecordSetList.
func (in *RecordSetList) DeepCopy() *RecordSetList {
	if in == nil {
		return nil
	}
	out := new(RecordSetList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RecordSetList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecordSetObservation) DeepCopyInto(out *RecordSetObservation) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecordSetObservation.
func (in *RecordSetObservation) DeepCopy() *RecordSetObservation {
	if in == nil {
		return nil
	}
	out := new(RecordSetObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecordSetParameters) DeepCopyInto(out *RecordSetParameters) {
	*out = *in
	if in.ResourceGroupNameRef != nil {
		in, out := &in.ResourceGroupNameRef, &out.ResourceGroupNameRef
		*out = new(v1.Reference)
		**out = **in
	}
	if in.ResourceGroupNameSelector != nil {
		in, out := &in.ResourceGroupNameSelector, &out.ResourceGroupNameSelector
		*out = new(v1.Selector)
		(*in).DeepCopyInto(*out)
	}
	if in.ZoneNameRef != nil {
		in, out := &in.ZoneNameRef, &out.ZoneNameRef
		*out = new(v1.Reference)
		**out = **in
	}
	if in.ZoneNameSelector != nil {
		in, out := &in.ZoneNameSelector, &out.ZoneNameSelector
		*out = new(v1.Selector)
		(*in).DeepCopyInto(*out)
	}
	if in.Metadata != nil {
		in, out := &in.Metadata, &out.Metadata
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.TargetResource != nil {
		in, out := &in.TargetResource, &out.TargetResource
		*out = new(SubResource)
		(*in).DeepCopyInto(*out)
	}
	if in.ARecords != nil {
		in, out := &in.ARecords, &out.ARecords
		*out = make([]ARecord, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AAAARecords != nil {
		in, out := &in.AAAARecords, &out.AAAARecords
		*out = make([]AAAARecord, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.MXRecords != nil {
		in, out := &in.MXRecords, &out.MXRecords
		*out = make([]MXRecord, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.NSRecords != nil {
		in, out := &in.NSRecords, &out.NSRecords
		*out = make([]NSRecord, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.PTRRecords != nil {
		in, out := &in.PTRRecords, &out.PTRRecords
		*out = make([]PTRRecord, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.SRVRecords != nil {
		in, out := &in.SRVRecords, &out.SRVRecords
		*out = make([]SRVRecord, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.TXTRecords != nil {
		in, out := &in.TXTRecords, &out.TXTRecords
		*out = make([]TXTRecord, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	in.CNAMERecord.DeepCopyInto(&out.CNAMERecord)
	in.SOARecord.DeepCopyInto(&out.SOARecord)
	if in.CAARecords != nil {
		in, out := &in.CAARecords, &out.CAARecords
		*out = make([]CAARecord, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecordSetParameters.
func (in *RecordSetParameters) DeepCopy() *RecordSetParameters {
	if in == nil {
		return nil
	}
	out := new(RecordSetParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecordSetSpec) DeepCopyInto(out *RecordSetSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecordSetSpec.
func (in *RecordSetSpec) DeepCopy() *RecordSetSpec {
	if in == nil {
		return nil
	}
	out := new(RecordSetSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RecordSetStatus) DeepCopyInto(out *RecordSetStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	out.AtProvider = in.AtProvider
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RecordSetStatus.
func (in *RecordSetStatus) DeepCopy() *RecordSetStatus {
	if in == nil {
		return nil
	}
	out := new(RecordSetStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SOARecord) DeepCopyInto(out *SOARecord) {
	*out = *in
	if in.Host != nil {
		in, out := &in.Host, &out.Host
		*out = new(string)
		**out = **in
	}
	if in.Email != nil {
		in, out := &in.Email, &out.Email
		*out = new(string)
		**out = **in
	}
	if in.SerialNumber != nil {
		in, out := &in.SerialNumber, &out.SerialNumber
		*out = new(int)
		**out = **in
	}
	if in.RefreshTime != nil {
		in, out := &in.RefreshTime, &out.RefreshTime
		*out = new(int)
		**out = **in
	}
	if in.RetryTime != nil {
		in, out := &in.RetryTime, &out.RetryTime
		*out = new(int)
		**out = **in
	}
	if in.ExpireTime != nil {
		in, out := &in.ExpireTime, &out.ExpireTime
		*out = new(int)
		**out = **in
	}
	if in.MinimumTTL != nil {
		in, out := &in.MinimumTTL, &out.MinimumTTL
		*out = new(int)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SOARecord.
func (in *SOARecord) DeepCopy() *SOARecord {
	if in == nil {
		return nil
	}
	out := new(SOARecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SRVRecord) DeepCopyInto(out *SRVRecord) {
	*out = *in
	if in.Priority != nil {
		in, out := &in.Priority, &out.Priority
		*out = new(int)
		**out = **in
	}
	if in.Weight != nil {
		in, out := &in.Weight, &out.Weight
		*out = new(int)
		**out = **in
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(int)
		**out = **in
	}
	if in.Target != nil {
		in, out := &in.Target, &out.Target
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SRVRecord.
func (in *SRVRecord) DeepCopy() *SRVRecord {
	if in == nil {
		return nil
	}
	out := new(SRVRecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubResource) DeepCopyInto(out *SubResource) {
	*out = *in
	if in.ID != nil {
		in, out := &in.ID, &out.ID
		*out = new(string)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubResource.
func (in *SubResource) DeepCopy() *SubResource {
	if in == nil {
		return nil
	}
	out := new(SubResource)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TXTRecord) DeepCopyInto(out *TXTRecord) {
	*out = *in
	if in.Value != nil {
		in, out := &in.Value, &out.Value
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TXTRecord.
func (in *TXTRecord) DeepCopy() *TXTRecord {
	if in == nil {
		return nil
	}
	out := new(TXTRecord)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Zone) DeepCopyInto(out *Zone) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Zone.
func (in *Zone) DeepCopy() *Zone {
	if in == nil {
		return nil
	}
	out := new(Zone)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Zone) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZoneList) DeepCopyInto(out *ZoneList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Zone, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZoneList.
func (in *ZoneList) DeepCopy() *ZoneList {
	if in == nil {
		return nil
	}
	out := new(ZoneList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ZoneList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZoneObservation) DeepCopyInto(out *ZoneObservation) {
	*out = *in
	if in.NameServers != nil {
		in, out := &in.NameServers, &out.NameServers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZoneObservation.
func (in *ZoneObservation) DeepCopy() *ZoneObservation {
	if in == nil {
		return nil
	}
	out := new(ZoneObservation)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZoneParameters) DeepCopyInto(out *ZoneParameters) {
	*out = *in
	if in.ResourceGroupNameRef != nil {
		in, out := &in.ResourceGroupNameRef, &out.ResourceGroupNameRef
		*out = new(v1.Reference)
		**out = **in
	}
	if in.ResourceGroupNameSelector != nil {
		in, out := &in.ResourceGroupNameSelector, &out.ResourceGroupNameSelector
		*out = new(v1.Selector)
		(*in).DeepCopyInto(*out)
	}
	if in.ZoneType != nil {
		in, out := &in.ZoneType, &out.ZoneType
		*out = new(string)
		**out = **in
	}
	if in.RegistrationVirtualNetworks != nil {
		in, out := &in.RegistrationVirtualNetworks, &out.RegistrationVirtualNetworks
		*out = make([]SubResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.ResolutionVirtualNetworks != nil {
		in, out := &in.ResolutionVirtualNetworks, &out.ResolutionVirtualNetworks
		*out = make([]SubResource, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Tags != nil {
		in, out := &in.Tags, &out.Tags
		*out = make(map[string]*string, len(*in))
		for key, val := range *in {
			var outVal *string
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = new(string)
				**out = **in
			}
			(*out)[key] = outVal
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZoneParameters.
func (in *ZoneParameters) DeepCopy() *ZoneParameters {
	if in == nil {
		return nil
	}
	out := new(ZoneParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZoneSpec) DeepCopyInto(out *ZoneSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.ForProvider.DeepCopyInto(&out.ForProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZoneSpec.
func (in *ZoneSpec) DeepCopy() *ZoneSpec {
	if in == nil {
		return nil
	}
	out := new(ZoneSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ZoneStatus) DeepCopyInto(out *ZoneStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
	in.AtProvider.DeepCopyInto(&out.AtProvider)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ZoneStatus.
func (in *ZoneStatus) DeepCopy() *ZoneStatus {
	if in == nil {
		return nil
	}
	out := new(ZoneStatus)
	in.DeepCopyInto(out)
	return out
}
