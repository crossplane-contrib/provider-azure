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

package roleassignment

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
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

	"github.com/crossplane/provider-azure/apis/rbac/v1alpha1"
	"github.com/crossplane/provider-azure/pkg/clients/rbac/fake"
)

const (
	name  = "roleAssignment"
	scope = "scope"
	uid   = types.UID("definitely-a-uuid")
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

type roleAssignmentModifier func(*v1alpha1.RoleAssignment)

func withConditions(c ...xpv1.Condition) roleAssignmentModifier {
	return func(r *v1alpha1.RoleAssignment) { r.Status.ConditionedStatus.Conditions = c }
}

func withExternalName(name string) roleAssignmentModifier {
	return func(r *v1alpha1.RoleAssignment) { meta.SetExternalName(r, name) }
}

func roleAssignment(sm ...roleAssignmentModifier) *v1alpha1.RoleAssignment {
	r := &v1alpha1.RoleAssignment{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha1.RoleAssignmentSpec{
			ForProvider: v1alpha1.RoleAssignmentParameters{
				PrincipalID: string(uid),
				RoleID:      string(uid),
				Scope:       scope,
			},
		},
		Status: v1alpha1.RoleAssignmentStatus{},
	}

	meta.SetExternalName(r, "")

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
			name:    "NotRoleAssignment",
			e:       &external{c: &fake.MockRoleAssignmentClient{}},
			r:       &v1alpha1.ServicePrincipal{},
			want:    &v1alpha1.ServicePrincipal{},
			wantErr: errors.New(errNotRoleAssignment),
		},
		{
			name: "SuccessfulCreate",
			e: &external{c: &fake.MockRoleAssignmentClient{
				MockCreate: func(ctx context.Context, scope string, roleAssignmentName string, parameters authorization.RoleAssignmentCreateParameters) (result authorization.RoleAssignment, err error) {
					return authorization.RoleAssignment{}, nil
				},
			}},
			r:    roleAssignment(),
			want: roleAssignment(),
		},
		{
			name: "FailedCreate",
			e: &external{c: &fake.MockRoleAssignmentClient{
				MockCreate: func(ctx context.Context, scope string, roleAssignmentName string, parameters authorization.RoleAssignmentCreateParameters) (result authorization.RoleAssignment, err error) {
					return authorization.RoleAssignment{}, errorBoom
				},
			}},
			r:       roleAssignment(),
			want:    roleAssignment(),
			wantErr: errors.Wrap(errorBoom, errCreateRoleAssignment),
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
			name:    "NotRoleAssignment",
			e:       &external{c: &fake.MockRoleAssignmentClient{}},
			r:       &v1alpha1.ServicePrincipal{},
			want:    &v1alpha1.ServicePrincipal{},
			wantErr: errors.New(errNotRoleAssignment),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{c: &fake.MockRoleAssignmentClient{
				MockListForScopeComplete: func(ctx context.Context, scope string, filter string) (result authorization.RoleAssignmentListResultIterator, err error) {
					return authorization.RoleAssignmentListResultIterator{}, nil
				},
			}},
			r:    roleAssignment(),
			want: roleAssignment(),
		},
		{
			name: "FailedObserve",
			e: &external{c: &fake.MockRoleAssignmentClient{
				MockListForScopeComplete: func(ctx context.Context, scope string, filter string) (result authorization.RoleAssignmentListResultIterator, err error) {
					return authorization.RoleAssignmentListResultIterator{}, errorBoom
				},
			}},
			r:       roleAssignment(),
			want:    roleAssignment(),
			wantErr: errors.Wrap(errorBoom, errGetRoleAssignment),
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
			name:    "UpdateNotSupported",
			e:       &external{c: &fake.MockRoleAssignmentClient{}},
			r:       &v1alpha1.Application{},
			want:    &v1alpha1.Application{},
			wantErr: errors.New(errRoleAssignmentUpdateNotSupported),
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
			name:    "NotRoleAssignment",
			e:       &external{c: &fake.MockRoleAssignmentClient{}},
			r:       &v1alpha1.ServicePrincipal{},
			want:    &v1alpha1.ServicePrincipal{},
			wantErr: errors.New(errNotRoleAssignment),
		},
		{
			name: "Successful",
			e: &external{c: &fake.MockRoleAssignmentClient{
				MockDelete: func(ctx context.Context, scope string, roleAssignmentName string) (result authorization.RoleAssignment, err error) {
					return authorization.RoleAssignment{}, nil
				},
			}},
			r:    roleAssignment(),
			want: roleAssignment(),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{c: &fake.MockRoleAssignmentClient{
				MockDelete: func(ctx context.Context, scope string, roleAssignmentName string) (result authorization.RoleAssignment, err error) {
					return authorization.RoleAssignment{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			r:    roleAssignment(),
			want: roleAssignment(),
		},
		{
			name: "Failed",
			e: &external{c: &fake.MockRoleAssignmentClient{
				MockDelete: func(ctx context.Context, scope string, roleAssignmentName string) (result authorization.RoleAssignment, err error) {
					return authorization.RoleAssignment{}, errorBoom
				},
			}},
			r:       roleAssignment(),
			want:    roleAssignment(),
			wantErr: errors.Wrap(errorBoom, errDeleteRoleAssignment),
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
