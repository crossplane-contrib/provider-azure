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

package networkinterface

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

	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/network/fake"
)

const (
	name              = "coolNetworkInterface"
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

type networkInterfaceModifier func(address *v1alpha3.NetworkInterface)

func withConditions(c ...xpv1.Condition) networkInterfaceModifier {
	return func(r *v1alpha3.NetworkInterface) { r.Status.ConditionedStatus.Conditions = c }
}

func withState(s string) networkInterfaceModifier {
	return func(r *v1alpha3.NetworkInterface) { r.Status.State = s }
}

func networkInterfaceAddress(sm ...networkInterfaceModifier) *v1alpha3.NetworkInterface {
	r := &v1alpha3.NetworkInterface{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.NetworkInterfaceSpec{
			ResourceGroupName: resourceGroupName,
			NetworkInterfaceFormat: v1alpha3.NetworkInterfaceFormat{
				Location:         "",
				IPConfigurations: make([]*v1alpha3.InterfaceIPConfiguration, 0),
			},
			Tags: make(map[string]*string),
		},
		Status: v1alpha3.NetworkInterfaceStatus{},
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
			name:    "NotNetworkInterface",
			e:       &external{client: &fake.MockNetworkInterfaceClient{}},
			r:       &v1alpha3.Subnet{},
			want:    &v1alpha3.Subnet{},
			wantErr: errors.New(errNotNetworkInterface),
		},
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockNetworkInterfaceClient{
				MockCreateOrUpdate: func(ctx context.Context, resourceGroupName string, networkInterfaceName string, parameters network.Interface) (result network.InterfacesCreateOrUpdateFuture, err error) {
					return network.InterfacesCreateOrUpdateFuture{}, nil
				},
			}},
			r: networkInterfaceAddress(),
			want: networkInterfaceAddress(
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.MockNetworkInterfaceClient{
				MockCreateOrUpdate: func(ctx context.Context, resourceGroupName string, networkInterfaceName string, parameters network.Interface) (result network.InterfacesCreateOrUpdateFuture, err error) {
					return network.InterfacesCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r: networkInterfaceAddress(),
			want: networkInterfaceAddress(
				withConditions(xpv1.Creating()),
			),
			wantErr: errors.Wrap(errorBoom, errCreateNetworkInterface),
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
			name:    "NotNetworkInterface",
			e:       &external{client: &fake.MockNetworkInterfaceClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotNetworkInterface),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{client: &fake.MockNetworkInterfaceClient{
				MockGet: func(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, err error) {
					return network.Interface{InterfacePropertiesFormat: &network.InterfacePropertiesFormat{}}, autorest.DetailedError{StatusCode: http.StatusNotFound}
				},
			}},
			r:    networkInterfaceAddress(),
			want: networkInterfaceAddress(),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{client: &fake.MockNetworkInterfaceClient{
				MockGet: func(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, err error) {
					return network.Interface{
						InterfacePropertiesFormat: &network.InterfacePropertiesFormat{ProvisioningState: azure.ToStringPtr(string(network.Available))},
					}, nil
				},
			}},
			r: networkInterfaceAddress(),
			want: networkInterfaceAddress(
				withConditions(xpv1.Available()),
				withState(string(network.Available)),
			),
		},
		{
			name: "FailedObserve",
			e: &external{client: &fake.MockNetworkInterfaceClient{
				MockGet: func(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, err error) {
					return network.Interface{}, errorBoom
				},
			}},
			r:       networkInterfaceAddress(),
			want:    networkInterfaceAddress(),
			wantErr: errors.Wrap(errorBoom, errGetNetworkInterface),
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
			name:    "NotNetworkInterface",
			e:       &external{client: &fake.MockNetworkInterfaceClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotNetworkInterface),
		},
		{
			name: "SuccessfulDoesNotNeedUpdate",
			e: &external{client: &fake.MockNetworkInterfaceClient{
				MockGet: func(ctx context.Context, resourceGroupName string, networkInterfaceName string, expand string) (result network.Interface, err error) {
					return network.Interface{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
				},
			}},
			r:    networkInterfaceAddress(),
			want: networkInterfaceAddress(),
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
			name:    "NotNetworkInterface",
			e:       &external{client: &fake.MockNetworkInterfaceClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotNetworkInterface),
		},
		{
			name: "Successful",
			e: &external{client: &fake.MockNetworkInterfaceClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, networkInterfaceName string) (result network.InterfacesDeleteFuture, err error) {
					return network.InterfacesDeleteFuture{}, nil
				},
			}},
			r: networkInterfaceAddress(),
			want: networkInterfaceAddress(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{client: &fake.MockNetworkInterfaceClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, networkInterfaceName string) (result network.InterfacesDeleteFuture, err error) {
					return network.InterfacesDeleteFuture{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			r: networkInterfaceAddress(),
			want: networkInterfaceAddress(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "Failed",
			e: &external{client: &fake.MockNetworkInterfaceClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, networkInterfaceName string) (result network.InterfacesDeleteFuture, err error) {
					return network.InterfacesDeleteFuture{}, errorBoom
				},
			}},
			r: networkInterfaceAddress(),
			want: networkInterfaceAddress(
				withConditions(xpv1.Deleting()),
			),
			wantErr: errors.Wrap(errorBoom, errDeleteNetworkInterface),
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
