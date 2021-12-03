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

package cache

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/Azure/azure-sdk-for-go/profiles/latest/redis/mgmt/redis/redisapi"
	"github.com/Azure/azure-sdk-for-go/services/redis/mgmt/2018-03-01/redis"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/cache/v1beta1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	redisclient "github.com/crossplane/provider-azure/pkg/clients/redis"
	"github.com/crossplane/provider-azure/pkg/clients/redis/fake"
)

const (
	name      = "cool-redis-53scf"
	namespace = "cool-namespace"

	connectionSecretName = "cool-connection-secret"
)

var (
	enableNonSSLPort = true
	subnetID         = "coolsubnet"
	staticIP         = "172.16.0.1"
	shardCount       = 3
	location         = "coolplace"
	minTLSVersion    = "1.1"
	tenantSettings   = map[string]string{"tenant1": "is-crazy"}
	hostName         = "108.8.8.1"
	port             = 6374
	primaryKey       = "secretpass"
	skuName          = "basic"
	skuFamily        = "C"
	skuCapacity      = 1
)

var (
	errorBoom          = errors.New("boom")
	redisConfiguration = map[string]string{"cool": "socool"}
)

type redisResourceModifier func(*v1beta1.Redis)

func withConditions(c ...xpv1.Condition) redisResourceModifier {
	return func(r *v1beta1.Redis) { r.Status.ConditionedStatus.Conditions = c }
}

func withProvisioningState(s string) redisResourceModifier {
	return func(r *v1beta1.Redis) { r.Status.AtProvider.ProvisioningState = s }
}

func withHostName(h string) redisResourceModifier {
	return func(r *v1beta1.Redis) { r.Status.AtProvider.HostName = h }
}

func withPort(p int) redisResourceModifier {
	return func(r *v1beta1.Redis) { r.Status.AtProvider.Port = p }
}

func instance(rm ...redisResourceModifier) *v1beta1.Redis {
	r := &v1beta1.Redis{
		Spec: v1beta1.RedisSpec{
			ResourceSpec: xpv1.ResourceSpec{
				WriteConnectionSecretToReference: &xpv1.SecretReference{
					Namespace: namespace,
					Name:      connectionSecretName,
				},
			},
			ForProvider: v1beta1.RedisParameters{
				Location:          location,
				ResourceGroupName: "group1",
				SKU: v1beta1.SKU{
					Name:     skuName,
					Capacity: skuCapacity,
					Family:   skuFamily,
				},
				Zones:              []string{"us-east1a", "us-east1b"},
				Tags:               map[string]string{"key1": "val1"},
				SubnetID:           &subnetID,
				StaticIP:           &staticIP,
				EnableNonSSLPort:   &enableNonSSLPort,
				RedisConfiguration: redisConfiguration,
				TenantSettings:     tenantSettings,
				ShardCount:         &shardCount,
				MinimumTLSVersion:  &minTLSVersion,
			},
		},
	}

	meta.SetExternalName(r, name)

	for _, m := range rm {
		m(r)
	}

	return r
}

var _ managed.ExternalClient = &external{}
var _ managed.ExternalConnecter = &connector{}

