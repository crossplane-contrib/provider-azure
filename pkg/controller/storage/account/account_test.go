/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance With the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package account

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/crossplane/provider-azure/apis"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane/provider-azure/apis/storage/v1alpha3"
	v1alpha3test "github.com/crossplane/provider-azure/apis/storage/v1alpha3/test"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	azurestorage "github.com/crossplane/provider-azure/pkg/clients/storage"
	azurestoragefake "github.com/crossplane/provider-azure/pkg/clients/storage/fake"
)

func init() {
	_ = apis.AddToScheme(scheme.Scheme)
}

type MockAccountSecretupdater struct {
	MockUpdateSecret func(context.Context, *storage.Account) error
}

func (m *MockAccountSecretupdater) updatesecret(ctx context.Context, a *storage.Account) error {
	return m.MockUpdateSecret(ctx, a)
}

var _ secretupdater = &MockAccountSecretupdater{}

type MockAccountSyncbacker struct {
	MockSyncback func(context.Context, *storage.Account) (reconcile.Result, error)
}

func (m *MockAccountSyncbacker) syncback(ctx context.Context, a *storage.Account) (reconcile.Result, error) {
	return m.MockSyncback(ctx, a)
}

var _ syncbacker = &MockAccountSyncbacker{}

type MockAccountCreateUpdater struct {
	MockCreate func(context.Context) (reconcile.Result, error)
	MockUpdate func(context.Context, *storage.Account) (reconcile.Result, error)
}

func newMockAccountCreateUpdater() *MockAccountCreateUpdater {
	return &MockAccountCreateUpdater{
		MockUpdate: func(i context.Context, acct *storage.Account) (result reconcile.Result, e error) {
			return requeueOnSuccess, nil
		},
		MockCreate: func(i context.Context) (result reconcile.Result, e error) {
			return requeueOnSuccess, nil
		},
	}
}

func (m *MockAccountCreateUpdater) create(ctx context.Context) (reconcile.Result, error) {
	return m.MockCreate(ctx)
}

func (m *MockAccountCreateUpdater) update(ctx context.Context, a *storage.Account) (reconcile.Result, error) {
	return m.MockUpdate(ctx, a)
}

var _ createupdater = &MockAccountCreateUpdater{}

type MockAccountSyncDeleter struct {
	MockDelete func(context.Context) (reconcile.Result, error)
	MockSync   func(context.Context) (reconcile.Result, error)
}

func newMockAccountSyncDeleter() *MockAccountSyncDeleter {
	return &MockAccountSyncDeleter{
		MockSync: func(i context.Context) (result reconcile.Result, e error) {
			return requeueOnSuccess, nil
		},
		MockDelete: func(i context.Context) (result reconcile.Result, e error) {
			return result, nil
		},
	}
}

func (m *MockAccountSyncDeleter) delete(ctx context.Context) (reconcile.Result, error) {
	return m.MockDelete(ctx)
}

func (m *MockAccountSyncDeleter) sync(ctx context.Context) (reconcile.Result, error) {
	return m.MockSync(ctx)
}

var _ syncdeleter = &MockAccountSyncDeleter{}

type MockAccountHandleMaker struct {
	MockNew func(context.Context, *v1alpha3.Account) (syncdeleter, error)
}

func newMockAccountHandleMaker(rh syncdeleter, err error) *MockAccountHandleMaker {
	return &MockAccountHandleMaker{
		MockNew: func(i context.Context, bucket *v1alpha3.Account) (handler syncdeleter, e error) {
			return rh, err
		},
	}
}

func (m *MockAccountHandleMaker) newSyncdeleter(ctx context.Context, b *v1alpha3.Account) (syncdeleter, error) {
	return m.MockNew(ctx, b)
}

func newStorageAccountSpec() *v1alpha3.StorageAccountSpec {
	return v1alpha3.NewStorageAccountSpec(&storage.Account{})
}

func newStoragAccountSpecWithProperties() *v1alpha3.StorageAccountSpec {
	return v1alpha3.NewStorageAccountSpec(&storage.Account{AccountProperties: &storage.AccountProperties{}})
}

