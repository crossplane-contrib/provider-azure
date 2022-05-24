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

package application

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
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
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/rbac/fake"
)

const (
	name     = "application"
	homePage = "homePage"
	uid      = types.UID("definitely-a-uuid")
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

type applicationModifier func(*v1alpha1.Application)

func withConditions(c ...xpv1.Condition) applicationModifier {
	return func(r *v1alpha1.Application) { r.Status.ConditionedStatus.Conditions = c }
}

func withExternalName(name string) applicationModifier {
	return func(r *v1alpha1.Application) { meta.SetExternalName(r, name) }
}

func application(sm ...applicationModifier) *v1alpha1.Application {
	r := &v1alpha1.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha1.ApplicationSpec{
			ForProvider: v1alpha1.ApplicationParameters{
				AvailableToOtherTenants: azure.ToBoolPtr(true),
				DisplayName:             azure.ToStringPtr(name),
				Homepage:                azure.ToStringPtr(homePage),
				IdentifierURIs:          []string{homePage},
			},
		},
		Status: v1alpha1.ApplicationStatus{},
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
			name:    "NotApplication",
			e:       &external{c: &fake.MockApplicationsClient{}},
			r:       &v1alpha1.ServicePrincipal{},
			want:    &v1alpha1.ServicePrincipal{},
			wantErr: errors.New(errNotApplication),
		},
		{
			name: "SuccessfulCreate",
			e: &external{c: &fake.MockApplicationsClient{
				MockCreate: func(ctx context.Context, parameters graphrbac.ApplicationCreateParameters) (result graphrbac.Application, err error) {
					return graphrbac.Application{}, nil
				},
			}},
			r:    application(),
			want: application(),
		},
		{
			name: "FailedCreate",
			e: &external{c: &fake.MockApplicationsClient{
				MockCreate: func(ctx context.Context, parameters graphrbac.ApplicationCreateParameters) (result graphrbac.Application, err error) {
					return graphrbac.Application{}, errorBoom
				},
			}},
			r:       application(),
			want:    application(),
			wantErr: errors.Wrap(errorBoom, errCreateApplication),
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
			name:    "NotApplication",
			e:       &external{c: &fake.MockApplicationsClient{}},
			r:       &v1alpha1.ServicePrincipal{},
			want:    &v1alpha1.ServicePrincipal{},
			wantErr: errors.New(errNotApplication),
		},
		{
			name: "SuccessfulObserveNotCreated",
			e:    &external{c: &fake.MockApplicationsClient{}},
			r:    application(),
			want: application(),
		},
		{
			name: "SuccessfulObserveNotExist",
			e: &external{c: &fake.MockApplicationsClient{
				MockGet: func(ctx context.Context, applicationObjectID string) (result graphrbac.Application, err error) {
					return graphrbac.Application{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
				},
			}},
			r:    application(withExternalName(name)),
			want: application(withExternalName(name)),
		},
		{
			name: "SuccessfulObserveExists",
			e: &external{c: &fake.MockApplicationsClient{
				MockGet: func(ctx context.Context, applicationObjectID string) (result graphrbac.Application, err error) {
					return graphrbac.Application{}, nil
				},
			}},
			r: application(withExternalName(name)),
			want: application(
				withConditions(xpv1.Available()),
				withExternalName(name),
			),
		},
		{
			name: "FailedObserve",
			e: &external{c: &fake.MockApplicationsClient{
				MockGet: func(ctx context.Context, applicationObjectID string) (result graphrbac.Application, err error) {
					return graphrbac.Application{}, errorBoom
				},
			}},
			r:       application(withExternalName(name)),
			want:    application(withExternalName(name)),
			wantErr: errors.Wrap(errorBoom, errGetApplication),
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
			e:       &external{c: &fake.MockApplicationsClient{}},
			r:       &v1alpha1.Application{},
			want:    &v1alpha1.Application{},
			wantErr: nil,
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
			name:    "NotApplication",
			e:       &external{c: &fake.MockApplicationsClient{}},
			r:       &v1alpha1.ServicePrincipal{},
			want:    &v1alpha1.ServicePrincipal{},
			wantErr: errors.New(errNotApplication),
		},
		{
			name: "Successful",
			e: &external{c: &fake.MockApplicationsClient{
				MockDelete: func(ctx context.Context, applicationObjectID string) (result autorest.Response, err error) {
					return autorest.Response{}, nil
				},
			}},
			r:    application(),
			want: application(),
		},
		{
			name: "SuccessfulNotFound",
			e: &external{c: &fake.MockApplicationsClient{
				MockDelete: func(ctx context.Context, applicationObjectID string) (result autorest.Response, err error) {
					return autorest.Response{}, autorest.DetailedError{
						StatusCode: http.StatusNotFound,
					}
				},
			}},
			r:    application(),
			want: application(),
		},
		{
			name: "Failed",
			e: &external{c: &fake.MockApplicationsClient{
				MockDelete: func(ctx context.Context, applicationObjectID string) (result autorest.Response, err error) {
					return autorest.Response{}, errorBoom
				},
			}},
			r:       application(),
			want:    application(),
			wantErr: errors.Wrap(errorBoom, errDeleteApplication),
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
