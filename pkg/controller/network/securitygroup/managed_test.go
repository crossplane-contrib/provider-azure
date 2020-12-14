package SecurityGroup

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/Azure/go-autorest/autorest"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"
	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/network/fake"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"net/http"
	"testing"
)

const (
	name              = "coolNSG"
	ruleName1         = "coolSecurityRule1"
	ruleName2         = "coolSecurityRule2"
	uid               = types.UID("definitely-a-uuid")
	etagRule1         = "definitely-a-etag1"
	etagRule2         = "definitely-a-etag2"
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
	sg      resource.Managed
	want    resource.Managed
	wantErr error
}

type securityGroupModifier func(group *v1alpha3.SecurityGroup)

func withConditions(c ...runtimev1alpha1.Condition) securityGroupModifier {
	return func(r *v1alpha3.SecurityGroup) { r.Status.ConditionedStatus.Conditions = c }
}

func withState(s string) securityGroupModifier {
	return func(r *v1alpha3.SecurityGroup) { r.Status.State = s }
}

func SecurityGroup(sm ...securityGroupModifier) *v1alpha3.SecurityGroup {
	s := &v1alpha3.SecurityGroup{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.SecurityGroupSpec{
			ResourceGroupName: resourceGroupName,
			Location:          location,
			SecurityGroupPropertiesFormat: v1alpha3.SecurityGroupPropertiesFormat{
				SecurityRules:        setRules(),
				DefaultSecurityRules: nil,
			},
			Tags: tags,
		},
		Status: v1alpha3.SecurityGroupStatus{},
	}
	meta.SetExternalName(s, name)
	for _, m := range sm {
		m(s)
	}

	return s
}

func setRules() *[]v1alpha3.SecurityRule {
	var securityRules = new([]v1alpha3.SecurityRule)
	var rule1 = v1alpha3.SecurityRule{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Properties: v1alpha3.SecurityRulePropertiesFormat{
			Description:              "Test Description",
			Protocol:                 "TEST",
			SourcePortRange:          "8080",
			DestinationPortRange:     "80",
			SourceAddressPrefix:      "Internet",
			DestinationAddressPrefix: "*",
			Access:                   "Allow",
			Priority:                 120,
			Direction:                "Inbound",
		},
		Name: ruleName1,
		Etag: etagRule1,
	}
	var rule2 = v1alpha3.SecurityRule{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Properties: v1alpha3.SecurityRulePropertiesFormat{
			Description:              "Test Description",
			Protocol:                 "TEST",
			SourcePortRange:          "8080",
			DestinationPortRange:     "80",
			SourceAddressPrefix:      "Internet",
			DestinationAddressPrefix: "*",
			Access:                   "Deny",
			Priority:                 130,
			Direction:                "Outbound",
		},
		Name: ruleName2,
		Etag: etagRule2,
	}
	*securityRules = append(*securityRules, rule1)
	*securityRules = append(*securityRules, rule2)
	return securityRules
}

func setSecurityRules() *[]network.SecurityRule {
	var securityRules = new([]network.SecurityRule)
	var rule1 = network.SecurityRule{
		SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
			Description:              azure.ToStringPtr("Test Description"),
			Protocol:                 network.SecurityRuleProtocol("TCP"),
			SourcePortRange:          azure.ToStringPtr("8080"),
			DestinationPortRange:     azure.ToStringPtr("80"),
			SourceAddressPrefix:      azure.ToStringPtr("Internet"),
			DestinationAddressPrefix: azure.ToStringPtr("*"),
			Access:                   "Allow",
			Priority:                 azure.ToInt32Ptr(120),
			Direction:                "Inbound",
		},
		Name: azure.ToStringPtr(ruleName1),
		Etag: azure.ToStringPtr(etagRule1),
	}
	var rule2 = network.SecurityRule{
		SecurityRulePropertiesFormat: &network.SecurityRulePropertiesFormat{
			Description:              azure.ToStringPtr("Test Description"),
			Protocol:                 network.SecurityRuleProtocol("TCP"),
			SourcePortRange:          azure.ToStringPtr("8080"),
			DestinationPortRange:     azure.ToStringPtr("80"),
			SourceAddressPrefix:      azure.ToStringPtr("Internet"),
			DestinationAddressPrefix: azure.ToStringPtr("*"),
			Access:                   "Deny",
			Priority:                 azure.ToInt32Ptr(130),
			Direction:                "Outbound",
		},
		Name: azure.ToStringPtr(ruleName2),
		Etag: azure.ToStringPtr(etagRule2),
	}
	*securityRules = append(*securityRules, rule1)
	*securityRules = append(*securityRules, rule2)
	return securityRules
}