type storageAccount struct {
	*storage.Account
}

func newStorageAccount() *storageAccount {
	return &storageAccount{
		Account: &storage.Account{},
	}
}

func (sa *storageAccount) withAccountProperties(ap *storage.AccountProperties) *storageAccount {
	sa.Account.AccountProperties = ap
	return sa
}

type storageAccountProperties struct {
	*storage.AccountProperties
}

func newStorageAccountProperties() *storageAccountProperties {
	return &storageAccountProperties{
		AccountProperties: &storage.AccountProperties{},
	}
}

func (sap *storageAccountProperties) withProvisioningStage(ps storage.ProvisioningState) *storageAccountProperties {
	sap.ProvisioningState = ps
	return sap
}

const (
	testNamespace   = "default"
	testAccountName = "testAccount"
)

func TestReconciler_Reconcile(t *testing.T) {
	name := testAccountName
	key := types.NamespacedName{Name: name}
	req := reconcile.Request{NamespacedName: key}
	ctx := context.TODO()
	rsDone := reconcile.Result{}
	errBoom := errors.New("boom")

	type fields struct {
		client client.Client
		maker  syncdeleterMaker
	}
	type want struct {
		res  reconcile.Result
		err  error
		acct *v1alpha3.Account
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name:   "GetErrNotFound",
			fields: fields{client: fake.NewFakeClient(), maker: nil},
			want:   want{res: rsDone},
		},
		{
			name: "GetErrorOther",
			fields: fields{
				client: &test.MockClient{
					MockGet: func(context.Context, client.ObjectKey, runtime.Object) error {
						return errBoom
					},
				},
			},
			want: want{res: rsDone, err: errBoom},
		},
		{
			name: "AccountHandlerError",
			fields: fields{
				client: fake.NewFakeClient(v1alpha3test.NewMockAccount(name).WithFinalizer("foo.bar").Account),
				maker:  newMockAccountHandleMaker(nil, errBoom),
			},
			want: want{
				res: resultRequeue,
				acct: v1alpha3test.NewMockAccount(name).
					WithStatusConditions(
						runtimev1alpha1.ReconcileError(errBoom),
					).
					WithFinalizer("foo.bar").Account,
			},
		},
		{
			name: "ReconcileDelete",
			fields: fields{
				client: fake.NewFakeClient(v1alpha3test.NewMockAccount(name).
					WithDeleteTimestamp(metav1.NewTime(time.Now())).Account),
				maker: newMockAccountHandleMaker(newMockAccountSyncDeleter(), nil),
			},
			want: want{res: rsDone},
		},
		{
			name: "ReconcileSync",
			fields: fields{
				client: fake.NewFakeClient(v1alpha3test.NewMockAccount(name).Account),
				maker:  newMockAccountHandleMaker(newMockAccountSyncDeleter(), nil),
			},
			want: want{res: requeueOnSuccess},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Reconciler{
				Client:           tt.fields.client,
				syncdeleterMaker: tt.fields.maker,
				Initializer:      managed.NewNameAsExternalName(tt.fields.client),
				log:              logging.NewNopLogger(),
			}
			got, err := r.Reconcile(req)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("Reconciler.Reconcile(): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("Reconciler.Reconcile(): -want, +got:\n%s", diff)
			}
			if tt.want.acct != nil {
				b := &v1alpha3.Account{}
				if err := r.Get(ctx, key, b); err != nil {
					t.Errorf("Reconciler.Reconcile() account error: %s", err)
				}
				// NOTE(muvaf): Get call returns TypeMeta and ResourceVersion
				// that we are not interested. Since diff shows them, we have to
				// assign the right values.
				b.TypeMeta = tt.want.acct.TypeMeta
				tt.want.acct.ResourceVersion = b.ResourceVersion
				if diff := cmp.Diff(tt.want.acct, b, test.EquateConditions()); diff != "" {
					t.Errorf("Reconciler.Reconcile() account: -want, +got\n%s", diff)
				}
			}
		})
	}
}

