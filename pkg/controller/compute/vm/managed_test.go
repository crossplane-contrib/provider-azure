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

package vm

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/go-autorest/autorest"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/compute/fake"
)

const (
	name              = "coolVirtualMachine"
	uid               = types.UID("definitely-a-uuid")
	resourceGroupName = "coolRG"
)

var (
	ctx       = context.Background()
	errorBoom = errors.New("boom")
)

type testCase struct {
	name    string
	e       managed.ExternalClient
	r       resource.Managed
	want    resource.Managed
	wantErr error
}

type virtualMachineModifier func(address *v1alpha3.VirtualMachine)

func withConditions(c ...xpv1.Condition) virtualMachineModifier {
	return func(r *v1alpha3.VirtualMachine) { r.Status.ConditionedStatus.Conditions = c }
}

func withState(s string) virtualMachineModifier {
	return func(r *v1alpha3.VirtualMachine) { r.Status.State = s }
}

func virtualMachine(sm ...virtualMachineModifier) *v1alpha3.VirtualMachine {
	r := &v1alpha3.VirtualMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.VirtualMachineSpec{
			ResourceGroupName: resourceGroupName,
			VirtualMachineParameters: &v1alpha3.VirtualMachineParameters{
				Location: "West US 2",
				HardwareProfile: &v1alpha3.HardwareProfileParameters{
					VMSize: "Standard_B1s",
				},
				StorageProfile: &v1alpha3.StorageProfileParameters{
					ImageReference: &v1alpha3.ImageReferenceParameters{
						Publisher: "Canonical",
						Offer:     "UbuntuServer",
						Sku:       "16.04-LTS",
						Version:   "latest",
					},
				},
				OsProfile: &v1alpha3.OSProfileParameters{
					ComputerName:  "example-vm",
					AdminUsername: "testuser",
					AdminPassword: "t2st-uSer",
					LinuxConfiguration: &v1alpha3.LinuxConfigurationParameters{
						SSH: &v1alpha3.SSHConfigurationParameters{
							PublicKeys: make([]*v1alpha3.SSHPublicKey, 0),
						},
					},
				},
				NetworkProfile: &v1alpha3.NetworkProfileParameters{
					NetworkInterfaces: make([]*v1alpha3.NetworkInterfaceReferenceParameters, 0),
				},
			},
		},
		Status: v1alpha3.VirtualMachineStatus{OSDiskName: "test-name"},
	}
	meta.SetExternalName(r, name)
	for _, m := range sm {
		m(r)
	}
	return r
}

// Test that our Reconciler implementation satisfies the Reconciler interface.
var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connecter{}

