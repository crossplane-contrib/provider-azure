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

package v1alpha3

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

const (
	// DefaultNodeCount is the default node count for a cluster.
	DefaultNodeCount = 1
)

// AKSClusterParameters define the desired state of an Azure Kubernetes Engine
// cluster.
type AKSClusterParameters struct {
	// ResourceGroupName is the name of the resource group that the cluster will
	// be created in
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup to retrieve its
	// name
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to a ResourceGroup to
	// retrieve its name
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	// Location is the Azure location that the cluster will be created in
	Location string `json:"location"`

	// Version is the Kubernetes version that will be deployed to the cluster
	Version string `json:"version"`

	// VnetSubnetID is the subnet to which the cluster will be deployed.
	// +optional
	VnetSubnetID string `json:"vnetSubnetID,omitempty"`

	// ResourceGroupNameRef - A reference to a Subnet to retrieve its ID
	VnetSubnetIDRef *xpv1.Reference `json:"vnetSubnetIDRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to a Subnet to retrieve
	// its ID
	VnetSubnetIDSelector *xpv1.Selector `json:"vnetSubnetIDSelector,omitempty"`

	// NodeCount is the number of nodes that the cluster will initially be
	// created with.  This can be scaled over time and defaults to 1.
	// +kubebuilder:validation:Maximum=100
	// +kubebuilder:validation:Minimum=0
	// +optional
	NodeCount *int `json:"nodeCount,omitempty"`

	// NodeVMSize is the name of the worker node VM size, e.g., Standard_B2s,
	// Standard_F2s_v2, etc.
	// +optional
	NodeVMSize string `json:"nodeVMSize"`

	// DNSNamePrefix is the DNS name prefix to use with the hosted Kubernetes
	// API server FQDN. You will use this to connect to the Kubernetes API when
	// managing containers after creating the cluster.
	// +optional
	DNSNamePrefix string `json:"dnsNamePrefix"`

	// DisableRBAC determines whether RBAC will be disabled or enabled in the
	// cluster.
	// +optional
	DisableRBAC bool `json:"disableRBAC,omitempty"`
}

// An AKSClusterSpec defines the desired state of a AKSCluster.
type AKSClusterSpec struct {
	xpv1.ResourceSpec    `json:",inline"`
	AKSClusterParameters `json:",inline"`
}

// An AKSClusterStatus represents the observed state of an AKSCluster.
type AKSClusterStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// State is the current state of the cluster.
	State string `json:"state,omitempty"`

	// ProviderID is the external ID to identify this resource in the cloud
	// provider.
	ProviderID string `json:"providerID,omitempty"`

	// Endpoint is the endpoint where the cluster can be reached
	Endpoint string `json:"endpoint,omitempty"`
}

// +kubebuilder:object:root=true

// An AKSCluster is a managed resource that represents an Azure Kubernetes
// Engine cluster.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="ENDPOINT",type="string",JSONPath=".status.endpoint"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.location"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
// +kubebuilder:subresource:status
type AKSCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AKSClusterSpec   `json:"spec"`
	Status AKSClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AKSClusterList contains a list of AKSCluster.
type AKSClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AKSCluster `json:"items"`
}

// ImageReferenceParameters specifies information about the image to use. You can specify information about platform
// images, marketplace images, or virtual machine images. This element is required when you want to use a
// platform image, marketplace image, or virtual machine image, but is not used in other creation
// operations. NOTE: Image reference publisher and offer can only be set when you create the scale set.
type ImageReferenceParameters struct {
	// Publisher - The image publisher.
	Publisher string `json:"publisher"`
	// Offer - Specifies the offer of the platform image or marketplace image used to create the virtual machine.
	Offer string `json:"offer"`
	// Sku - The image SKU.
	Sku string `json:"sku"`
	// Version - Specifies the version of the platform image or marketplace image used to create the virtual machine. The allowed formats are Major.Minor.Build or 'latest'. Major, Minor, and Build are decimal numbers. Specify 'latest' to use the latest version of an image available at deploy time. Even if you use 'latest', the VM image will not automatically update after deploy time even if a new version becomes available.
	Version string `json:"version"`
}

// OSDiskParameters specifies information about the operating system disk used by the virtual machine. <br><br> For
// more information about disks, see [About disks and VHDs for Azure virtual
// machines](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-windows-about-disks-vhds?toc=%2fazure%2fvirtual-machines%2fwindows%2ftoc.json).
type OSDiskParameters struct {
	// Name - The disk name.
	Name string `json:"name"`
	// Vhd - The virtual hard disk.
	Vhd *VirtualHardDiskParameters `json:"vhd"`
}

// VirtualHardDiskParameters describes the uri of a disk.
type VirtualHardDiskParameters struct {
	// URI - Specifies the virtual hard disk's uri.
	URI string `json:"uri,omitempty"`
}

// StorageProfileParameters specifies the storage settings for the virtual machine disks.
type StorageProfileParameters struct {
	// ImageReference - Specifies information about the image to use. You can specify information about platform images, marketplace images, or virtual machine images. This element is required when you want to use a platform image, marketplace image, or virtual machine image, but is not used in other creation operations.
	ImageReference *ImageReferenceParameters `json:"imageReference"`
	// OsDisk - Specifies information about the operating system disk used by the virtual machine. <br><br> For more information about disks, see [About disks and VHDs for Azure virtual machines](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-windows-about-disks-vhds?toc=%2fazure%2fvirtual-machines%2fwindows%2ftoc.json).
	OsDisk *OSDiskParameters `json:"osDisk,omitempty"`
}

// OSProfileParameters specifies the operating system settings for the virtual machine. Some of the settings cannot
// be changed once VM is provisioned.
type OSProfileParameters struct {
	// ComputerName - Specifies the host OS name of the virtual machine. <br><br> This name cannot be updated after the VM is created. <br><br> **Max-length (Windows):** 15 characters <br><br> **Max-length (Linux):** 64 characters. <br><br> For naming conventions and restrictions see [Azure infrastructure services implementation guidelines](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-infrastructure-subscription-accounts-guidelines?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json#1-naming-conventions).
	ComputerName string `json:"computerName"`
	// AdminUsername - Specifies the name of the administrator account. <br><br> This property cannot be updated after the VM is created. <br><br> **Windows-only restriction:** Cannot end in "." <br><br> **Disallowed values:** "administrator", "admin", "user", "user1", "test", "user2", "test1", "user3", "admin1", "1", "123", "a", "actuser", "adm", "admin2", "aspnet", "backup", "console", "david", "guest", "john", "owner", "root", "server", "sql", "support", "support_388945a0", "sys", "test2", "test3", "user4", "user5". <br><br> **Minimum-length (Linux):** 1  character <br><br> **Max-length (Linux):** 64 characters <br><br> **Max-length (Windows):** 20 characters  <br><br><li> For root access to the Linux VM, see [Using root privileges on Linux virtual machines in Azure](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-use-root-privileges?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json)<br><li> For a list of built-in system users on Linux that should not be used in this field, see [Selecting User Names for Linux on Azure](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-usernames?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json)
	AdminUsername string `json:"adminUsername"`
	// AdminPassword - Specifies the password of the administrator account. <br><br> **Minimum-length (Windows):** 8 characters <br><br> **Minimum-length (Linux):** 6 characters <br><br> **Max-length (Windows):** 123 characters <br><br> **Max-length (Linux):** 72 characters <br><br> **Complexity requirements:** 3 out of 4 conditions below need to be fulfilled <br> Has lower characters <br>Has upper characters <br> Has a digit <br> Has a special character (Regex match [\W_]) <br><br> **Disallowed values:** "abc@123", "P@$$w0rd", "P@ssw0rd", "P@ssword123", "Pa$$word", "pass@word1", "Password!", "Password1", "Password22", "iloveyou!" <br><br> For resetting the password, see [How to reset the Remote Desktop service or its login password in a Windows VM](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-windows-reset-rdp?toc=%2fazure%2fvirtual-machines%2fwindows%2ftoc.json) <br><br> For resetting root password, see [Manage users, SSH, and check or repair disks on Azure Linux VMs using the VMAccess Extension](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-using-vmaccess-extension?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json#reset-root-password)
	AdminPassword string `json:"adminPassword"`

	// LinuxConfiguration - Specifies the Linux operating system settings on the virtual machine. <br><br>For a list of supported Linux distributions, see [Linux on Azure-Endorsed Distributions](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-endorsed-distros?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json) <br><br> For running non-endorsed distributions, see [Information for Non-Endorsed Distributions](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-create-upload-generic?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json).
	LinuxConfiguration *LinuxConfigurationParameters `json:"linuxConfiguration"`
}

// LinuxConfigurationParameters specifies the Linux operating system settings on the virtual machine. <br><br>For a
// list of supported Linux distributions, see [Linux on Azure-Endorsed
// Distributions](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-endorsed-distros?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json)
// <br><br> For running non-endorsed distributions, see [Information for Non-Endorsed
// Distributions](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-create-upload-generic?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json).
type LinuxConfigurationParameters struct {
	// SSH - Specifies the ssh key configuration for a Linux OS.
	SSH *SSHConfigurationParameters `json:"ssh"`
}

// SSHConfigurationParameters SSH configuration for Linux based VMs running on Azure
type SSHConfigurationParameters struct {
	// PublicKeys - The list of SSH public keys used to authenticate with linux based VMs.
	PublicKeys []*SSHPublicKey `json:"publicKeys,omitempty"`
}

// SSHPublicKey contains information about SSH certificate public key and the path on the Linux VM where
// the public key is placed.
type SSHPublicKey struct {
	// Path - Specifies the full path on the created VM where ssh public key is stored. If the file already exists, the specified key is appended to the file. Example: /home/user/.ssh/authorized_keys
	Path string `json:"path"`
	// KeyData - SSH public key certificate used to authenticate with the VM through ssh. The key needs to be at least 2048-bit and in ssh-rsa format. <br><br> For creating ssh keys, see [Create SSH keys on Linux and Mac for Linux VMs in Azure](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-linux-mac-create-ssh-keys?toc=%2fazure%2fvirtual-machines%2flinux%2ftoc.json).
	KeyData string `json:"keyData"`
}

// HardwareProfileParameters specifies the hardware settings for the virtual machine.
type HardwareProfileParameters struct {
	// VMSize - Specifies the size of the virtual machine. For more information about virtual machine sizes, see [Sizes for virtual machines](https://docs.microsoft.com/azure/virtual-machines/virtual-machines-windows-sizes?toc=%2fazure%2fvirtual-machines%2fwindows%2ftoc.json). <br><br> The available VM sizes depend on region and availability set. For a list of available sizes use these APIs:  <br><br> [List all available virtual machine sizes in an availability set](https://docs.microsoft.com/rest/api/compute/availabilitysets/listavailablesizes) <br><br> [List all available virtual machine sizes in a region](https://docs.microsoft.com/rest/api/compute/virtualmachinesizes/list) <br><br> [List all available virtual machine sizes for resizing](https://docs.microsoft.com/rest/api/compute/virtualmachines/listavailablesizes). Possible values include: 'VirtualMachineSizeTypesBasicA0', 'VirtualMachineSizeTypesBasicA1', 'VirtualMachineSizeTypesBasicA2', 'VirtualMachineSizeTypesBasicA3', 'VirtualMachineSizeTypesBasicA4', 'VirtualMachineSizeTypesStandardA0', 'VirtualMachineSizeTypesStandardA1', 'VirtualMachineSizeTypesStandardA2', 'VirtualMachineSizeTypesStandardA3', 'VirtualMachineSizeTypesStandardA4', 'VirtualMachineSizeTypesStandardA5', 'VirtualMachineSizeTypesStandardA6', 'VirtualMachineSizeTypesStandardA7', 'VirtualMachineSizeTypesStandardA8', 'VirtualMachineSizeTypesStandardA9', 'VirtualMachineSizeTypesStandardA10', 'VirtualMachineSizeTypesStandardA11', 'VirtualMachineSizeTypesStandardA1V2', 'VirtualMachineSizeTypesStandardA2V2', 'VirtualMachineSizeTypesStandardA4V2', 'VirtualMachineSizeTypesStandardA8V2', 'VirtualMachineSizeTypesStandardA2mV2', 'VirtualMachineSizeTypesStandardA4mV2', 'VirtualMachineSizeTypesStandardA8mV2', 'VirtualMachineSizeTypesStandardB1s', 'VirtualMachineSizeTypesStandardB1ms', 'VirtualMachineSizeTypesStandardB2s', 'VirtualMachineSizeTypesStandardB2ms', 'VirtualMachineSizeTypesStandardB4ms', 'VirtualMachineSizeTypesStandardB8ms', 'VirtualMachineSizeTypesStandardD1', 'VirtualMachineSizeTypesStandardD2', 'VirtualMachineSizeTypesStandardD3', 'VirtualMachineSizeTypesStandardD4', 'VirtualMachineSizeTypesStandardD11', 'VirtualMachineSizeTypesStandardD12', 'VirtualMachineSizeTypesStandardD13', 'VirtualMachineSizeTypesStandardD14', 'VirtualMachineSizeTypesStandardD1V2', 'VirtualMachineSizeTypesStandardD2V2', 'VirtualMachineSizeTypesStandardD3V2', 'VirtualMachineSizeTypesStandardD4V2', 'VirtualMachineSizeTypesStandardD5V2', 'VirtualMachineSizeTypesStandardD2V3', 'VirtualMachineSizeTypesStandardD4V3', 'VirtualMachineSizeTypesStandardD8V3', 'VirtualMachineSizeTypesStandardD16V3', 'VirtualMachineSizeTypesStandardD32V3', 'VirtualMachineSizeTypesStandardD64V3', 'VirtualMachineSizeTypesStandardD2sV3', 'VirtualMachineSizeTypesStandardD4sV3', 'VirtualMachineSizeTypesStandardD8sV3', 'VirtualMachineSizeTypesStandardD16sV3', 'VirtualMachineSizeTypesStandardD32sV3', 'VirtualMachineSizeTypesStandardD64sV3', 'VirtualMachineSizeTypesStandardD11V2', 'VirtualMachineSizeTypesStandardD12V2', 'VirtualMachineSizeTypesStandardD13V2', 'VirtualMachineSizeTypesStandardD14V2', 'VirtualMachineSizeTypesStandardD15V2', 'VirtualMachineSizeTypesStandardDS1', 'VirtualMachineSizeTypesStandardDS2', 'VirtualMachineSizeTypesStandardDS3', 'VirtualMachineSizeTypesStandardDS4', 'VirtualMachineSizeTypesStandardDS11', 'VirtualMachineSizeTypesStandardDS12', 'VirtualMachineSizeTypesStandardDS13', 'VirtualMachineSizeTypesStandardDS14', 'VirtualMachineSizeTypesStandardDS1V2', 'VirtualMachineSizeTypesStandardDS2V2', 'VirtualMachineSizeTypesStandardDS3V2', 'VirtualMachineSizeTypesStandardDS4V2', 'VirtualMachineSizeTypesStandardDS5V2', 'VirtualMachineSizeTypesStandardDS11V2', 'VirtualMachineSizeTypesStandardDS12V2', 'VirtualMachineSizeTypesStandardDS13V2', 'VirtualMachineSizeTypesStandardDS14V2', 'VirtualMachineSizeTypesStandardDS15V2', 'VirtualMachineSizeTypesStandardDS134V2', 'VirtualMachineSizeTypesStandardDS132V2', 'VirtualMachineSizeTypesStandardDS148V2', 'VirtualMachineSizeTypesStandardDS144V2', 'VirtualMachineSizeTypesStandardE2V3', 'VirtualMachineSizeTypesStandardE4V3', 'VirtualMachineSizeTypesStandardE8V3', 'VirtualMachineSizeTypesStandardE16V3', 'VirtualMachineSizeTypesStandardE32V3', 'VirtualMachineSizeTypesStandardE64V3', 'VirtualMachineSizeTypesStandardE2sV3', 'VirtualMachineSizeTypesStandardE4sV3', 'VirtualMachineSizeTypesStandardE8sV3', 'VirtualMachineSizeTypesStandardE16sV3', 'VirtualMachineSizeTypesStandardE32sV3', 'VirtualMachineSizeTypesStandardE64sV3', 'VirtualMachineSizeTypesStandardE3216V3', 'VirtualMachineSizeTypesStandardE328sV3', 'VirtualMachineSizeTypesStandardE6432sV3', 'VirtualMachineSizeTypesStandardE6416sV3', 'VirtualMachineSizeTypesStandardF1', 'VirtualMachineSizeTypesStandardF2', 'VirtualMachineSizeTypesStandardF4', 'VirtualMachineSizeTypesStandardF8', 'VirtualMachineSizeTypesStandardF16', 'VirtualMachineSizeTypesStandardF1s', 'VirtualMachineSizeTypesStandardF2s', 'VirtualMachineSizeTypesStandardF4s', 'VirtualMachineSizeTypesStandardF8s', 'VirtualMachineSizeTypesStandardF16s', 'VirtualMachineSizeTypesStandardF2sV2', 'VirtualMachineSizeTypesStandardF4sV2', 'VirtualMachineSizeTypesStandardF8sV2', 'VirtualMachineSizeTypesStandardF16sV2', 'VirtualMachineSizeTypesStandardF32sV2', 'VirtualMachineSizeTypesStandardF64sV2', 'VirtualMachineSizeTypesStandardF72sV2', 'VirtualMachineSizeTypesStandardG1', 'VirtualMachineSizeTypesStandardG2', 'VirtualMachineSizeTypesStandardG3', 'VirtualMachineSizeTypesStandardG4', 'VirtualMachineSizeTypesStandardG5', 'VirtualMachineSizeTypesStandardGS1', 'VirtualMachineSizeTypesStandardGS2', 'VirtualMachineSizeTypesStandardGS3', 'VirtualMachineSizeTypesStandardGS4', 'VirtualMachineSizeTypesStandardGS5', 'VirtualMachineSizeTypesStandardGS48', 'VirtualMachineSizeTypesStandardGS44', 'VirtualMachineSizeTypesStandardGS516', 'VirtualMachineSizeTypesStandardGS58', 'VirtualMachineSizeTypesStandardH8', 'VirtualMachineSizeTypesStandardH16', 'VirtualMachineSizeTypesStandardH8m', 'VirtualMachineSizeTypesStandardH16m', 'VirtualMachineSizeTypesStandardH16r', 'VirtualMachineSizeTypesStandardH16mr', 'VirtualMachineSizeTypesStandardL4s', 'VirtualMachineSizeTypesStandardL8s', 'VirtualMachineSizeTypesStandardL16s', 'VirtualMachineSizeTypesStandardL32s', 'VirtualMachineSizeTypesStandardM64s', 'VirtualMachineSizeTypesStandardM64ms', 'VirtualMachineSizeTypesStandardM128s', 'VirtualMachineSizeTypesStandardM128ms', 'VirtualMachineSizeTypesStandardM6432ms', 'VirtualMachineSizeTypesStandardM6416ms', 'VirtualMachineSizeTypesStandardM12864ms', 'VirtualMachineSizeTypesStandardM12832ms', 'VirtualMachineSizeTypesStandardNC6', 'VirtualMachineSizeTypesStandardNC12', 'VirtualMachineSizeTypesStandardNC24', 'VirtualMachineSizeTypesStandardNC24r', 'VirtualMachineSizeTypesStandardNC6sV2', 'VirtualMachineSizeTypesStandardNC12sV2', 'VirtualMachineSizeTypesStandardNC24sV2', 'VirtualMachineSizeTypesStandardNC24rsV2', 'VirtualMachineSizeTypesStandardNC6sV3', 'VirtualMachineSizeTypesStandardNC12sV3', 'VirtualMachineSizeTypesStandardNC24sV3', 'VirtualMachineSizeTypesStandardNC24rsV3', 'VirtualMachineSizeTypesStandardND6s', 'VirtualMachineSizeTypesStandardND12s', 'VirtualMachineSizeTypesStandardND24s', 'VirtualMachineSizeTypesStandardND24rs', 'VirtualMachineSizeTypesStandardNV6', 'VirtualMachineSizeTypesStandardNV12', 'VirtualMachineSizeTypesStandardNV24'
	VMSize string `json:"vmSize"`
}

// NetworkProfileParameters specifies the network interfaces of the virtual machine.
type NetworkProfileParameters struct {
	// NetworkInterfaces - Specifies the list of resource Ids for the network interfaces associated with the virtual machine.
	NetworkInterfaces []*NetworkInterfaceReferenceParameters `json:"networkInterfaces"`
}

// NetworkInterfaceReferenceParameters describes a network interface reference.
type NetworkInterfaceReferenceParameters struct {
	// NetworkInterfaceID is the id of network interface that the cluster will be created in
	NetworkInterfaceID string `json:"networkInterfaceID,omitempty"`

	// NetworkInterfaceIDRef - A reference to a NetworkInterface to retrieve its id
	NetworkInterfaceIDRef *xpv1.Reference `json:"networkInterfaceIDRef,omitempty"`

	// NetworkInterfaceIDSelector - Select a reference to a NetworkInterface to retrieve its id
	NetworkInterfaceIDSelector *xpv1.Selector `json:"networkInterfaceIDSelector,omitempty"`

	// Primary - Specifies the primary network interface in case the virtual machine has more than 1 network interface.
	Primary bool `json:"primary,omitempty"`
}

// VirtualMachineParameters describes a Virtual Machine.
type VirtualMachineParameters struct {
	// Location - Resource location
	Location string `json:"location"`
	// HardwareProfile - Specifies the hardware settings for the virtual machine.
	HardwareProfile *HardwareProfileParameters `json:"hardwareProfile"`
	// StorageProfile - Specifies the storage settings for the virtual machine disks.
	StorageProfile *StorageProfileParameters `json:"storageProfile"`
	// OsProfile - Specifies the operating system settings used while creating the virtual machine. Some of the settings cannot be changed once VM is provisioned.
	OsProfile *OSProfileParameters `json:"osProfile"`
	// NetworkProfile - Specifies the network interfaces of the virtual machine.
	NetworkProfile *NetworkProfileParameters `json:"networkProfile"`
}

// An VirtualMachineSpec defines the desired state of a VirtualMachine.
type VirtualMachineSpec struct {
	xpv1.ResourceSpec `json:",inline"`

	// ResourceGroupName is the name of the resource group that the cluster will
	// be created in
	ResourceGroupName string `json:"resourceGroupName,omitempty"`

	// ResourceGroupNameRef - A reference to a ResourceGroup to retrieve its
	// name
	ResourceGroupNameRef *xpv1.Reference `json:"resourceGroupNameRef,omitempty"`

	// ResourceGroupNameSelector - Select a reference to a ResourceGroup to
	// retrieve its name
	ResourceGroupNameSelector *xpv1.Selector `json:"resourceGroupNameSelector,omitempty"`

	VirtualMachineParameters *VirtualMachineParameters `json:"properties"`
}

// An PrimaryInterfaceStatus represents the observed state of primary interface of VirtualMachine.
type PrimaryInterfaceStatus struct {
	Subnet          string `json:"subnet"`
	PublicIPAddress string `json:"publicIPAddress"`
}

// An VirtualMachineStatus represents the observed state of an VirtualMachine.
type VirtualMachineStatus struct {
	xpv1.ResourceStatus `json:",inline"`

	// State is the current state of the cluster.
	State string `json:"state,omitempty"`

	// OSDiskName - The disk name.
	OSDiskName string `json:"OSDiskName,omitempty"`
}

// +kubebuilder:object:root=true

// An VirtualMachine is a managed resource that represents an VirtualMachine
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="LOCATION",type="string",JSONPath=".spec.properties.location"
// +kubebuilder:printcolumn:name="SIZE",type="string",JSONPath=".spec.properties.hardwareProfile.vmSize"
// +kubebuilder:printcolumn:name="OS_OFFER",type="string",JSONPath=".spec.properties.storageProfile.imageReference.offer"
// +kubebuilder:printcolumn:name="OS_SKU",type="string",JSONPath=".spec.properties.storageProfile.imageReference.sku"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,azure}
// +kubebuilder:subresource:status
type VirtualMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VirtualMachineSpec   `json:"spec"`
	Status VirtualMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// VirtualMachineList contains a list of VirtualMachine.
type VirtualMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VirtualMachine `json:"items"`
}