func TestObserve(t *testing.T) {
	type args struct {
		cr   *v1beta1.Redis
		r    redisapi.ClientAPI
		kube client.Client
	}
	type want struct {
		cr  *v1beta1.Redis
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				r: &fake.MockClient{
					MockGet: func(_ context.Context, resourceGroupName string, name string) (result redis.ResourceType, err error) {
						return redis.ResourceType{
							Properties: &redis.Properties{
								ProvisioningState: redis.Succeeded,
								HostName:          &hostName,
								Port:              azure.ToInt32(&port),
							},
						}, nil
					},
					MockListKeys: func(ctx context.Context, resourceGroupName string, name string) (result redis.AccessKeys, err error) {
						return redis.AccessKeys{
							PrimaryKey: azure.ToStringPtr(primaryKey),
						}, nil
					},
				},
			},
			want: want{
				cr: instance(
					withProvisioningState(redisclient.ProvisioningStateSucceeded),
					withHostName(hostName),
					withPort(port),
					withConditions(xpv1.Available()),
				),
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: false,
					ConnectionDetails: managed.ConnectionDetails{
						xpv1.ResourceCredentialsSecretEndpointKey: []byte(hostName),
						xpv1.ResourceCredentialsSecretPortKey:     []byte(strconv.Itoa(port)),
						xpv1.ResourceCredentialsSecretPasswordKey: []byte(primaryKey),
					},
				},
			},
		},
		"GetFailed": {
			args: args{
				cr: instance(),
				r: &fake.MockClient{
					MockGet: func(_ context.Context, resourceGroupName string, name string) (result redis.ResourceType, err error) {
						return redis.ResourceType{}, errorBoom
					},
				},
			},
			want: want{
				cr:  instance(),
				err: errors.Wrap(errorBoom, errGetFailed),
			},
		},
		"KubeUpdateFailed": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(errorBoom),
				},
				r: &fake.MockClient{
					MockGet: func(_ context.Context, resourceGroupName string, name string) (result redis.ResourceType, err error) {
						return redis.ResourceType{}, nil
					},
				},
			},
			want: want{
				cr:  instance(),
				err: errors.Wrap(errorBoom, errUpdateRedisCRFailed),
			},
		},
		"ListAccessKeysFailed": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				r: &fake.MockClient{
					MockGet: func(_ context.Context, resourceGroupName string, name string) (result redis.ResourceType, err error) {
						return redis.ResourceType{Properties: &redis.Properties{ProvisioningState: redis.Succeeded}}, nil
					},
					MockListKeys: func(_ context.Context, resourceGroupName string, name string) (result redis.AccessKeys, err error) {
						return redis.AccessKeys{}, errorBoom
					},
				},
			},
			want: want{
				cr: instance(
					withProvisioningState(redisclient.ProvisioningStateSucceeded),
				),
				err: errors.Wrap(errorBoom, errListAccessKeysFailed),
			},
		},
		"Creating": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				r: &fake.MockClient{
					MockGet: func(_ context.Context, resourceGroupName string, name string) (result redis.ResourceType, err error) {
						return redis.ResourceType{Properties: &redis.Properties{ProvisioningState: redis.Creating}}, nil
					},
					MockListKeys: func(_ context.Context, resourceGroupName string, name string) (result redis.AccessKeys, err error) {
						return redis.AccessKeys{}, nil
					},
				},
			},
			want: want{
				cr: instance(
					withProvisioningState(redisclient.ProvisioningStateCreating),
					withConditions(xpv1.Creating()),
				),
				o: managed.ExternalObservation{
					ResourceUpToDate: false,
					ResourceExists:   true,
				},
			},
		},
		"Deleting": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				r: &fake.MockClient{
					MockGet: func(_ context.Context, resourceGroupName string, name string) (result redis.ResourceType, err error) {
						return redis.ResourceType{Properties: &redis.Properties{ProvisioningState: redis.Deleting}}, nil
					},
					MockListKeys: func(_ context.Context, resourceGroupName string, name string) (result redis.AccessKeys, err error) {
						return redis.AccessKeys{}, nil
					},
				},
			},
			want: want{
				cr: instance(
					withProvisioningState(redisclient.ProvisioningStateDeleting),
					withConditions(xpv1.Deleting()),
				),
				o: managed.ExternalObservation{
					ResourceUpToDate: false,
					ResourceExists:   true,
				},
			},
		},
		"Unavailable": {
			args: args{
				cr: instance(),
				kube: &test.MockClient{
					MockUpdate: test.NewMockUpdateFn(nil),
				},
				r: &fake.MockClient{
					MockGet: func(_ context.Context, resourceGroupName string, name string) (result redis.ResourceType, err error) {
						return redis.ResourceType{Properties: &redis.Properties{ProvisioningState: redis.Failed}}, nil
					},
					MockListKeys: func(_ context.Context, resourceGroupName string, name string) (result redis.AccessKeys, err error) {
						return redis.AccessKeys{}, nil
					},
				},
			},
			want: want{
				cr: instance(
					withProvisioningState(redisclient.ProvisioningStateFailed),
					withConditions(xpv1.Unavailable()),
				),
				o: managed.ExternalObservation{
					ResourceUpToDate: false,
					ResourceExists:   true,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{
				kube:   tc.kube,
				client: tc.r,
			}
			o, err := e.Observe(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Observe(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Observe(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, o); diff != "" {
				t.Errorf("Observe(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestCreate(t *testing.T) {
	type args struct {
		cr *v1beta1.Redis
		r  redisapi.ClientAPI
	}
	type want struct {
		cr  *v1beta1.Redis
		o   managed.ExternalCreation
		err error
	}
	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr: instance(),
				r: &fake.MockClient{
					MockCreate: func(_ context.Context, resourceGroupName string, name string, parameters redis.CreateParameters) (result redis.CreateFuture, err error) {
						return redis.CreateFuture{}, nil
					},
				},
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Creating()),
				),
			},
		},
		"Failed": {
			args: args{
				cr: instance(),
				r: &fake.MockClient{
					MockCreate: func(_ context.Context, resourceGroupName string, name string, parameters redis.CreateParameters) (result redis.CreateFuture, err error) {
						return redis.CreateFuture{}, errorBoom
					},
				},
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Creating()),
				),
				err: errors.Wrap(errorBoom, errCreateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.r}

			c, err := e.Create(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Create(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Create(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, c); diff != "" {
				t.Errorf("Create(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type args struct {
		cr *v1beta1.Redis
		r  redisapi.ClientAPI
	}
	type want struct {
		cr  *v1beta1.Redis
		o   managed.ExternalUpdate
		err error
	}
	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr: instance(withProvisioningState(redisclient.ProvisioningStateSucceeded)),
				r: &fake.MockClient{
					MockGet: func(_ context.Context, _ string, _ string) (result redis.ResourceType, err error) {
						return redis.ResourceType{}, nil
					},
					MockUpdate: func(_ context.Context, resourceGroupName string, name string, parameters redis.UpdateParameters) (result redis.ResourceType, err error) {
						return redis.ResourceType{}, nil
					},
				},
			},
			want: want{
				cr: instance(withProvisioningState(redisclient.ProvisioningStateSucceeded)),
			},
		},
		"NotReady": {
			args: args{
				cr: instance(withProvisioningState(redisclient.ProvisioningStateFailed)),
			},
			want: want{
				cr: instance(withProvisioningState(redisclient.ProvisioningStateFailed)),
			},
		},
		"GetFailed": {
			args: args{
				cr: instance(withProvisioningState(redisclient.ProvisioningStateSucceeded)),
				r: &fake.MockClient{
					MockGet: func(_ context.Context, _ string, _ string) (result redis.ResourceType, err error) {
						return redis.ResourceType{}, errorBoom
					},
				},
			},
			want: want{
				cr:  instance(withProvisioningState(redisclient.ProvisioningStateSucceeded)),
				err: errors.Wrap(errorBoom, errGetFailed),
			},
		},
		"UpdateFailed": {
			args: args{
				cr: instance(withProvisioningState(redisclient.ProvisioningStateSucceeded)),
				r: &fake.MockClient{
					MockGet: func(_ context.Context, _ string, _ string) (result redis.ResourceType, err error) {
						return redis.ResourceType{Properties: &redis.Properties{ProvisioningState: redis.Succeeded}}, nil
					},
					MockUpdate: func(_ context.Context, resourceGroupName string, name string, parameters redis.UpdateParameters) (result redis.ResourceType, err error) {
						return redis.ResourceType{}, errorBoom
					},
				},
			},
			want: want{
				cr:  instance(withProvisioningState(redisclient.ProvisioningStateSucceeded)),
				err: errors.Wrap(errorBoom, errUpdateFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.r}

			c, err := e.Update(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.o, c); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type args struct {
		cr *v1beta1.Redis
		r  redisapi.ClientAPI
	}
	type want struct {
		cr  *v1beta1.Redis
		err error
	}
	cases := map[string]struct {
		args
		want
	}{
		"Successful": {
			args: args{
				cr: instance(),
				r: &fake.MockClient{
					MockDelete: func(_ context.Context, resourceGroupName string, name string) (result redis.DeleteFuture, err error) {
						return redis.DeleteFuture{}, nil
					},
				},
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"AlreadyDeleted": {
			args: args{
				cr: instance(),
				r: &fake.MockClient{
					MockDelete: func(_ context.Context, resourceGroupName string, name string) (result redis.DeleteFuture, err error) {
						return redis.DeleteFuture{}, autorest.DetailedError{StatusCode: http.StatusNotFound}
					},
				},
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Deleting()),
				),
			},
		},
		"AlreadyDeleting": {
			args: args{
				cr: instance(withProvisioningState(redisclient.ProvisioningStateDeleting)),
			},
			want: want{
				cr: instance(
					withConditions(xpv1.Deleting()),
					withProvisioningState(redisclient.ProvisioningStateDeleting),
				),
			},
		},
		"Failed": {
			args: args{
				cr: instance(),
				r: &fake.MockClient{
					MockDelete: func(_ context.Context, resourceGroupName string, name string) (result redis.DeleteFuture, err error) {
						return redis.DeleteFuture{}, errorBoom
					},
				},
			},
			want: want{
				cr:  instance(withConditions(xpv1.Deleting())),
				err: errors.Wrap(errorBoom, errDeleteFailed),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := external{client: tc.r}

			err := e.Delete(context.Background(), tc.args.cr)
			if diff := cmp.Diff(tc.want.cr, tc.args.cr); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Update(...): -want, +got\n%s", diff)
			}
		})
	}
}
