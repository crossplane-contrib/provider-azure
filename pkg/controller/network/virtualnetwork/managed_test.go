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

package virtualnetwork

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/network/fake"
)

const (
	name              = "coolSubnet"
	uid               = types.UID("definitely-a-uuid")
	addressPrefix     = "10.0.0.0/16"
	resourceGroupName = "coolRG"
	location          = "coolplace"
)

var (
	ctx       = context.Background()
	errorBoom = errors.New("boom")
	tags      = map[string]string{"one": "test", "two": "test"}
)

type testCase struct {
	name    string
	e       managed.ExternalClient
	r       resource.Managed
	want    resource.Managed
	wantErr error
}

type virtualNetworkModifier func(*v1alpha3.VirtualNetwork)

func withConditions(c ...xpv1.Condition) virtualNetworkModifier {
	return func(r *v1alpha3.VirtualNetwork) { r.Status.ConditionedStatus.Conditions = c }
}

func withState(s string) virtualNetworkModifier {
	return func(r *v1alpha3.VirtualNetwork) { r.Status.State = s }
}

func virtualNetwork(vm ...virtualNetworkModifier) *v1alpha3.VirtualNetwork {
	r := &v1alpha3.VirtualNetwork{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.VirtualNetworkSpec{
			ResourceGroupName: resourceGroupName,
			VirtualNetworkPropertiesFormat: v1alpha3.VirtualNetworkPropertiesFormat{
				AddressSpace: v1alpha3.AddressSpace{
					AddressPrefixes: []string{addressPrefix},
				},
				EnableDDOSProtection: true,
				EnableVMProtection:   true,
			},
			Location: location,
			Tags:     tags,
		},
		Status: v1alpha3.VirtualNetworkStatus{},
	}
	meta.SetExternalName(r, name)

	for _, m := range vm {
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
			name:    "NotVirtualNetwok",
			e:       &external{client: &fake.MockVirtualNetworksClient{}},
			r:       &v1alpha3.Subnet{},
			want:    &v1alpha3.Subnet{},
			wantErr: errors.New(errNotVirtualNetwork),
		},
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.VirtualNetwork) (result network.VirtualNetworksCreateOrUpdateFuture, err error) {
					return network.VirtualNetworksCreateOrUpdateFuture{}, nil
				},
			}},
			r: virtualNetwork(),
			want: virtualNetwork(
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.VirtualNetwork) (result network.VirtualNetworksCreateOrUpdateFuture, err error) {
					return network.VirtualNetworksCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r: virtualNetwork(),
			want: virtualNetwork(
				withConditions(xpv1.Creating()),
			),
			wantErr: errors.Wrap(errorBoom, errCreateVirtualNetwork),
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
			name:    "NotVirtualNetwok",
			e:       &external{client: &fake.MockVirtualNetworksClient{}},
			r:       &v1alpha3.Subnet{},
			want:    &v1alpha3.Subnet{},
			wantErr: errors.New(errNotVirtualNetwork),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.VirtualNetwork, err error) {
					return network.VirtualNetwork{
							Tags: azure.ToStringPtrMap(tags),
							VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
								AddressSpace: &network.AddressSpace{
									AddressPrefixes: &[]string{addressPrefix},
								},
								EnableDdosProtection: azure.ToBoolPtr(true),
								EnableVMProtection:   azure.ToBoolPtr(true),
							},
						}, autorest.DetailedError{
							StatusCode: http.StatusNotFound,
						}
				},
			}},
			r:    virtualNetwork(),
			want: virtualNetwork(),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.VirtualNetwork, err error) {
					return network.VirtualNetwork{
						Tags: azure.ToStringPtrMap(tags),
						VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
							AddressSpace: &network.AddressSpace{
								AddressPrefixes: &[]string{addressPrefix},
							},
							EnableDdosProtection: azure.ToBoolPtr(true),
							EnableVMProtection:   azure.ToBoolPtr(true),
							ProvisioningState:    azure.ToStringPtr(string(network.Available)),
						},
					}, nil
				},
			}},
			r: virtualNetwork(),
			want: virtualNetwork(
				withConditions(xpv1.Available()),
				withState(string(network.Available)),
			),
		},
		{
			name: "FailedObserve",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.VirtualNetwork, err error) {
					return network.VirtualNetwork{}, errorBoom
				},
			}},
			r:       virtualNetwork(),
			want:    virtualNetwork(),
			wantErr: errors.Wrap(errorBoom, errGetVirtualNetwork),
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
			name:    "NotVirtualNetwok",
			e:       &external{client: &fake.MockVirtualNetworksClient{}},
			r:       &v1alpha3.Subnet{},
			want:    &v1alpha3.Subnet{},
			wantErr: errors.New(errNotVirtualNetwork),
		},
		{
			name: "SuccessfulDoesNotNeedUpdate",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.VirtualNetwork, err error) {
					return network.VirtualNetwork{
						Tags: azure.ToStringPtrMap(tags),
						VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
							AddressSpace: &network.AddressSpace{
								AddressPrefixes: &[]string{addressPrefix},
							},
							EnableDdosProtection: azure.ToBoolPtr(true),
							EnableVMProtection:   azure.ToBoolPtr(true),
						},
					}, nil
				},
			}},
			r:    virtualNetwork(),
			want: virtualNetwork(),
		},
		{
			name: "SuccessfulNeedsUpdate",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.VirtualNetwork, err error) {
					return network.VirtualNetwork{
						Tags: azure.ToStringPtrMap(tags),
						VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
							AddressSpace: &network.AddressSpace{
								AddressPrefixes: &[]string{"10.1.0.0/16"},
							},
							EnableDdosProtection: azure.ToBoolPtr(true),
							EnableVMProtection:   azure.ToBoolPtr(true),
						},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.VirtualNetwork) (result network.VirtualNetworksCreateOrUpdateFuture, err error) {
					return network.VirtualNetworksCreateOrUpdateFuture{}, nil
				},
			}},
			r:    virtualNetwork(),
			want: virtualNetwork(),
		},
		{
			name: "UnsuccessfulGet",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.VirtualNetwork, err error) {
					return network.VirtualNetwork{
						Tags: azure.ToStringPtrMap(tags),
						VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
							AddressSpace: &network.AddressSpace{
								AddressPrefixes: &[]string{"10.1.0.0/16"},
							},
							EnableDdosProtection: azure.ToBoolPtr(true),
							EnableVMProtection:   azure.ToBoolPtr(true),
						},
					}, errorBoom
				},
			}},
			r:       virtualNetwork(),
			want:    virtualNetwork(),
			wantErr: errors.Wrap(errorBoom, errGetVirtualNetwork),
		},
		{
			name: "UnsuccessfulUpdate",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.VirtualNetwork, err error) {
					return network.VirtualNetwork{
						Tags: azure.ToStringPtrMap(tags),
						VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
							AddressSpace: &network.AddressSpace{
								AddressPrefixes: &[]string{"10.1.0.0/16"},
							},
							EnableDdosProtection: azure.ToBoolPtr(true),
							EnableVMProtection:   azure.ToBoolPtr(true),
						},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.VirtualNetwork) (result network.VirtualNetworksCreateOrUpdateFuture, err error) {
					return network.VirtualNetworksCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r:       virtualNetwork(),
			want:    virtualNetwork(),
			wantErr: errors.Wrap(errorBoom, errUpdateVirtualNetwork),
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
			name:    "NotVirtualNetwok",
			e:       &external{client: &fake.MockVirtualNetworksClient{}},
			r:       &v1alpha3.Subnet{},
			want:    &v1alpha3.Subnet{},
			wantErr: errors.New(errNotVirtualNetwork),
		},
		{
			name: "Successful",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockDelete: func(_ context.Context, _ string, _ string) (result network.VirtualNetworksDeleteFuture, err error) {
					return network.VirtualNetworksDeleteFuture{}, nil
				},
			}},
			r: virtualNetwork(),
			want: virtualNetwork(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockDelete: func(_ context.Context, _ string, _ string) (result network.VirtualNetworksDeleteFuture, err error) {
					return network.VirtualNetworksDeleteFuture{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			r: virtualNetwork(),
			want: virtualNetwork(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "Failed",
			e: &external{client: &fake.MockVirtualNetworksClient{
				MockDelete: func(_ context.Context, _ string, _ string) (result network.VirtualNetworksDeleteFuture, err error) {
					return network.VirtualNetworksDeleteFuture{}, errorBoom
				},
			}},
			r: virtualNetwork(),
			want: virtualNetwork(
				withConditions(xpv1.Deleting()),
			),
			wantErr: errors.Wrap(errorBoom, errDeleteVirtualNetwork),
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
