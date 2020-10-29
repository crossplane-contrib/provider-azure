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
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/Azure/go-autorest/autorest"

	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	networkclient "github.com/crossplane/provider-azure/pkg/clients/network"
	"github.com/crossplane/provider-azure/pkg/clients/network/fake"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
)

const (
	namespace          = "coolNamespace"
	name               = "coolSubnet"
	uid                = types.UID("definitely-a-uuid")
	addressPrefix      = "10.0.0.0/16"
	virtualNetworkName = "coolVnet"
	resourceGroupName  = "coolRG"

	providerName       = "cool-aws"
	providerSecretName = "cool-aws-secret"
	providerSecretKey  = "credentials"
	providerSecretData = "definitelyini"
)

var (
	ctx       = context.Background()
	errorBoom = errors.New("boom")

	provider = azurev1alpha3.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: providerName},
		Spec: azurev1alpha3.ProviderSpec{
			ProviderSpec: runtimev1alpha1.ProviderSpec{
				CredentialsSecretRef: &runtimev1alpha1.SecretKeySelector{
					SecretReference: runtimev1alpha1.SecretReference{
						Namespace: namespace,
						Name:      providerSecretName,
					},
					Key: providerSecretKey,
				},
			},
		},
	}

	providerSecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data:       map[string][]byte{providerSecretKey: []byte(providerSecretData)},
	}
)

type testCase struct {
	name    string
	e       managed.ExternalClient
	r       resource.Managed
	want    resource.Managed
	wantErr error
}

type subnetModifier func(*v1alpha3.Subnet)

func withConditions(c ...runtimev1alpha1.Condition) subnetModifier {
	return func(r *v1alpha3.Subnet) { r.Status.ConditionedStatus.Conditions = c }
}

func withState(s string) subnetModifier {
	return func(r *v1alpha3.Subnet) { r.Status.State = s }
}

func withProviderRef(p runtimev1alpha1.Reference) subnetModifier {
	return func(r *v1alpha3.Subnet) { r.Spec.ProviderReference = p.DeepCopy() }
}

func subnet(sm ...subnetModifier) *v1alpha3.Subnet {
	r := &v1alpha3.Subnet{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.SubnetSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &runtimev1alpha1.Reference{Name: providerName},
			},
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
				withConditions(runtimev1alpha1.Creating()),
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
				withConditions(runtimev1alpha1.Creating()),
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
				withConditions(runtimev1alpha1.Available()),
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
				withConditions(runtimev1alpha1.Deleting()),
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
				withConditions(runtimev1alpha1.Deleting()),
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
				withConditions(runtimev1alpha1.Deleting()),
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

func TestConnect(t *testing.T) {
	cases := []struct {
		name    string
		conn    *connecter
		i       resource.Managed
		want    managed.ExternalClient
		wantErr error
	}{
		{
			name:    "NotSubnet",
			conn:    &connecter{client: &test.MockClient{}},
			i:       &v1alpha3.VirtualNetwork{},
			want:    nil,
			wantErr: errors.New(errNotSubnet),
		},
		{
			name: "ErrGetProviderSecret",
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							*obj.(*azurev1alpha3.Provider) = provider
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							return errorBoom
						}
						return nil
					},
				},
				newClientFn: func(_ context.Context, _ []byte) (networkclient.SubnetsClient, error) {
					return &fake.MockSubnetsClient{}, nil
				},
			},
			i:       subnet(withProviderRef(runtimev1alpha1.Reference{Name: providerName})),
			wantErr: errors.Wrapf(errorBoom, "cannot get provider secret %s", fmt.Sprintf("%s/%s", namespace, providerSecretName)),
		},
		{
			name: "ErrProviderSecretNil",
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							nilSecretProvider := provider
							nilSecretProvider.SetCredentialsSecretReference(nil)
							*obj.(*azurev1alpha3.Provider) = nilSecretProvider
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							return errorBoom
						}
						return nil
					},
				},
				newClientFn: func(_ context.Context, _ []byte) (networkclient.SubnetsClient, error) {
					return &fake.MockSubnetsClient{}, nil
				},
			},
			i:       subnet(withProviderRef(runtimev1alpha1.Reference{Name: providerName})),
			wantErr: errors.New(errProviderSecretNil),
		},
		{
			name: "SuccessfulConnect",
			conn: &connecter{
				client: &test.MockClient{
					MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
						switch key {
						case client.ObjectKey{Name: providerName}:
							*obj.(*azurev1alpha3.Provider) = provider
						case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
							*obj.(*corev1.Secret) = providerSecret
						}
						return nil
					},
				},
				newClientFn: func(_ context.Context, _ []byte) (networkclient.SubnetsClient, error) {
					return &fake.MockSubnetsClient{}, nil
				},
			},
			i:    subnet(withProviderRef(runtimev1alpha1.Reference{Name: providerName})),
			want: &external{client: &fake.MockSubnetsClient{}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, gotErr := tc.conn.Connect(ctx, tc.i)

			if diff := cmp.Diff(tc.wantErr, gotErr, test.EquateErrors()); diff != "" {
				t.Errorf("tc.conn.Connect(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, got, test.EquateConditions(), cmp.AllowUnexported(external{})); diff != "" {
				t.Errorf("tc.conn.Connect(...): -want, +got:\n%s", diff)
			}
		})
	}
}