// Test that our Reconciler implementation satisfies the Reconciler interface.
var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connecter{}

func TestCreate(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotSecurityGroup",
			e:       &external{client: &fake.MockSecurityGroupClient{}},
			sg:      nil,
			want:    nil,
			wantErr: errors.New(errNotSecurityGroup),
		},
		{
			name: "SuccessfulCreate",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.SecurityGroup) (result network.SecurityGroupsCreateOrUpdateFuture, err error) {
					return network.SecurityGroupsCreateOrUpdateFuture{}, nil
				},
			}},
			sg: SecurityGroup(),
			want: SecurityGroup(
				withConditions(runtimev1alpha1.Creating()),
			),
		},
		{
			name: "FailedCreate",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.SecurityGroup) (result network.SecurityGroupsCreateOrUpdateFuture, err error) {
					return network.SecurityGroupsCreateOrUpdateFuture{}, errorBoom
				},
			}},
			sg: SecurityGroup(),
			want: SecurityGroup(
				withConditions(runtimev1alpha1.Creating()),
			),
			wantErr: errors.Wrap(errorBoom, errCreateSecurityGroup),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Create(ctx, tc.sg)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Create(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.sg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotSecurityGroup",
			e:       &external{client: &fake.MockSecurityGroupClient{}},
			sg:      nil,
			want:    nil,
			wantErr: errors.New(errNotSecurityGroup),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.SecurityGroup, err error) {
					return network.SecurityGroup{
							SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
								SecurityRules:        setSecurityRules(),
								DefaultSecurityRules: setSecurityRules(),
							},
							Name:     azure.ToStringPtr(name),
							Location: azure.ToStringPtr(location),
							Tags:     azure.ToStringPtrMap(tags),
						}, autorest.DetailedError{
							StatusCode: http.StatusNotFound,
						}
				},
			}},
			sg:   SecurityGroup(),
			want: SecurityGroup(),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.SecurityGroup, err error) {
					return network.SecurityGroup{
						SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
							SecurityRules:        setSecurityRules(),
							DefaultSecurityRules: setSecurityRules(),
							ProvisioningState:    azure.ToStringPtr(string(network.Available)),
						},
						Name:     azure.ToStringPtr(name),
						Location: azure.ToStringPtr(location),
						Tags:     azure.ToStringPtrMap(tags),
					}, nil
				},
			}},
			sg: SecurityGroup(),
			want: SecurityGroup(
				withConditions(runtimev1alpha1.Available()),
				withState(string(network.Available)),
			),
		},
		{
			name: "FailedObserve",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.SecurityGroup, err error) {
					return network.SecurityGroup{}, errorBoom
				},
			}},
			sg:      SecurityGroup(),
			want:    SecurityGroup(),
			wantErr: errors.Wrap(errorBoom, errGetSecurityGroup),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Observe(ctx, tc.sg)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.sg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotSecurityGroup",
			e:       &external{client: &fake.MockSecurityGroupClient{}},
			sg:      nil,
			want:    nil,
			wantErr: errors.New(errNotSecurityGroup),
		},
		{
			name: "SuccessfulDoesNotNeedUpdate",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.SecurityGroup, err error) {
					return network.SecurityGroup{
						SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
							SecurityRules:        setSecurityRules(),
							DefaultSecurityRules: nil,
							ProvisioningState:    azure.ToStringPtr(string(network.Available)),
						},
						Name:     azure.ToStringPtr(name),
						Location: azure.ToStringPtr(location),
						Tags:     azure.ToStringPtrMap(tags),
					}, nil
				},
			}},
			sg:   SecurityGroup(),
			want: SecurityGroup(),
		},
		{
			name: "SuccessfulNeedsUpdate",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.SecurityGroup, err error) {
					return network.SecurityGroup{
						SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
							SecurityRules:        setSecurityRules(),
							DefaultSecurityRules: nil,
							ProvisioningState:    azure.ToStringPtr(string(network.Available)),
						},
						Name:     azure.ToStringPtr("new Name"),
						Location: azure.ToStringPtr(location),
						Tags:     azure.ToStringPtrMap(tags),
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.SecurityGroup) (result network.SecurityGroupsCreateOrUpdateFuture, err error) {
					return network.SecurityGroupsCreateOrUpdateFuture{}, nil
				},
			}},
			sg:   SecurityGroup(),
			want: SecurityGroup(),
		},
		{
			name: "UnsuccessfulGet",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.SecurityGroup, err error) {
					return network.SecurityGroup{
						SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
							SecurityRules:        setSecurityRules(),
							DefaultSecurityRules: nil,
							ProvisioningState:    azure.ToStringPtr(string(network.Available)),
						},
						Name:     azure.ToStringPtr(name),
						Location: azure.ToStringPtr(location),
						Tags:     azure.ToStringPtrMap(tags),
					}, errorBoom
				},
			}},
			sg:      SecurityGroup(),
			want:    SecurityGroup(),
			wantErr: errors.Wrap(errorBoom, errGetSecurityGroup),
		},
		{
			name: "UnsuccessfulUpdate",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockGet: func(_ context.Context, _ string, _ string, _ string) (result network.SecurityGroup, err error) {
					return network.SecurityGroup{
						SecurityGroupPropertiesFormat: &network.SecurityGroupPropertiesFormat{
							SecurityRules:        setSecurityRules(),
							DefaultSecurityRules: nil,
							ProvisioningState:    azure.ToStringPtr(string(network.Available)),
						},
						Name:     azure.ToStringPtr("new name"),
						Location: azure.ToStringPtr(location),
						Tags:     azure.ToStringPtrMap(tags),
					}, nil
				},
				MockCreateOrUpdate: func(_ context.Context, _ string, _ string, _ network.SecurityGroup) (result network.SecurityGroupsCreateOrUpdateFuture, err error) {
					return network.SecurityGroupsCreateOrUpdateFuture{}, errorBoom
				},
			}},
			sg:      SecurityGroup(),
			want:    SecurityGroup(),
			wantErr: errors.Wrap(errorBoom, errUpdateSecurityGroup),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.e.Update(ctx, tc.sg)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Update(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.sg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	cases := []testCase{
		{
			name:    "NotSecurityGroup",
			e:       &external{client: &fake.MockSecurityGroupClient{}},
			sg:      nil,
			want:    nil,
			wantErr: errors.New(errNotSecurityGroup),
		},
		{
			name: "Successful",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockDelete: func(_ context.Context, _ string, _ string) (result network.SecurityGroupsDeleteFuture, err error) {
					return network.SecurityGroupsDeleteFuture{}, nil
				},
			}},
			sg: SecurityGroup(),
			want: SecurityGroup(
				withConditions(runtimev1alpha1.Deleting()),
			),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockDelete: func(_ context.Context, _ string, _ string) (result network.SecurityGroupsDeleteFuture, err error) {
					return network.SecurityGroupsDeleteFuture{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			sg: SecurityGroup(),
			want: SecurityGroup(
				withConditions(runtimev1alpha1.Deleting()),
			),
		},
		{
			name: "Failed",
			e: &external{client: &fake.MockSecurityGroupClient{
				MockDelete: func(_ context.Context, _ string, _ string) (result network.SecurityGroupsDeleteFuture, err error) {
					return network.SecurityGroupsDeleteFuture{}, errorBoom
				},
			}},
			sg: SecurityGroup(),
			want: SecurityGroup(
				withConditions(runtimev1alpha1.Deleting()),
			),
			wantErr: errors.Wrap(errorBoom, errDeleteSecurityGroup),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.e.Delete(ctx, tc.sg)

			if diff := cmp.Diff(tc.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want, tc.sg, test.EquateConditions()); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