func TestCreate(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotVirtualMachine",
			e:       &external{client: &fake.VirtualMachineClient{}},
			r:       &v1alpha3.AKSCluster{},
			want:    &v1alpha3.AKSCluster{},
			wantErr: errors.New(errNotVirtualMachine),
		},
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.VirtualMachineClient{
				MockCreateOrUpdate: func(ctx context.Context, resourceGroupName string, VMName string, parameters compute.VirtualMachine) (result compute.VirtualMachinesCreateOrUpdateFuture, err error) {
					return compute.VirtualMachinesCreateOrUpdateFuture{}, nil
				},
			}},
			r: virtualMachine(),
			want: virtualMachine(
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.VirtualMachineClient{
				MockCreateOrUpdate: func(ctx context.Context, resourceGroupName string, VMName string, parameters compute.VirtualMachine) (result compute.VirtualMachinesCreateOrUpdateFuture, err error) {
					return compute.VirtualMachinesCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r: virtualMachine(),
			want: virtualMachine(
				withConditions(xpv1.Creating()),
			),
			wantErr: errors.Wrap(errorBoom, errCreateVirtualMachine),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Create(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Create(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotVirtualMachine",
			e:       &external{client: &fake.VirtualMachineClient{}},
			r:       &v1alpha3.AKSCluster{},
			want:    &v1alpha3.AKSCluster{},
			wantErr: errors.New(errNotVirtualMachine),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{client: &fake.VirtualMachineClient{
				MockGet: func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
					return compute.VirtualMachine{VirtualMachineProperties: &compute.VirtualMachineProperties{}}, autorest.DetailedError{StatusCode: http.StatusNotFound}
				},
			}, diskClient: &fake.DiskClient{
				MockGet: func(ctx context.Context, resourceGroupName string, diskName string) (result compute.Disk, err error) {
					return compute.Disk{DiskProperties: &compute.DiskProperties{}}, autorest.DetailedError{StatusCode: http.StatusNotFound}
				},
			}},
			r:    virtualMachine(),
			want: virtualMachine(),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{client: &fake.VirtualMachineClient{
				MockGet: func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
					return compute.VirtualMachine{
						VirtualMachineProperties: &compute.VirtualMachineProperties{
							ProvisioningState: azure.ToStringPtr(string(compute.ProvisioningStateSucceeded)),
							StorageProfile: &compute.StorageProfile{
								OsDisk: &compute.OSDisk{
									Name: azure.ToStringPtr("test-name"),
								},
							}},
					}, nil
				},
			}},
			r: virtualMachine(),
			want: virtualMachine(
				withConditions(xpv1.Available()),
				withState(string(compute.ProvisioningStateSucceeded)),
			),
		},
		{
			name: "FailedObserve",
			e: &external{client: &fake.VirtualMachineClient{
				MockGet: func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
					return compute.VirtualMachine{}, errorBoom
				},
			}},
			r:       virtualMachine(),
			want:    virtualMachine(),
			wantErr: errors.Wrap(errorBoom, errGetVirtualMachine),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Observe(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotVirtualMachine",
			e:       &external{client: &fake.VirtualMachineClient{}},
			r:       &v1alpha3.AKSCluster{},
			want:    &v1alpha3.AKSCluster{},
			wantErr: errors.New(errNotVirtualMachine),
		},
		{
			name: "SuccessfulDoesNotNeedUpdate",
			e: &external{client: &fake.VirtualMachineClient{
				MockGet: func(ctx context.Context, resourceGroupName string, VMName string, expand compute.InstanceViewTypes) (result compute.VirtualMachine, err error) {
					return compute.VirtualMachine{VirtualMachineProperties: &compute.VirtualMachineProperties{}}, autorest.DetailedError{StatusCode: http.StatusNotFound}
				},
			}},
			r:    virtualMachine(),
			want: virtualMachine(),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Update(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Update(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotVirtualMachine",
			e:       &external{client: &fake.VirtualMachineClient{}},
			r:       &v1alpha3.AKSCluster{},
			want:    &v1alpha3.AKSCluster{},
			wantErr: errors.New(errNotVirtualMachine),
		},
		{
			name: "Successful",
			e: &external{client: &fake.VirtualMachineClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, VMName string) (result compute.VirtualMachinesDeleteFuture, err error) {
					return compute.VirtualMachinesDeleteFuture{}, nil
				},
			}, diskClient: &fake.DiskClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, diskName string) (result compute.DisksDeleteFuture, err error) {
					return compute.DisksDeleteFuture{}, nil
				},
			}},
			r: virtualMachine(),
			want: virtualMachine(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{client: &fake.VirtualMachineClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, VMName string) (result compute.VirtualMachinesDeleteFuture, err error) {
					return compute.VirtualMachinesDeleteFuture{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
				},
			}},
			r: virtualMachine(),
			want: virtualMachine(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "Failed",
			e: &external{client: &fake.VirtualMachineClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, VMName string) (result compute.VirtualMachinesDeleteFuture, err error) {
					return compute.VirtualMachinesDeleteFuture{}, errorBoom
				},
			}},
			r: virtualMachine(),
			want: virtualMachine(
				withConditions(xpv1.Deleting()),
			),
			wantErr: errors.Wrap(errorBoom, errDeleteVirtualMachine),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.e.Delete(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
