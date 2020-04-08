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

package compute

import (
	"context"

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azurev1alpha3 "github.com/crossplane/provider-azure/apis/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/compute"
)

// Error strings.
const (
	errNewClient         = "cannot create new AKSCluster client"
	errGetProvider       = "cannot get Azure provider"
	errGetProviderSecret = "cannot get Azure provider Secret"
	errProviderSecretNil = "Azure provider does not have a secret reference"
	errGenPassword       = "cannot generate service principal secret"
	errNotAKSCluster     = "managed resource is not a AKSCluster"
	errCreateAKSCluster  = "cannot create AKSCluster"
	errGetAKSCluster     = "cannot get AKSCluster"
	errGetKubeConfig     = "cannot get AKSCluster kubeconfig"
	errDeleteAKSCluster  = "cannot delete AKSCluster"
)

// SetupAKSCluster adds a controller that reconciles AKSClusters.
func SetupAKSCluster(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.AKSClusterGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.AKSCluster{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.AKSClusterGroupVersionKind),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient(), newClientFn: compute.NewAggregateClient}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client      client.Client
	newClientFn func(credentials []byte) (compute.AKSClient, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	v, ok := mg.(*v1alpha3.AKSCluster)
	if !ok {
		return nil, errors.New(errNotAKSCluster)
	}

	p := &azurev1alpha3.Provider{}
	if err := c.client.Get(ctx, meta.NamespacedNameOf(v.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, errGetProvider)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	s := &corev1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.client.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, errGetProviderSecret)
	}
	aksClient, err := c.newClientFn(s.Data[p.Spec.CredentialsSecretRef.Key])
	return &external{kube: c.client, client: aksClient, newPasswordFn: password.Generate}, errors.Wrap(err, errNewClient)
}

type external struct {
	kube          client.Client
	client        compute.AKSClient
	newPasswordFn func() (password string, err error)
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha3.AKSCluster)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotAKSCluster)
	}

	c, err := e.client.GetManagedCluster(ctx, cr)
	if azure.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetAKSCluster)
	}

	cr.Status.ProviderID = to.String(c.ID)
	cr.Status.State = to.String(c.ProvisioningState)
	cr.Status.Endpoint = to.String(c.Fqdn)

	if cr.Status.State != "Succeeded" {
		// AKS clusters are always up to date because we can't yet update them.
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	}

	kubeconfig, err := e.client.GetKubeConfig(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetKubeConfig)
	}

	cd, err := connectionDetails(kubeconfig, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetAKSCluster)
	}

	cr.SetConditions(runtimev1alpha1.Available())
	resource.SetBindable(cr)

	// AKS clusters are always up to date because we can't yet update them.
	o := managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: cd,
	}
	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha3.AKSCluster)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotAKSCluster)
	}
	cr.SetConditions(runtimev1alpha1.Creating())
	secret, err := e.newPasswordFn()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGenPassword)
	}
	return managed.ExternalCreation{}, errors.Wrap(e.client.EnsureManagedCluster(ctx, cr, secret), errCreateAKSCluster)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(negz): Support updates.
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha3.AKSCluster)
	if !ok {
		return errors.New(errNotAKSCluster)
	}
	cr.SetConditions(runtimev1alpha1.Deleting())
	return errors.Wrap(e.client.DeleteManagedCluster(ctx, cr), errDeleteAKSCluster)
}

func connectionDetails(kubeconfig []byte, name string) (managed.ConnectionDetails, error) {
	kcfg, err := clientcmd.Load(kubeconfig)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse kubeconfig file")

	}
	kctx, ok := kcfg.Contexts[name]
	if !ok {
		return nil, errors.Errorf("context configuration is not found for cluster: %s", name)
	}
	cluster, ok := kcfg.Clusters[kctx.Cluster]
	if !ok {
		return nil, errors.Errorf("cluster configuration is not found: %s", kctx.Cluster)
	}
	auth, ok := kcfg.AuthInfos[kctx.AuthInfo]
	if !ok {
		return nil, errors.Errorf("auth-info configuration is not found: %s", kctx.AuthInfo)
	}

	return managed.ConnectionDetails{
		runtimev1alpha1.ResourceCredentialsSecretEndpointKey:   []byte(cluster.Server),
		runtimev1alpha1.ResourceCredentialsSecretCAKey:         cluster.CertificateAuthorityData,
		runtimev1alpha1.ResourceCredentialsSecretClientCertKey: auth.ClientCertificateData,
		runtimev1alpha1.ResourceCredentialsSecretClientKeyKey:  auth.ClientKeyData,
		runtimev1alpha1.ResourceCredentialsSecretKubeconfigKey: kubeconfig,
	}, nil
}