func Test_syncdeleter_delete(t *testing.T) {
	ctx := context.TODO()
	bucketName := "test-account"
	errBoom := errors.New("boom")

	type fields struct {
		ao   azurestorage.AccountOperations
		cc   client.Client
		acct *v1alpha3.Account
	}
	type want struct {
		err  error
		res  reconcile.Result
		acct *v1alpha3.Account
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "RetainPolicy",
			fields: fields{
				acct: v1alpha3test.NewMockAccount(bucketName).WithSpecDeletionPolicy(runtimev1alpha1.DeletionOrphan).
					WithFinalizers([]string{finalizer, "test"}).Account,
				cc: &test.MockClient{
					MockUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error {
						return nil
					},
				},
			},
			want: want{
				err: nil,
				res: reconcile.Result{},
				acct: v1alpha3test.NewMockAccount(bucketName).
					WithSpecDeletionPolicy(runtimev1alpha1.DeletionOrphan).
					WithFinalizer("test").
					WithStatusConditions(runtimev1alpha1.Deleting()).
					Account,
			},
		},
		{
			name: "DeleteSuccessful",
			fields: fields{
				acct: v1alpha3test.NewMockAccount(bucketName).WithSpecDeletionPolicy(runtimev1alpha1.DeletionDelete).
					WithFinalizer(finalizer).Account,
				cc: &test.MockClient{
					MockUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error {
						return nil
					},
				},
				ao: azurestoragefake.NewMockAccountOperations(),
			},
			want: want{
				err: nil,
				res: reconcile.Result{},
				acct: v1alpha3test.NewMockAccount(bucketName).
					WithFinalizers([]string{}).
					WithSpecDeletionPolicy(runtimev1alpha1.DeletionDelete).
					WithStatusConditions(runtimev1alpha1.Deleting()).
					Account,
			},
		},
		{
			name: "DeleteFailed",
			fields: fields{
				acct: v1alpha3test.NewMockAccount(bucketName).WithSpecDeletionPolicy(runtimev1alpha1.DeletionDelete).
					WithFinalizer(finalizer).Account,
				cc: &test.MockClient{
					MockStatusUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error {
						return nil
					},
				},
				ao: &azurestoragefake.MockAccountOperations{
					MockDelete: func(ctx context.Context) error {
						return errBoom
					},
				},
			},
			want: want{
				err: nil,
				res: resultRequeue,
				acct: v1alpha3test.NewMockAccount(bucketName).WithSpecDeletionPolicy(runtimev1alpha1.DeletionDelete).
					WithFinalizer(finalizer).
					WithStatusConditions(runtimev1alpha1.Deleting(), runtimev1alpha1.ReconcileError(errBoom)).
					Account,
			},
		},
		{
			name: "DeleteNonExistent",
			fields: fields{
				acct: v1alpha3test.NewMockAccount(bucketName).WithSpecDeletionPolicy(runtimev1alpha1.DeletionDelete).
					WithFinalizer(finalizer).Account,
				cc: &test.MockClient{
					MockUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error { return nil },
				},
				ao: &azurestoragefake.MockAccountOperations{
					MockDelete: func(ctx context.Context) error {
						return autorest.DetailedError{
							StatusCode: http.StatusNotFound,
						}
					},
				},
			},
			want: want{
				err: nil,
				res: reconcile.Result{},
				acct: v1alpha3test.NewMockAccount(bucketName).
					WithFinalizers([]string{}).
					WithSpecDeletionPolicy(runtimev1alpha1.DeletionDelete).
					WithStatusConditions(runtimev1alpha1.Deleting()).
					Account,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := newAccountSyncDeleter(tt.fields.ao, tt.fields.cc, tt.fields.acct)
			got, err := bh.delete(ctx)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("accountSyncDeleter.delete(): -want error, +got error: \n%s", diff)
			}
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("accountSyncDeleter.delete(): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.acct, tt.fields.acct, test.EquateConditions()); diff != "" {
				t.Errorf("accountSyncDeleter.delete() account: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_syncdeleter_sync(t *testing.T) {
	ctx := context.TODO()
	name := testAccountName
	errBoom := errors.New("boom")

	type fields struct {
		ao   azurestorage.AccountOperations
		kube client.Client
		acct *v1alpha3.Account
	}
	type want struct {
		err  error
		res  reconcile.Result
		acct *v1alpha3.Account
	}
	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "AttrsError",
			fields: fields{
				ao: &azurestoragefake.MockAccountOperations{
					MockGet: func(i context.Context) (attrs *storage.Account, e error) {
						return nil, errBoom
					},
				},
				kube: &test.MockClient{
					MockCreate: func(ctx context.Context, obj runtime.Object, _ ...client.CreateOption) error { return nil },
					MockStatusUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error {
						return nil
					},
				},
				acct: v1alpha3test.NewMockAccount(name).WithUID("test-uid").Account,
			},
			want: want{
				res: resultRequeue,
				acct: v1alpha3test.NewMockAccount(name).
					WithUID("test-uid").
					WithStatusConditions(runtimev1alpha1.ReconcileError(errBoom)).
					Account,
			},
		},
		{
			name: "AttrsNotFoundCreate",
			fields: fields{
				kube: &test.MockClient{
					MockCreate: func(ctx context.Context, obj runtime.Object, _ ...client.CreateOption) error { return nil },
				},
				ao: &azurestoragefake.MockAccountOperations{
					MockGet: func(i context.Context) (attrs *storage.Account, e error) {
						return nil, autorest.DetailedError{
							StatusCode: http.StatusNotFound,
						}
					},
				},
				acct: v1alpha3test.NewMockAccount(name).WithUID("test-uid").Account,
			},
			want: want{
				res:  requeueOnSuccess,
				acct: v1alpha3test.NewMockAccount(name).WithUID("test-uid").Account,
			},
		},
		{
			name: "Update",
			fields: fields{
				kube: &test.MockClient{
					MockCreate: func(ctx context.Context, obj runtime.Object, _ ...client.CreateOption) error { return nil },
				},
				ao: &azurestoragefake.MockAccountOperations{
					MockGet: func(i context.Context) (attrs *storage.Account, e error) {
						return &storage.Account{}, nil
					},
				},
				acct: v1alpha3test.NewMockAccount(name).WithUID("test-uid").Account,
			},
			want: want{
				res:  requeueOnSuccess,
				acct: v1alpha3test.NewMockAccount(name).WithUID("test-uid").Account,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &accountSyncDeleter{
				createupdater:     newMockAccountCreateUpdater(),
				AccountOperations: tt.fields.ao,
				kube:              tt.fields.kube,
				acct:              tt.fields.acct,
			}

			got, err := bh.sync(ctx)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("accountSyncDeleter.delete() -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("accountSyncDeleter.delete(): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.acct, tt.fields.acct, test.EquateConditions()); diff != "" {
				t.Errorf("accountSyncDeleter.delete() account: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_createupdater_create(t *testing.T) {
	ctx := context.TODO()
	name := testAccountName
	errBoom := errors.New("boom")

	type fields struct {
		sb        syncbacker
		ao        azurestorage.AccountOperations
		kube      client.Client
		acct      *v1alpha3.Account
		projectID string
	}
	type want struct {
		err error
		res reconcile.Result
		obj *v1alpha3.Account
	}

	tests := []struct {
		name   string
		fields fields
		want   want
	}{
		{
			name: "CreateFailed",
			fields: fields{
				ao: &azurestoragefake.MockAccountOperations{
					MockCreate: func(ctx context.Context, params storage.AccountCreateParameters) (*storage.Account, error) {
						return nil, errBoom
					},
				},
				kube: &test.MockClient{
					MockStatusUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error {
						return nil
					},
				},
				acct: v1alpha3test.NewMockAccount(name).
					Account,
			},
			want: want{
				res: resultRequeue,
				obj: v1alpha3test.NewMockAccount(name).
					WithStatusConditions(runtimev1alpha1.Creating(), runtimev1alpha1.ReconcileError(errBoom)).
					WithFinalizer(finalizer).
					Account,
			},
		},
		{
			name: "CreateSuccessful",
			fields: fields{
				sb: &MockAccountSyncbacker{
					MockSyncback: func(ctx context.Context, a *storage.Account) (result reconcile.Result, e error) {
						return resultRequeue, errBoom
					},
				},
				ao:   azurestoragefake.NewMockAccountOperations(),
				kube: test.NewMockClient(),
				acct: v1alpha3test.NewMockAccount(name).
					WithUID("test-uid").
					Account,
			},
			want: want{
				err: errBoom,
				res: resultRequeue,
				obj: v1alpha3test.NewMockAccount(name).
					WithUID("test-uid").
					WithFinalizer(finalizer).
					WithStatusConditions(runtimev1alpha1.Creating()).
					Account,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &accountCreateUpdater{
				syncbacker:        tt.fields.sb,
				AccountOperations: tt.fields.ao,
				kube:              tt.fields.kube,
				acct:              tt.fields.acct,
				projectID:         tt.fields.projectID,
			}
			got, err := bh.create(ctx)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("accountCreateUpdater.create(): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("accountCreateUpdater.create(): -want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tt.want.obj, tt.fields.acct, test.EquateConditions()); diff != "" {
				t.Errorf("accountCreateUpdater.create() account: -want, +got:\n%s", diff)
			}
		})
	}
}

func Test_bucketCreateUpdater_update(t *testing.T) {
	ctx := context.TODO()
	name := testAccountName
	errBoom := errors.New("boom")

	type fields struct {
		sb   syncbacker
		ao   azurestorage.AccountOperations
		kube client.Client
		acct *v1alpha3.Account
	}
	type want struct {
		res  reconcile.Result
		err  error
		acct *v1alpha3.Account
	}
	tests := []struct {
		name   string
		fields fields
		attrs  *storage.Account
		want   want
	}{
		{
			name: "NotReady",
			attrs: newStorageAccount().
				withAccountProperties(newStorageAccountProperties().
					withProvisioningStage(storage.Creating).AccountProperties).Account,
			fields: fields{
				sb: &MockAccountSyncbacker{
					MockSyncback: func(ctx context.Context, a *storage.Account) (result reconcile.Result, e error) {
						return requeueOnSuccess, nil
					},
				},
			},
			want: want{
				res: requeueOnSuccess,
			},
		},
		{
			name: "NoChanges",
			attrs: &storage.Account{
				AccountProperties: &storage.AccountProperties{ProvisioningState: storage.Succeeded},
			},
			fields: fields{
				acct: v1alpha3test.NewMockAccount(name).
					WithSpecStorageAccountSpec(newStoragAccountSpecWithProperties()).
					Account,
				kube: test.NewMockClient(),
			},
			want: want{
				res: requeueOnSuccess,
				acct: v1alpha3test.NewMockAccount(name).
					WithSpecStorageAccountSpec(newStoragAccountSpecWithProperties()).
					WithStatusConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess()).
					Account,
			},
		},
		{
			name: "UpdateFailed",
			attrs: &storage.Account{
				AccountProperties: &storage.AccountProperties{ProvisioningState: storage.Succeeded},
				Location:          to.StringPtr("test-location"),
			},
			fields: fields{
				acct: v1alpha3test.NewMockAccount(name).WithSpecStorageAccountSpec(newStoragAccountSpecWithProperties()).Account,
				ao: &azurestoragefake.MockAccountOperations{
					MockUpdate: func(ctx context.Context, update storage.AccountUpdateParameters) (attrs *storage.Account, e error) {
						return nil, errBoom
					},
				},
				kube: &test.MockClient{
					MockStatusUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error { return nil },
				},
			},
			want: want{
				res: resultRequeue,
				acct: v1alpha3test.NewMockAccount(name).
					WithSpecStorageAccountSpec(newStoragAccountSpecWithProperties()).
					WithStatusConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileError(errBoom)).
					Account,
			},
		},
		{
			name: "UpdateSuccess",
			attrs: &storage.Account{
				AccountProperties: &storage.AccountProperties{ProvisioningState: storage.Succeeded},
				Location:          to.StringPtr("test-location"),
			},
			fields: fields{
				sb: &MockAccountSyncbacker{
					MockSyncback: func(ctx context.Context, a *storage.Account) (result reconcile.Result, e error) {
						return requeueOnSuccess, nil
					},
				},
				acct: v1alpha3test.NewMockAccount(name).WithSpecStorageAccountSpec(newStoragAccountSpecWithProperties()).Account,
				ao: &azurestoragefake.MockAccountOperations{
					MockUpdate: func(ctx context.Context, update storage.AccountUpdateParameters) (attrs *storage.Account, e error) {
						return &storage.Account{Location: to.StringPtr("test-location")}, nil
					},
				},
				kube: test.NewMockClient(),
			},
			want: want{
				res: requeueOnSuccess,
				acct: v1alpha3test.NewMockAccount(name).
					WithSpecStorageAccountSpec(newStoragAccountSpecWithProperties()).
					WithStatusConditions(runtimev1alpha1.Available()).
					Account,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bh := &accountCreateUpdater{
				syncbacker:        tt.fields.sb,
				AccountOperations: tt.fields.ao,
				kube:              tt.fields.kube,
				acct:              tt.fields.acct,
			}
			got, err := bh.update(ctx, tt.attrs)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("accountCreateUpdater.update() -want error, +got error:\n%s", diff)
				return
			}
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("accountCreateUpdater.update(): -want, +got:\n%s", diff)
				return
			}
			if diff := cmp.Diff(tt.want.acct, tt.fields.acct, test.EquateConditions()); diff != "" {
				t.Errorf("accountCreateUpdater.update() account: -want, +got:\n%s", diff)
				return
			}
		})
	}
}

func Test_accountSyncBacker_syncback(t *testing.T) {
	ctx := context.TODO()
	name := testAccountName
	errBoom := errors.New("boom")

	type fields struct {
		secretupdater secretupdater
		kube          client.Client
		acct          *v1alpha3.Account
	}
	type want struct {
		res  reconcile.Result
		err  error
		acct *v1alpha3.Account
	}
	tests := []struct {
		name   string
		fields fields
		acct   *storage.Account
		want   want
	}{
		{
			name: "UpdateDailed",
			fields: fields{
				secretupdater: &MockAccountSecretupdater{},
				acct:          v1alpha3test.NewMockAccount(name).Account,
				kube: &test.MockClient{
					MockUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error {
						return errBoom
					},
				},
			},
			acct: &storage.Account{},
			want: want{
				err:  errBoom,
				res:  resultRequeue,
				acct: v1alpha3test.NewMockAccount(name).WithSpecStorageAccountSpec(newStorageAccountSpec()).Account,
			},
		},
		{
			name: "ProvisionStatusIsNotSucceeded",
			fields: fields{
				acct: v1alpha3test.NewMockAccount(name).Account,
				kube: test.NewMockClient(),
			},
			acct: newStorageAccount().
				withAccountProperties(newStorageAccountProperties().
					withProvisioningStage(storage.Creating).AccountProperties).Account,
			want: want{
				res: requeueOnWait,
				acct: v1alpha3test.NewMockAccount(name).
					WithSpecStatusFromProperties(&storage.AccountProperties{ProvisioningState: storage.Creating}).
					WithStatusConditions(runtimev1alpha1.ReconcileSuccess()).
					Account,
			},
		},
		{
			name: "UpdateSecretFailed",
			fields: fields{
				secretupdater: &MockAccountSecretupdater{
					MockUpdateSecret: func(ctx context.Context, a *storage.Account) error {
						return errBoom
					},
				},
				acct: v1alpha3test.NewMockAccount(name).Account,
				kube: test.NewMockClient(),
			},
			acct: &storage.Account{AccountProperties: &storage.AccountProperties{ProvisioningState: storage.Succeeded}},
			want: want{
				res: resultRequeue,
				acct: v1alpha3test.NewMockAccount(name).
					WithSpecStatusFromProperties(&storage.AccountProperties{ProvisioningState: storage.Succeeded}).
					WithStatusConditions(runtimev1alpha1.ReconcileError(errBoom)).Account,
			},
		},
		{
			name: "Success",
			fields: fields{
				secretupdater: &MockAccountSecretupdater{
					MockUpdateSecret: func(ctx context.Context, a *storage.Account) error { return nil },
				},
				acct: v1alpha3test.NewMockAccount(name).
					WithSpecStorageAccountSpec(v1alpha3.NewStorageAccountSpec(&storage.Account{})).
					Account,
				kube: test.NewMockClient(),
			},
			acct: &storage.Account{AccountProperties: &storage.AccountProperties{ProvisioningState: storage.Succeeded}},
			want: want{
				res: requeueOnSuccess,
				acct: v1alpha3test.NewMockAccount(name).
					WithSpecStatusFromProperties(&storage.AccountProperties{ProvisioningState: storage.Succeeded}).
					WithStatusConditions(runtimev1alpha1.ReconcileSuccess()).
					Account,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			acu := &accountSyncbacker{
				secretupdater: tt.fields.secretupdater,
				kube:          tt.fields.kube,
				acct:          tt.fields.acct,
			}
			got, err := acu.syncback(ctx, tt.acct)
			if diff := cmp.Diff(tt.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("accountSyncBackSecretUpdater.syncback() -want error, +got error:\n%s", diff)
				return
			}
			if diff := cmp.Diff(tt.want.res, got); diff != "" {
				t.Errorf("accountSyncBackSecretUpdater.syncback(): -want, +got:\n%s", diff)
				return
			}
			if diff := cmp.Diff(tt.want.acct, tt.fields.acct, test.EquateConditions()); diff != "" {
				t.Errorf("accountSyncBackSecretUpdater.syncback() account: -want, +got:\n%s", diff)
				return
			}
		})
	}
}

func Test_accountSecretUpdater_updatesecret(t *testing.T) {
	ctx := context.TODO()
	ns := testNamespace
	name := testAccountName
	csName := "connectionsecret"

	type fields struct {
		ops  azurestorage.AccountOperations
		kube client.Client
		acct *v1alpha3.Account
	}

	tests := []struct {
		name    string
		fields  fields
		acct    *storage.Account
		wantErr error
	}{
		{
			name: "FailedListKeys",
			fields: fields{
				ops: &azurestoragefake.MockAccountOperations{
					MockListKeys: func(ctx context.Context) (keys []storage.AccountKey, e error) {
						return nil, errors.New("test-list-keys-error")
					},
				},
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
						return kerrors.NewNotFound(schema.GroupResource{Group: azurev1alpha3.Group, Resource: "secret"}, name)
					},
				},
				acct: v1alpha3test.NewMockAccount(name).WithSpecWriteConnectionSecretToReference(ns, csName).Account,
			},
			acct: &storage.Account{
				AccountProperties: &storage.AccountProperties{
					PrimaryEndpoints: &storage.Endpoints{
						Blob: to.StringPtr("test-blob-endpoint"),
					},
				},
			},
			wantErr: errors.Wrapf(errors.New("test-list-keys-error"), "failed to list account keys"),
		},
		{
			name: "AccountKeysListEmpty",
			fields: fields{
				ops: &azurestoragefake.MockAccountOperations{
					MockListKeys: func(ctx context.Context) (keys []storage.AccountKey, e error) {
						return []storage.AccountKey{}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
						return kerrors.NewNotFound(schema.GroupResource{Group: azurev1alpha3.Group, Resource: "secret"}, name)
					},
				},
				acct: v1alpha3test.NewMockAccount(name).WithSpecWriteConnectionSecretToReference(ns, csName).Account,
			},
			acct: &storage.Account{
				AccountProperties: &storage.AccountProperties{
					PrimaryEndpoints: &storage.Endpoints{
						Blob: to.StringPtr("test-blob-endpoint"),
					},
				},
			},
			wantErr: errors.New("account keys are empty"),
		},
		{
			name: "CreateNewSecret",
			fields: fields{
				ops: &azurestoragefake.MockAccountOperations{
					MockListKeys: func(ctx context.Context) (keys []storage.AccountKey, e error) {
						return []storage.AccountKey{
							{
								KeyName: to.StringPtr("test-key"),
								Value:   to.StringPtr("test-value"),
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
						return kerrors.NewNotFound(schema.GroupResource{Group: azurev1alpha3.Group, Resource: "secret"}, name)
					},
					MockCreate: func(ctx context.Context, obj runtime.Object, _ ...client.CreateOption) error {
						return nil
					},
				},
				acct: v1alpha3test.NewMockAccount(name).WithSpecWriteConnectionSecretToReference(ns, csName).Account,
			},
			acct: &storage.Account{
				AccountProperties: &storage.AccountProperties{
					PrimaryEndpoints: &storage.Endpoints{
						Blob: to.StringPtr("test-blob-endpoint"),
					},
				},
			},
		},
		{
			name: "CreateNewSecretFailed",
			fields: fields{
				ops: &azurestoragefake.MockAccountOperations{
					MockListKeys: func(ctx context.Context) (keys []storage.AccountKey, e error) {
						return []storage.AccountKey{
							{
								KeyName: to.StringPtr("test-key"),
								Value:   to.StringPtr("test-value"),
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
						return kerrors.NewNotFound(schema.GroupResource{Group: azurev1alpha3.Group, Resource: "secret"}, name)
					},
					MockCreate: func(ctx context.Context, obj runtime.Object, _ ...client.CreateOption) error {
						return errors.New("test-create-secret-error")
					},
				},
				acct: v1alpha3test.NewMockAccount(name).WithSpecWriteConnectionSecretToReference(ns, csName).Account,
			},
			acct: &storage.Account{
				AccountProperties: &storage.AccountProperties{
					PrimaryEndpoints: &storage.Endpoints{
						Blob: to.StringPtr("test-blob-endpoint"),
					},
				},
			},
			wantErr: errors.Wrapf(errors.New("test-create-secret-error"), "failed to create secret: %s/%s", ns, csName),
		},
		{
			name: "UpdateExistingSecret",
			fields: fields{
				ops: &azurestoragefake.MockAccountOperations{
					MockListKeys: func(ctx context.Context) (keys []storage.AccountKey, e error) {
						return []storage.AccountKey{
							{
								KeyName: to.StringPtr("test-key"),
								Value:   to.StringPtr("test-value"),
							},
						}, nil
					},
				},
				kube: &test.MockClient{
					MockGet: func(ctx context.Context, key client.ObjectKey, obj runtime.Object) error {
						return nil
					},
					MockCreate: func(ctx context.Context, obj runtime.Object, _ ...client.CreateOption) error {
						return kerrors.NewAlreadyExists(schema.GroupResource{Group: azurev1alpha3.Group, Resource: "secret"}, name)
					},
					MockUpdate: func(ctx context.Context, obj runtime.Object, _ ...client.UpdateOption) error {
						return nil
					},
				},
				acct: v1alpha3test.NewMockAccount(name).WithSpecWriteConnectionSecretToReference(ns, csName).Account,
			},
			acct: &storage.Account{
				AccountProperties: &storage.AccountProperties{
					PrimaryEndpoints: &storage.Endpoints{
						Blob: to.StringPtr("test-blob-endpoint"),
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asu := &accountSecretUpdater{
				AccountOperations: tt.fields.ops,
				acct:              tt.fields.acct,
				kube:              tt.fields.kube,
			}
			err := asu.updatesecret(ctx, tt.acct)
			if diff := cmp.Diff(tt.wantErr, err, test.EquateErrors()); diff != "" {
				t.Errorf("accountSyncBackSecretUpdater.syncback() -want error, +got error:\n%s", diff)
			}
		})
	}
}
