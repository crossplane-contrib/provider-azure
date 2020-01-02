/*
Copyright 2020 The Crossplane Authors.

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

package account

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"
	storagev1alpha1 "github.com/crossplaneio/crossplane/apis/storage/v1alpha1"

	"github.com/crossplaneio/stack-azure/apis/storage/v1alpha3"
)

var _ resource.ManagedConfigurator = resource.ManagedConfiguratorFn(ConfigureAccount)

func TestConfigureAccount(t *testing.T) {
	type args struct {
		ctx context.Context
		cm  resource.Claim
		cs  resource.Class
		mg  resource.Managed
	}

	type want struct {
		mg  resource.Managed
		err error
	}

	claimName := "coolaccount"
	claimUID := types.UID("definitely-a-uuid")
	classNamespace := "coolns"
	providerName := "coolprovider"
	bucketName := "coolbucket"
	bucketPrivate := storagev1alpha1.ACLPrivate

	cases := map[string]struct {
		args args
		want want
	}{
		"Successful": {
			args: args{
				cm: &storagev1alpha1.Bucket{
					ObjectMeta: metav1.ObjectMeta{Name: claimName, UID: claimUID},
					Spec: storagev1alpha1.BucketSpec{
						Name:          bucketName,
						PredefinedACL: &bucketPrivate,
					},
				},
				cs: &v1alpha3.AccountClass{
					ObjectMeta: metav1.ObjectMeta{Namespace: classNamespace},
					SpecTemplate: v1alpha3.AccountClassSpecTemplate{
						ClassSpecTemplate: runtimev1alpha1.ClassSpecTemplate{
							ProviderReference: &corev1.ObjectReference{Name: providerName},
							ReclaimPolicy:     runtimev1alpha1.ReclaimDelete,
						},
					},
				},
				mg: &v1alpha3.Account{},
			},
			want: want{
				mg: &v1alpha3.Account{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: classNamespace,
						Name:      claimName,
					},
					Spec: v1alpha3.AccountSpec{
						ResourceSpec: runtimev1alpha1.ResourceSpec{
							ReclaimPolicy:                    runtimev1alpha1.ReclaimDelete,
							WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{Name: string(claimUID)},
							ProviderReference:                &corev1.ObjectReference{Name: providerName},
						},
						AccountParameters: v1alpha3.AccountParameters{
							StorageAccountName: bucketName,
							StorageAccountSpec: &v1alpha3.StorageAccountSpec{},
						},
					},
				},
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			err := ConfigureAccount(tc.args.ctx, tc.args.cm, tc.args.cs, tc.args.mg)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("ConfigureAccount(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.mg, tc.args.mg, test.EquateConditions()); diff != "" {
				t.Errorf("ConfigureAccount(...) Managed: -want, +got:\n%s", diff)
			}
		})
	}
}
