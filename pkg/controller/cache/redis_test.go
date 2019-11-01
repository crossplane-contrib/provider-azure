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
	"testing"

	"github.com/crossplaneio/crossplane-runtime/pkg/meta"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	"github.com/crossplaneio/stack-azure/apis/cache/v1alpha3"
	azurev1alpha3 "github.com/crossplaneio/stack-azure/apis/v1alpha3"
)

const (
	name                   = "cool-redis-53scf"
	namespace              = "cool-namespace"
	uid                    = types.UID("definitely-a-uuid")
	redisResourceGroupName = "coolgroup"
	location               = "coolplace"
	subscription           = "totally-a-uuid"
	qualifiedName          = "/subscriptions/" + subscription + "/redisResourceGroups/" + redisResourceGroupName + "/providers/Microsoft.Cache/Redis/" + name
	host                   = "172.16.0.1"
	port                   = 6379
	sslPort                = 6380
	enableNonSSLPort       = true
	shardCount             = 3
	skuName                = v1alpha3.SKUNameBasic
	skuFamily              = v1alpha3.SKUFamilyC
	skuCapacity            = 1

	primaryAccessKey = "sosecret"

	providerName       = "cool-azure"
	providerSecretName = "cool-azure-secret"
	providerSecretKey  = "credentials"
	providerSecretData = "definitelyjson"

	connectionSecretName = "cool-connection-secret"
)

var (
	ctx                = context.Background()
	errorBoom          = errors.New("boom")
	redisConfiguration = map[string]string{"cool": "socool"}

	provider = azurev1alpha3.Provider{
		ObjectMeta: metav1.ObjectMeta{Name: providerName},
		Spec: azurev1alpha3.ProviderSpec{
			Secret: runtimev1alpha1.SecretKeySelector{
				SecretReference: runtimev1alpha1.SecretReference{
					Namespace: namespace,
					Name:      providerSecretName,
				},
				Key: providerSecretKey,
			},
		},
	}

	providerSecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: providerSecretName},
		Data:       map[string][]byte{providerSecretKey: []byte(providerSecretData)},
	}
)

type redisResourceModifier func(*v1alpha3.Redis)

func withConditions(c ...runtimev1alpha1.Condition) redisResourceModifier {
	return func(r *v1alpha3.Redis) { r.Status.ConditionedStatus.Conditions = c }
}

func withBindingPhase(p runtimev1alpha1.BindingPhase) redisResourceModifier {
	return func(r *v1alpha3.Redis) { r.Status.SetBindingPhase(p) }
}

func withState(s string) redisResourceModifier {
	return func(r *v1alpha3.Redis) { r.Status.State = s }
}

func instance(rm ...redisResourceModifier) *v1alpha3.Redis {
	r := &v1alpha3.Redis{
		ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{
				meta.ExternalNameAnnotationKey: name,
			},
		},
		Spec: v1alpha3.RedisSpec{
			ResourceSpec: runtimev1alpha1.ResourceSpec{
				ProviderReference: &corev1.ObjectReference{Name: providerName},
				WriteConnectionSecretToReference: &runtimev1alpha1.SecretReference{
					Namespace: namespace,
					Name:      connectionSecretName,
				},
			},
			RedisParameters: v1alpha3.RedisParameters{
				ResourceGroupName:  redisResourceGroupName,
				Location:           location,
				RedisConfiguration: redisConfiguration,
				EnableNonSSLPort:   enableNonSSLPort,
				ShardCount:         shardCount,
				SKU: v1alpha3.SKU{
					Name:     skuName,
					Family:   skuFamily,
					Capacity: skuCapacity,
				},
			},
		},
		Status: v1alpha3.RedisStatus{
			Endpoint:   host,
			Port:       port,
			ProviderID: qualifiedName,
		},
	}

	for _, m := range rm {
		m(r)
	}

	return r
}

var _ resource.ExternalClient = &external{}
var _ resource.ExternalConnecter = &connector{}

func TestConnect(t *testing.T) {

}

func TestObserve(t *testing.T) {
	//type args struct {
	//	cr *v1alpha3.Redis
	//	r redisapi.ClientAPI
	//	kube client.Client
	//}
	//type want struct {
	//	cr *v1alpha3.Redis
	//}
	//
	//cases := map[string]struct {
	//	args
	//	want
	//}{
	//	""
	//}
}

func TestCreate(t *testing.T) {

}

func TestUpdate(t *testing.T) {

}

func TestDelete(t *testing.T) {

}
