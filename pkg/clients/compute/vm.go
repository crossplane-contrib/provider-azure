/*
Copyright 2021 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the c.Specific language governing permissions and
limitations under the License.
*/

package compute

import (
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

// NewStorageProfileParameters converts to Azure StorageProfile
func NewStorageProfileParameters(profile *v1alpha3.StorageProfileParameters) *compute.StorageProfile {
	return &compute.StorageProfile{
		ImageReference: NewImageParameters(profile.ImageReference),
		OsDisk:         NewOSDiskParameters(profile.OsDisk),
	}
}

// NewImageParameters converts to Azure ImageReference
func NewImageParameters(img *v1alpha3.ImageReferenceParameters) *compute.ImageReference {
	return &compute.ImageReference{
		Publisher: azure.ToStringPtr(img.Publisher),
		Offer:     azure.ToStringPtr(img.Offer),
		Sku:       azure.ToStringPtr(img.Sku),
		Version:   azure.ToStringPtr(img.Version),
	}
}

// NewOSDiskParameters converts to Azure OSDisk
func NewOSDiskParameters(img *v1alpha3.OSDiskParameters) *compute.OSDisk {
	if img == nil {
		return nil
	}
	return &compute.OSDisk{
		Name:         azure.ToStringPtr(img.Name),
		Vhd:          &compute.VirtualHardDisk{URI: azure.ToStringPtr(img.Vhd.URI)},
		CreateOption: compute.DiskCreateOptionTypesFromImage,
	}
}

// NewOSProfileParameters converts to Azure OSProfile
func NewOSProfileParameters(os *v1alpha3.OSProfileParameters) *compute.OSProfile {
	return &compute.OSProfile{
		ComputerName:  to.StringPtr(os.ComputerName),
		AdminUsername: to.StringPtr(os.AdminUsername),
		AdminPassword: to.StringPtr(os.AdminPassword),
		LinuxConfiguration: &compute.LinuxConfiguration{
			SSH: &compute.SSHConfiguration{
				PublicKeys: NewSSHPublicKeys(os.LinuxConfiguration.SSH.PublicKeys),
			},
		},
	}
}

// NewSSHPublicKeys converts to Azure SSHPublicKey
func NewSSHPublicKeys(keys []*v1alpha3.SSHPublicKey) *[]compute.SSHPublicKey {
	v := make([]compute.SSHPublicKey, len(keys))
	for i, key := range keys {
		v[i] = compute.SSHPublicKey{
			Path:    azure.ToStringPtr(key.Path),
			KeyData: azure.ToStringPtr(key.KeyData),
		}
	}
	return &v
}

// NewHardwareProfileParameters converts to Azure HardwareProfile
func NewHardwareProfileParameters(hwprof *v1alpha3.HardwareProfileParameters) *compute.HardwareProfile {
	return &compute.HardwareProfile{
		VMSize: compute.VirtualMachineSizeTypes(hwprof.VMSize),
	}
}

// NewNetworkProfileParameters converts to Azure NetworkProfile
func NewNetworkProfileParameters(netprof *v1alpha3.NetworkProfileParameters) *compute.NetworkProfile {
	return &compute.NetworkProfile{
		NetworkInterfaces: NewNetworkInterfaces(netprof.NetworkInterfaces),
	}
}

// NewNetworkInterfaces converts to Azure NetworkInterfaceReference
func NewNetworkInterfaces(ifaces []*v1alpha3.NetworkInterfaceReferenceParameters) *[]compute.NetworkInterfaceReference {
	ifs := make([]compute.NetworkInterfaceReference, len(ifaces))
	for i, iface := range ifaces {
		ifs[i] = compute.NetworkInterfaceReference{
			ID: azure.ToStringPtr(iface.NetworkInterfaceID),
			NetworkInterfaceReferenceProperties: &compute.NetworkInterfaceReferenceProperties{
				Primary: azure.ToBoolPtr(iface.Primary),
			},
		}
	}
	return &ifs
}

// NewVirtualMachine converts to Azure VirtualMachine
func NewVirtualMachine(vmparams *v1alpha3.VirtualMachineParameters) compute.VirtualMachine {
	return compute.VirtualMachine{
		Location: to.StringPtr(vmparams.Location),
		VirtualMachineProperties: &compute.VirtualMachineProperties{
			HardwareProfile: NewHardwareProfileParameters(vmparams.HardwareProfile),
			StorageProfile:  NewStorageProfileParameters(vmparams.StorageProfile),
			OsProfile:       NewOSProfileParameters(vmparams.OsProfile),
			NetworkProfile:  NewNetworkProfileParameters(vmparams.NetworkProfile),
		},
	}
}

// UpdateVirtualMachineStatus updates the status related to the external
// Azure network interface in the VirtualMachine
func UpdateVirtualMachineStatus(v *v1alpha3.VirtualMachine, az *compute.VirtualMachine) {
	v.Status.State = azure.ToString(az.ProvisioningState)
	v.Status.OSDiskName = azure.ToString(az.StorageProfile.OsDisk.Name)
}

// UpdateDiskStatus updates the status related to the external
// Azure disk in the VirtualMachine
func UpdateDiskStatus(v *v1alpha3.VirtualMachine, az *compute.Disk) {
	v.Status.OSDiskName = azure.ToString(az.Name)
}
