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

package zone

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

	"github.com/crossplane/provider-azure/apis/classic/dns/v1alpha1"
)

type MockZoneAPI struct {
	MockGet            func(ctx context.Context, z *v1alpha1.Zone) (dns.Zone, error)
	MockCreateOrUpdate func(ctx context.Context, z *v1alpha1.Zone) error
	MockDelete         func(ctx context.Context, z *v1alpha1.Zone) error
}

func (m *MockZoneAPI) Get(ctx context.Context, z *v1alpha1.Zone) (dns.Zone, error) {
	return m.MockGet(ctx, z)
}

func (m *MockZoneAPI) CreateOrUpdate(ctx context.Context, z *v1alpha1.Zone) error {
	return m.MockCreateOrUpdate(ctx, z)
}

func (m *MockZoneAPI) Delete(ctx context.Context, z *v1alpha1.Zone) error {
	return m.MockDelete(ctx, z)
}

type modifier func(configuration *v1alpha1.Zone)

func withExternalName(name string) modifier {
	return func(p *v1alpha1.Zone) {
		meta.SetExternalName(p, name)
	}
}

func zone(m ...modifier) *v1alpha1.Zone {
	p := &v1alpha1.Zone{}

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
				err: errors.New(errNotDNSZone),
			},
		},
		"ErrGetServer": {
			e: &external{
				client: &MockZoneAPI{
					MockGet: func(_ context.Context, _ *v1alpha1.Zone) (dns.Zone, error) {
						return dns.Zone{}, errBoom
					},
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  zone(),
			},
			want: want{
				err: errors.Wrap(errBoom, errGetDNSZone),
			},
		},
		"ServerNotFound": {
			e: &external{
				client: &MockZoneAPI{
					MockGet: func(_ context.Context, _ *v1alpha1.Zone) (dns.Zone, error) {
						return dns.Zone{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
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
				client: &MockZoneAPI{
					MockGet: func(_ context.Context, _ *v1alpha1.Zone) (dns.Zone, error) {
						return dns.Zone{
							ZoneProperties: &dns.ZoneProperties{},
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
					ResourceExists: true,
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

func TestCreate(t *testing.T) {
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}
	type want struct {
		ec  managed.ExternalCreation
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
				err: errors.New(errNotDNSZone),
			},
		},
		"ErrCreateServer": {
			e: &external{
				client: &MockZoneAPI{
					MockCreateOrUpdate: func(_ context.Context, _ *v1alpha1.Zone) error { return errBoom },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  zone(),
			},
			want: want{
				err: errors.Wrap(errBoom, errCreateDNSZone),
			},
		},
		"Successful": {
			e: &external{
				client: &MockZoneAPI{
					MockCreateOrUpdate: func(_ context.Context, _ *v1alpha1.Zone) error { return nil },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  zone(),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ec, err := tc.e.Create(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Create(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.ec, ec); diff != "" {
				t.Errorf("tc.e.Create(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want error
	}{
		"ErrNotAPostgreSQLServerConfiguration": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: errors.New(errNotDNSZone),
		},
		"ErrDeleteServer": {
			e: &external{
				client: &MockZoneAPI{
					MockGet: func(_ context.Context, _ *v1alpha1.Zone) (dns.Zone, error) {
						return dns.Zone{
							ZoneProperties: &dns.ZoneProperties{},
						}, nil
					},
					MockDelete: func(_ context.Context, _ *v1alpha1.Zone) error { return errBoom },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  zone(),
			},
			want: errors.Wrap(errBoom, errDeleteDNSZone),
		},
		"Successful": {
			e: &external{
				client: &MockZoneAPI{
					MockGet: func(_ context.Context, _ *v1alpha1.Zone) (dns.Zone, error) {
						return dns.Zone{
							ZoneProperties: &dns.ZoneProperties{},
						}, nil
					},
					MockDelete: func(_ context.Context, _ *v1alpha1.Zone) error { return nil },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  zone(),
			},
			want: nil,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := tc.e.Delete(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want, got, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): -want error, +got error:\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		eu  managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"ErrNotAMySQLServerConfiguration": {
			e: &external{},
			args: args{
				ctx: context.Background(),
			},
			want: want{
				err: errors.New(errNotDNSZone),
			},
		},
		"ErrUpdateServer": {
			e: &external{
				client: &MockZoneAPI{
					MockGet: func(_ context.Context, _ *v1alpha1.Zone) (dns.Zone, error) {
						return dns.Zone{
							ZoneProperties: &dns.ZoneProperties{},
						}, nil
					},
					MockCreateOrUpdate: func(_ context.Context, _ *v1alpha1.Zone) error { return errBoom },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  zone(),
			},
			want: want{
				err: errors.Wrap(errBoom, errUpdateDNSZone),
			},
		},
		"Successful": {
			e: &external{
				client: &MockZoneAPI{
					MockGet: func(_ context.Context, _ *v1alpha1.Zone) (dns.Zone, error) {
						return dns.Zone{
							ZoneProperties: &dns.ZoneProperties{},
						}, nil
					},
					MockCreateOrUpdate: func(_ context.Context, _ *v1alpha1.Zone) error { return nil },
				},
			},
			args: args{
				ctx: context.Background(),
				mg:  zone(),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			eu, err := tc.e.Update(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Update(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.eu, eu); diff != "" {
				t.Errorf("tc.e.Update(...): -want, +got:\n%s", diff)
			}
		})
	}
}
