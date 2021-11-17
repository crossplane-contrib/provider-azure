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

package recordset

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"

	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/dns/v1alpha1"
)

type MockRecordSetAPI struct {
	MockGet            func(ctx context.Context, z *v1alpha1.RecordSet) (dns.RecordSet, error)
	MockCreateOrUpdate func(ctx context.Context, z *v1alpha1.RecordSet) error
	MockDelete         func(ctx context.Context, z *v1alpha1.RecordSet) error
}

func (m *MockRecordSetAPI) Get(ctx context.Context, z *v1alpha1.RecordSet) (dns.RecordSet, error) {
	return m.MockGet(ctx, z)
}

func (m *MockRecordSetAPI) CreateOrUpdate(ctx context.Context, z *v1alpha1.RecordSet) error {
	return m.MockCreateOrUpdate(ctx, z)
}

func (m *MockRecordSetAPI) Delete(ctx context.Context, z *v1alpha1.RecordSet) error {
	return m.MockDelete(ctx, z)
}

type modifier func(configuration *v1alpha1.RecordSet)

func withExternalName(name string) modifier {
	return func(p *v1alpha1.RecordSet) {
		meta.SetExternalName(p, name)
	}
}

func zone(m ...modifier) *v1alpha1.RecordSet {
	p := &v1alpha1.RecordSet{}

	for _, mod := range m {
		mod(p)
	}
	return p
}

func TestObserve(t *testing.T) {
	errBoom := errors.New("boom")
	name := "coolserver"

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		eo  managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"ErrNotAPostgreSQLServerConfiguration": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotDNSRecordSet),
			},
		},
		"ErrGetServer": {
			e: &external{
				client: &MockRecordSetAPI{
					MockGet: func(_ context.Context, _ *v1alpha1.RecordSet) (dns.RecordSet, error) {
						return dns.RecordSet{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  zone(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetDNSRecordSet),
			},
		},
		"ServerNotFound": {
			e: &external{
				client: &MockRecordSetAPI{
					MockGet: func(_ context.Context, _ *v1alpha1.RecordSet) (dns.RecordSet, error) {
						return dns.RecordSet{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  zone(),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"ServerAvailable": {
			e: &external{
				client: &MockRecordSetAPI{
					MockGet: func(_ context.Context, _ *v1alpha1.RecordSet) (dns.RecordSet, error) {
						return dns.RecordSet{
							RecordSetProperties: &dns.RecordSetProperties{},
						}, nil
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg: zone(
					withExternalName(name),
				),
			},
			want: want{
				eo: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eo, err := tc.e.Observe(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.eo, eo); diff != "" {
				t.Errorf("tc.e.Observe(...): -want, +got:\n%s", diff)
			}
		})
	}
}
