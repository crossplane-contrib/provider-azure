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

package subnet

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
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

	azure "github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients"
	"github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients/network/fake"

	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/network/v1alpha3"
)

const (
	name               = "coolSubnet"
	uid                = types.UID("definitely-a-uuid")
	addressPrefix      = "10.0.0.0/16"
	virtualNetworkName = "coolVnet"
	resourceGroupName  = "coolRG"
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

type subnetModifier func(*v1alpha3.Subnet)

func withConditions(c ...xpv1.Condition) subnetModifier {
	return func(r *v1alpha3.Subnet) { r.Status.ConditionedStatus.Conditions = c }
}

func withState(s string) subnetModifier {
	return func(r *v1alpha3.Subnet) { r.Status.State = s }
}
func subnet(sm ...subnetModifier) *v1alpha3.Subnet {
	r := &v1alpha3.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.SubnetSpec{
			VirtualNetworkName: virtualNetworkName,
			ResourceGroupName:  resourceGroupName,
			SubnetPropertiesFormat: v1alpha3.SubnetPropertiesFormat{
				AddressPrefix: addressPrefix,
			},
		},
		Status: v1alpha3.SubnetStatus{},
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
			name:    "NotSubnet",
			e:       &external{client: &fake.MockSubnetsClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotSubnet),
		},
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockSubnetsClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ network.Subnet) (network.SubnetsCreateOrUpdateFuture, error) {
					return network.SubnetsCreateOrUpdateFuture{}, nil
				},
			}},
			r: subnet(),
			want: subnet(
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.MockSubnetsClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ network.Subnet) (network.SubnetsCreateOrUpdateFuture, error) {
					return network.SubnetsCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r: subnet(),
			want: subnet(
				withConditions(xpv1.Creating()),
			),
			wantErr: errors.Wrap(errorBoom, errCreateSubnet),
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
			name:    "NotSubnet",
			e:       &external{client: &fake.MockSubnetsClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotSubnet),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{client: &fake.MockSubnetsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string, _ string) (result network.Subnet, err error) {
					return network.Subnet{
							SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
								AddressPrefix: azure.ToStringPtr(addressPrefix),
							},
						}, autorest.DetailedError{
							StatusCode: http.StatusNotFound,
						}
				},
			}},
			r:    subnet(),
			want: subnet(),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{client: &fake.MockSubnetsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string, _ string) (result network.Subnet, err error) {
					return network.Subnet{
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix:     azure.ToStringPtr(addressPrefix),
							ProvisioningState: azure.ToStringPtr(string(network.Available)),
						},
					}, nil
				},
			}},
			r: subnet(),
			want: subnet(
				withConditions(xpv1.Available()),
				withState(string(network.Available)),
			),
		},
		{
			name: "FailedObserve",
			e: &external{client: &fake.MockSubnetsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string, _ string) (result network.Subnet, err error) {
					return network.Subnet{}, errorBoom
				},
			}},
			r:       subnet(),
			want:    subnet(),
			wantErr: errors.Wrap(errorBoom, errGetSubnet),
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
			name:    "NotSubnet",
			e:       &external{client: &fake.MockSubnetsClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotSubnet),
		},
		{
			name: "SuccessfulDoesNotNeedUpdate",
			e: &external{client: &fake.MockSubnetsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string, _ string) (result network.Subnet, err error) {
					return network.Subnet{
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: azure.ToStringPtr(addressPrefix),
						},
					}, nil
				},
			}},
			r:    subnet(),
			want: subnet(),
		},
		{
			name: "SuccessfulNeedsUpdate",
			e: &external{client: &fake.MockSubnetsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string, _ string) (result network.Subnet, err error) {
					return network.Subnet{
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: azure.ToStringPtr("10.1.0.0/16"),
						},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ network.Subnet) (network.SubnetsCreateOrUpdateFuture, error) {
					return network.SubnetsCreateOrUpdateFuture{}, nil
				},
			}},
			r:    subnet(),
			want: subnet(),
		},
		{
			name: "UnsuccessfulGet",
			e: &external{client: &fake.MockSubnetsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string, _ string) (result network.Subnet, err error) {
					return network.Subnet{
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: azure.ToStringPtr(addressPrefix),
						},
					}, errorBoom
				},
			}},
			r:       subnet(),
			want:    subnet(),
			wantErr: errors.Wrap(errorBoom, errGetSubnet),
		},
		{
			name: "UnsuccessfulUpdate",
			e: &external{client: &fake.MockSubnetsClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string, _ string) (result network.Subnet, err error) {
					return network.Subnet{
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: azure.ToStringPtr("10.1.0.0/16"),
						},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ string, _ network.Subnet) (network.SubnetsCreateOrUpdateFuture, error) {
					return network.SubnetsCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r:       subnet(),
			want:    subnet(),
			wantErr: errors.Wrap(errorBoom, errUpdateSubnet),
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
			name:    "NotSubnet",
			e:       &external{client: &fake.MockSubnetsClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotSubnet),
		},
		{
			name: "Successful",
			e: &external{client: &fake.MockSubnetsClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string) (result network.SubnetsDeleteFuture, err error) {
					return network.SubnetsDeleteFuture{}, nil
				},
			}},
			r: subnet(),
			want: subnet(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{client: &fake.MockSubnetsClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string) (result network.SubnetsDeleteFuture, err error) {
					return network.SubnetsDeleteFuture{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			r: subnet(),
			want: subnet(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "Failed",
			e: &external{client: &fake.MockSubnetsClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, virtualNetworkName string, subnetName string) (result network.SubnetsDeleteFuture, err error) {
					return network.SubnetsDeleteFuture{}, errorBoom
				},
			}},
			r: subnet(),
			want: subnet(
				withConditions(xpv1.Deleting()),
			),
			wantErr: errors.Wrap(errorBoom, errDeleteSubnet),
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
