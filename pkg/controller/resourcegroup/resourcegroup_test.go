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

package resourcegroup

import (
	"context"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2018-05-01/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/v1alpha3"
	"github.com/crossplane/provider-azure/pkg/clients/resourcegroup"
	fakerg "github.com/crossplane/provider-azure/pkg/clients/resourcegroup/fake"
)

const (
	namespace = "cool-namespace"
	uid       = types.UID("definitely-a-uuid")
	name      = "cool-rg"
	location  = "coolplace"

	providerName       = "cool-azure"
	providerSecretName = "cool-azure-secret"
	providerSecretKey  = "credentials"
	providerSecretData = "definitelyjson"
)

type resourceGroupModifier func(*v1alpha3.ResourceGroup)

func withConditions(c ...runtimev1alpha1.Condition) resourceGroupModifier {
	return func(r *v1alpha3.ResourceGroup) { r.Status.ConditionedStatus.Conditions = c }
}

func resourceGrp(rm ...resourceGroupModifier) *v1alpha3.ResourceGroup {
	r := &v1alpha3.ResourceGroup{
		ObjectMeta: metav1.ObjectMeta{
			Name:       name,
			UID:        uid,
			Finalizers: []string{},
		},
		Spec: v1alpha3.ResourceGroupSpec{
			Location: location,
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
			},
		},
		Status: v1alpha3.ResourceGroupStatus{},
	}

	meta.SetExternalName(r, name)

	for _, m := range rm {
		m(r)
	}

	return r
}

