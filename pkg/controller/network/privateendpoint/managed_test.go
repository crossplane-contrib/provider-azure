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

package privateendpoint

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2020-03-01/network"
	"github.com/Azure/go-autorest/autorest"
	"github.com/crossplane-contrib/provider-azure/apis/network/v1alpha3"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
	"github.com/crossplane-contrib/provider-azure/pkg/clients/network/fake"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

const (
	name                          = "coolEndpoint"
	uid                           = types.UID("definitely-a-uuid")
	addressPrefix                 = "10.0.0.0/16"
	subnetId                      = "coolsubnet"
	resourceGroupName             = "coolRG"
	location                      = "West Europe"
	privateConnectionResourceType = "redisCache"
	privateConnectionResourceId   = "coolResourceId"
)

var (
	ctx       = context.Background()
	errorBoom = errors.New("boom")
	tags      = map[string]string{"one": "test", "two": "test"}
	tags2     = map[string]string{"one": "test", "two": "test", "three": "test"}
)

type testCase struct {
	name    string
	e       managed.ExternalClient
	r       resource.Managed
	want    resource.Managed
	wantErr error
}

func TestCreate(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotPrivateEndpoint",
			e:       &external{client: &fake.MockPrivateEndpointClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotPrivateEndpoint),
		},
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.PrivateEndpoint) (network.PrivateEndpointsCreateOrUpdateFuture, error) {
					return network.PrivateEndpointsCreateOrUpdateFuture{}, nil
				},
			}},
			r: privateEndpoint(),
			want: privateEndpoint(
				withConditions(xpv1.Creating()),
			),
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.PrivateEndpoint) (network.PrivateEndpointsCreateOrUpdateFuture, error) {
					return network.PrivateEndpointsCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r: privateEndpoint(),
			want: privateEndpoint(
				withConditions(xpv1.Creating()),
			),
			wantErr: errors.Wrap(errorBoom, errCreatePrivateEndpoint),
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
			name:    "NotPrivateEndpoint",
			e:       &external{client: &fake.MockPrivateEndpointClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotPrivateEndpoint),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
							Tags: azure.ToStringPtrMap(tags),
							PrivateEndpointProperties: &network.PrivateEndpointProperties{
								PrivateLinkServiceConnections: &[]network.PrivateLinkServiceConnection{
									{Name: azure.ToStringPtr("myconnection"),
										Type: azure.ToStringPtr("redisCache")},
								},
							},
						}, autorest.DetailedError{
							StatusCode: http.StatusNotFound,
						}
				},
			}},
			r:    privateEndpoint(),
			want: privateEndpoint(),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
						Tags: azure.ToStringPtrMap(tags),
						PrivateEndpointProperties: &network.PrivateEndpointProperties{
							PrivateLinkServiceConnections: &[]network.PrivateLinkServiceConnection{
								{Name: azure.ToStringPtr("myconnection"),
									Type: azure.ToStringPtr("redisCache")},
							},
							ProvisioningState: network.ProvisioningState("Available"),
						},
					}, nil
				},
			}},
			r: privateEndpoint(),
			want: privateEndpoint(
				withConditions(xpv1.Available()),
				withState(string(network.Available)),
			),
		},
		{
			name: "FailedObserve",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{}, errorBoom
				},
			}},
			r:       privateEndpoint(),
			want:    privateEndpoint(),
			wantErr: errors.Wrap(errorBoom, errGetPrivateEndpoint),
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
			name:    "NotPrivateEndpoint",
			e:       &external{client: &fake.MockPrivateEndpointClient{}},
			r:       &v1alpha3.Subnet{},
			want:    &v1alpha3.Subnet{},
			wantErr: errors.New(errNotPrivateEndpoint),
		},
		{
			name: "SuccessfulDoesNotNeedUpdate",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
						Tags: azure.ToStringPtrMap(tags),
						PrivateEndpointProperties: &network.PrivateEndpointProperties{
							Subnet: &network.Subnet{
								ID: azure.ToStringPtr(subnetId),
							},
							PrivateLinkServiceConnections: &[]network.PrivateLinkServiceConnection{
								{Name: azure.ToStringPtr("myconnection"),
									Type: azure.ToStringPtr("redisCache")},
							},
							ProvisioningState: network.ProvisioningState("Available"),
						},
					}, nil
				},
			}},
			r:    privateEndpoint(),
			want: privateEndpoint(),
		},
		{
			name: "SuccessfulNeedsUpdate",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
						Tags: azure.ToStringPtrMap(tags2),
						PrivateEndpointProperties: &network.PrivateEndpointProperties{
							Subnet: &network.Subnet{
								ID: azure.ToStringPtr(subnetId),
							},
							PrivateLinkServiceConnections: &[]network.PrivateLinkServiceConnection{
								{Name: azure.ToStringPtr("myconnection"),
									Type: azure.ToStringPtr("redisCache")},
							},
							ProvisioningState: network.ProvisioningState("Available"),
						},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.PrivateEndpoint) (result network.PrivateEndpointsCreateOrUpdateFuture, err error) {
					return network.PrivateEndpointsCreateOrUpdateFuture{}, nil
				},
			}},
			r:    privateEndpoint(),
			want: privateEndpoint(),
		},
		{
			name: "UnsuccessfulGet",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
						Tags: azure.ToStringPtrMap(tags),
						PrivateEndpointProperties: &network.PrivateEndpointProperties{
							Subnet: &network.Subnet{
								ID: azure.ToStringPtr(subnetId),
							},
							PrivateLinkServiceConnections: &[]network.PrivateLinkServiceConnection{
								{Name: azure.ToStringPtr("myconnection"),
									Type: azure.ToStringPtr("redisCache")},
							},
							ProvisioningState: network.ProvisioningState("Available"),
						},
					}, errorBoom
				},
			}},
			r:       privateEndpoint(),
			want:    privateEndpoint(),
			wantErr: errors.Wrap(errorBoom, errGetPrivateEndpoint),
		},
		{
			name: "UnsuccessfulUpdate",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.PrivateEndpoint, err error) {
					return network.PrivateEndpoint{
						Tags: azure.ToStringPtrMap(tags2),
						PrivateEndpointProperties: &network.PrivateEndpointProperties{
							Subnet: &network.Subnet{
								ID: azure.ToStringPtr(subnetId),
							},
							PrivateLinkServiceConnections: &[]network.PrivateLinkServiceConnection{
								{Name: azure.ToStringPtr("myconnection"),
									Type: azure.ToStringPtr("redisCache")},
							},
							ProvisioningState: network.ProvisioningState("Available"),
						},
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.PrivateEndpoint) (result network.PrivateEndpointsCreateOrUpdateFuture, err error) {
					return network.PrivateEndpointsCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r:       privateEndpoint(),
			want:    privateEndpoint(),
			wantErr: errors.Wrap(errorBoom, errUpdatePrivateEndpoint),
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
			name:    "NotPrivateEndpoint",
			e:       &external{client: &fake.MockPrivateEndpointClient{}},
			r:       &v1alpha3.Subnet{},
			want:    &v1alpha3.Subnet{},
			wantErr: errors.New(errNotPrivateEndpoint),
		},
		{
			name: "Successful",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockDelete: func(_ context.Context, _ string, _ string) (result network.PrivateEndpointsDeleteFuture, err error) {
					return network.PrivateEndpointsDeleteFuture{}, nil
				},
			}},
			r: privateEndpoint(),
			want: privateEndpoint(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockDelete: func(_ context.Context, _ string, _ string) (result network.PrivateEndpointsDeleteFuture, err error) {
					return network.PrivateEndpointsDeleteFuture{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			r: privateEndpoint(),
			want: privateEndpoint(
				withConditions(xpv1.Deleting()),
			),
		},
		{
			name: "Failed",
			e: &external{client: &fake.MockPrivateEndpointClient{
				MockDelete: func(_ context.Context, _ string, _ string) (result network.PrivateEndpointsDeleteFuture, err error) {
					return network.PrivateEndpointsDeleteFuture{}, errorBoom
				},
			}},
			r: privateEndpoint(),
			want: privateEndpoint(
				withConditions(xpv1.Deleting()),
			),
			wantErr: errors.Wrap(errorBoom, errDeletePrivateEndpoint),
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

type privateEndpointModifier func(*v1alpha3.PrivateEndpoint)

func privateEndpoint(sm ...privateEndpointModifier) *v1alpha3.PrivateEndpoint {
	r := &v1alpha3.PrivateEndpoint{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.PrivateEndpointSpec{
			SubnetId:          subnetId,
			ResourceGroupName: resourceGroupName,
			Location:          location,
			PrivateConnectionDetails: v1alpha3.PrivateConnectionDetails{
				IsManualConnection:          true,
				ResourceType:                privateConnectionResourceType,
				PrivateConnectionResourceId: privateConnectionResourceId,
			},
			Tags: tags,
		},
		Status: v1alpha3.PrivateEndpointStatus{
			AtProvider: v1alpha3.PrivateEndpointStatusObservation{},
		},
	}

	meta.SetExternalName(r, name)

	for _, m := range sm {
		m(r)
	}

	return r
}

func withConditions(c ...xpv1.Condition) privateEndpointModifier {
	return func(r *v1alpha3.PrivateEndpoint) { r.Status.ConditionedStatus.Conditions = c }
}

func withState(s string) privateEndpointModifier {
	return func(r *v1alpha3.PrivateEndpoint) { r.Status.AtProvider.State = s }
}
