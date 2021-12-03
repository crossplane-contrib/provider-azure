package publicipaddress

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
	name              = "coolPublicIPAddress"
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

type publicIPAddressModifier func(address *v1alpha3.PublicIPAddress)

func withConditions(c ...xpv1.Condition) publicIPAddressModifier {
	return func(r *v1alpha3.PublicIPAddress) { r.Status.ConditionedStatus.Conditions = c }
}

func withState(s string) publicIPAddressModifier {
	return func(r *v1alpha3.PublicIPAddress) { r.Status.AtProvider.State = s }
}

func publicIPAddress(sm ...publicIPAddressModifier) *v1alpha3.PublicIPAddress {
	r := &v1alpha3.PublicIPAddress{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.PublicIPAddressSpec{
			ForProvider: v1alpha3.PublicIPAddressProperties{
				ResourceGroupName: resourceGroupName,
				Tags:              make(map[string]string),
			},
		},
		Status: v1alpha3.PublicIPAddressStatus{},
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
			name:    "NotPublicIPAddress",
			e:       &external{client: &fake.MockPublicIPAddressClient{}},
			r:       &v1alpha3.Subnet{},
			want:    &v1alpha3.Subnet{},
			wantErr: errors.New(errNotPublicIPAddress),
		},
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockPublicIPAddressClient{
				MockCreateOrUpdate: func(ctx context.Context, resourceGroupName string, publicIPAddressName string, parameters network.PublicIPAddress) (result network.PublicIPAddressesCreateOrUpdateFuture, err error) {
					return network.PublicIPAddressesCreateOrUpdateFuture{}, nil
				},
			}},
			r:    publicIPAddress(),
			want: publicIPAddress(),
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.MockPublicIPAddressClient{
				MockCreateOrUpdate: func(ctx context.Context, resourceGroupName string, publicIPAddressName string, parameters network.PublicIPAddress) (result network.PublicIPAddressesCreateOrUpdateFuture, err error) {
					return network.PublicIPAddressesCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r:       publicIPAddress(),
			want:    publicIPAddress(),
			wantErr: errors.Wrap(errorBoom, errCreatePublicIPAddress),
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
			name:    "NotPublicIPAddress",
			e:       &external{client: &fake.MockPublicIPAddressClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotPublicIPAddress),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{client: &fake.MockPublicIPAddressClient{
				MockGet: func(ctx context.Context, resourceGroupName string, publicIPAddressName string, expand string) (result network.PublicIPAddress, err error) {
					return network.PublicIPAddress{
							PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
								PublicIPAllocationMethod: "static",
							},
						}, autorest.DetailedError{
							StatusCode: http.StatusNotFound,
						}
				},
			}},
			r:    publicIPAddress(),
			want: publicIPAddress(),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				client: &fake.MockPublicIPAddressClient{
					MockGet: func(ctx context.Context, resourceGroupName string, publicIPAddressName string, expand string) (result network.PublicIPAddress, err error) {
						return network.PublicIPAddress{
							PublicIPAddressPropertiesFormat: &network.PublicIPAddressPropertiesFormat{
								PublicIPAllocationMethod: "static",
								ProvisioningState:        azure.ToStringPtr(string(network.Available)),
							},
						}, nil
					},
				}},
			r: publicIPAddress(),
			want: publicIPAddress(
				withConditions(xpv1.Available()),
				withState(string(network.Available)),
			),
		},
		{
			name: "FailedObserve",
			e: &external{client: &fake.MockPublicIPAddressClient{
				MockGet: func(ctx context.Context, resourceGroupName string, publicIPAddressName string, expand string) (result network.PublicIPAddress, err error) {
					return network.PublicIPAddress{}, errorBoom
				},
			}},
			r:       publicIPAddress(),
			want:    publicIPAddress(),
			wantErr: errors.Wrap(errorBoom, errGetPublicIPAddress),
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

func TestDelete(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotPublicIPAddress",
			e:       &external{client: &fake.MockPublicIPAddressClient{}},
			r:       &v1alpha3.VirtualNetwork{},
			want:    &v1alpha3.VirtualNetwork{},
			wantErr: errors.New(errNotPublicIPAddress),
		},
		{
			name: "Successful",
			e: &external{client: &fake.MockPublicIPAddressClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, publicIPAddressName string) (result network.PublicIPAddressesDeleteFuture, err error) {
					return network.PublicIPAddressesDeleteFuture{}, nil
				},
			}},
			r:    publicIPAddress(),
			want: publicIPAddress(),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{client: &fake.MockPublicIPAddressClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, publicIPAddressName string) (result network.PublicIPAddressesDeleteFuture, err error) {
					return network.PublicIPAddressesDeleteFuture{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			r:    publicIPAddress(),
			want: publicIPAddress(),
		},
		{
			name: "Failed",
			e: &external{client: &fake.MockPublicIPAddressClient{
				MockDelete: func(ctx context.Context, resourceGroupName string, publicIPAddressName string) (result network.PublicIPAddressesDeleteFuture, err error) {
					return network.PublicIPAddressesDeleteFuture{}, errorBoom
				},
			}},
			r:       publicIPAddress(),
			want:    publicIPAddress(),
			wantErr: errors.Wrap(errorBoom, errDeletePublicIPAddress),
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

func TestUpdate(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotPublicIPAddress",
			e:       &external{client: &fake.MockPublicIPAddressClient{}},
			r:       &v1alpha3.Subnet{},
			want:    &v1alpha3.Subnet{},
			wantErr: errors.New(errNotPublicIPAddress),
		},
		{
			name: "SuccessfulUpdate",
			e: &external{client: &fake.MockPublicIPAddressClient{
				MockCreateOrUpdate: func(ctx context.Context, resourceGroupName string, publicIPAddressName string, parameters network.PublicIPAddress) (result network.PublicIPAddressesCreateOrUpdateFuture, err error) {
					return network.PublicIPAddressesCreateOrUpdateFuture{}, nil
				},
			}},
			r:    publicIPAddress(),
			want: publicIPAddress(),
		},
		{
			name: "FailedUpdate",
			e: &external{client: &fake.MockPublicIPAddressClient{
				MockCreateOrUpdate: func(ctx context.Context, resourceGroupName string, publicIPAddressName string, parameters network.PublicIPAddress) (result network.PublicIPAddressesCreateOrUpdateFuture, err error) {
					return network.PublicIPAddressesCreateOrUpdateFuture{}, errorBoom
				},
			}},
			r:       publicIPAddress(),
			want:    publicIPAddress(),
			wantErr: errors.Wrap(errorBoom, errUpdatePublicIPAddress),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Update(ctx, tc.r)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Create(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.r, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