func TestConnect(t *testing.T) {
	errorBoom := errors.New("boom")

	provider := v1alpha3.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: providerName},
		Spec: v1alpha3.ProviderSpec{
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

	providerSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data:       map[string][]byte{providerSecretKey: []byte(providerSecretData)},
	}

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		c   managed.ExternalClient
		err error
	}

	cases := map[string]struct {
		conn managed.ExternalConnecter
		args args
		want want
	}{
		"NotResourceGroup": {
			conn: &connecter{},
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotResourceGroup),
			},
		},
		"SuccessfulConnect": {
			conn: &connecter{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*v1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = providerSecret
					}
					return nil
				}},
				newClientFn: func(_ []byte) (resourcegroup.GroupsClient, error) {
					return &fakerg.MockClient{}, nil
				},
			},
			args: args{
				mg: resourceGrp(),
			},
			want: want{
				c: &external{client: &fakerg.MockClient{}},
			},
		},
		"FailedToGetProvider": {
			conn: &connecter{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					return kerrors.NewNotFound(schema.GroupResource{}, providerName)
				}},
				newClientFn: func(_ []byte) (resourcegroup.GroupsClient, error) {
					return &fakerg.MockClient{}, nil
				},
			},
			args: args{
				mg: resourceGrp(),
			},
			want: want{
				err: errors.WithStack(errors.Errorf("cannot get provider /%s:  \"%s\" not found", providerName, providerName)),
			},
		},
		"FailedToGetProviderSecret": {
			conn: &connecter{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*v1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return kerrors.NewNotFound(schema.GroupResource{}, providerSecretName)
					}
					return nil
				}},
				newClientFn: func(_ []byte) (resourcegroup.GroupsClient, error) {
					return &fakerg.MockClient{}, nil
				},
			},
			args: args{
				mg: resourceGrp(),
			},
			want: want{
				err: errors.WithStack(errors.Errorf("cannot get provider secret %s/%s:  \"%s\" not found", namespace, providerSecretName, providerSecretName)),
			},
		},
		"ProviderSecretNil": {
			conn: &connecter{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						nilSecretProvider := provider
						nilSecretProvider.SetCredentialsSecretReference(nil)
						*obj.(*v1alpha3.Provider) = nilSecretProvider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						return kerrors.NewNotFound(schema.GroupResource{}, providerSecretName)
					}
					return nil
				}},
				newClientFn: func(_ []byte) (resourcegroup.GroupsClient, error) {
					return &fakerg.MockClient{}, nil
				},
			},
			args: args{
				mg: resourceGrp(),
			},
			want: want{
				err: errors.New(errProviderSecretNil),
			},
		},
		"FailedToCreateAzureGroupsClient": {
			conn: &connecter{
				kube: &test.MockClient{MockGet: func(_ context.Context, key client.ObjectKey, obj runtime.Object) error {
					switch key {
					case client.ObjectKey{Name: providerName}:
						*obj.(*v1alpha3.Provider) = provider
					case client.ObjectKey{Namespace: namespace, Name: providerSecretName}:
						*obj.(*corev1.Secret) = providerSecret
					}
					return nil
				}},
				newClientFn: func(_ []byte) (resourcegroup.GroupsClient, error) { return nil, errorBoom },
			},
			args: args{
				mg: resourceGrp(),
			},
			want: want{
				c:   &external{},
				err: errors.Wrap(errorBoom, "cannot create new Azure Resource Group client"),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.conn.Connect(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.conn.Connect(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.c, got, cmp.AllowUnexported(external{})); diff != "" {
				t.Errorf("tc.conn.Connect(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestObserve(t *testing.T) {
	errBoom := errors.New("boom")

	type args struct {
		ctx context.Context
		mg  resource.Managed
	}

	type want struct {
		o   managed.ExternalObservation
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"NotResourceGroup": {
			e: &external{},
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotResourceGroup),
			},
		},
		"CheckExistenceError": {
			e: &external{
				client: &fakerg.MockClient{
					MockCheckExistence: func(_ context.Context, _ string) (result autorest.Response, err error) {
						return autorest.Response{}, errBoom
					},
				},
			},
			args: args{
				mg: resourceGrp(),
			},
			want: want{
				mg:  resourceGrp(),
				err: errors.Wrap(errBoom, errGetResourceGroup),
			},
		},
		"ResourceGroupNotFound": {
			e: &external{
				client: &fakerg.MockClient{
					MockCheckExistence: func(_ context.Context, _ string) (result autorest.Response, err error) {
						return autorest.Response{Response: &http.Response{StatusCode: http.StatusNotFound}}, nil
					},
				},
			},
			args: args{
				mg: resourceGrp(),
			},
			want: want{
				o:  managed.ExternalObservation{ResourceExists: false},
				mg: resourceGrp(),
			},
		},
		"Success": {
			e: &external{
				client: &fakerg.MockClient{
					MockCheckExistence: func(_ context.Context, _ string) (result autorest.Response, err error) {
						return autorest.Response{Response: &http.Response{StatusCode: http.StatusOK}}, nil
					},
				},
			},
			args: args{
				mg: resourceGrp(),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
				mg: resourceGrp(withConditions(runtimev1alpha1.Available())),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.e.Observe(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Observe(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("tc.e.Observe(...): -want managed, +got managed:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.o, got); diff != "" {
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
		c   managed.ExternalCreation
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"NotResourceGroup": {
			e: &external{},
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotResourceGroup),
			},
		},
		"CreateOrUpdateError": {
			e: &external{
				client: &fakerg.MockClient{
					MockCreateOrUpdate: func(_ context.Context, _ string, _ resources.Group) (result resources.Group, err error) {
						return resources.Group{}, errBoom
					},
				},
			},
			args: args{
				mg: resourceGrp(),
			},
			want: want{
				mg:  resourceGrp(withConditions(runtimev1alpha1.Creating())),
				err: errors.Wrap(errBoom, errCreateResourceGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, err := tc.e.Create(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Create(...): want error != got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg); diff != "" {
				t.Errorf("tc.e.Create(...): -want managed, +got managed:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.c, got); diff != "" {
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

	type want struct {
		mg  resource.Managed
		err error
	}

	cases := map[string]struct {
		e    managed.ExternalClient
		args args
		want want
	}{
		"NotResourceGroup": {
			e: &external{},
			args: args{
				mg: nil,
			},
			want: want{
				err: errors.New(errNotResourceGroup),
			},
		},
		"DeleteError": {
			e: &external{
				client: &fakerg.MockClient{
					MockDelete: func(_ context.Context, _ string) (result resources.GroupsDeleteFuture, err error) {
						return resources.GroupsDeleteFuture{}, errBoom
					},
				},
			},
			args: args{
				mg: resourceGrp(),
			},
			want: want{
				mg:  resourceGrp(withConditions(runtimev1alpha1.Deleting())),
				err: errors.Wrap(errBoom, errDeleteResourceGroup),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := tc.e.Delete(tc.args.ctx, tc.args.mg)

			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): -want error, +got error:\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateErrors()); diff != "" {
				t.Errorf("tc.e.Delete(...): -want, +got:\n%s", diff)
			}
		})
	}
}
