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

	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/v1alpha3"
	"github.com/crossplane/provider-azure/pkg/clients/resourcegroup"
)

// Error strings
const (
	errProviderSecretNil   = "provider does not have a secret reference"
	errNotResourceGroup    = "managed resource is not an ResourceGroup"
	errCreateResourceGroup = "cannot create ResourceGroup"
	errCheckResourceGroup  = "cannot check existence of ResourceGroup"
	errGetResourceGroup    = "cannot get ResourceGroup"
	errDeleteResourceGroup = "cannot delete ResourceGroup"
)

// Setup adds a controller that reconciles ResourceGroups.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.ResourceGroupGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.ResourceGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.ResourceGroupGroupVersionKind),
			managed.WithConnectionPublishers(),
			managed.WithExternalConnecter(&connecter{kube: mgr.GetClient(), newClientFn: resourcegroup.NewClient}),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	kube        client.Client
	newClientFn func(creds []byte) (resourcegroup.GroupsClient, error)
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	r, ok := mg.(*v1alpha3.ResourceGroup)
	if !ok {
		return nil, errors.New(errNotResourceGroup)
	}

	p := &v1alpha3.Provider{}
	n := meta.NamespacedNameOf(r.Spec.ProviderReference)
	if err := c.kube.Get(ctx, n, p); err != nil {
		return nil, errors.Wrapf(err, "cannot get provider %s", n)
	}

	if p.GetCredentialsSecretReference() == nil {
		return nil, errors.New(errProviderSecretNil)
	}

	s := &corev1.Secret{}
	n = types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := c.kube.Get(ctx, n, s); err != nil {
		return nil, errors.Wrapf(err, "cannot get provider secret %s", n)
	}

	client, err := c.newClientFn(s.Data[p.Spec.CredentialsSecretRef.Key])
	return &external{client: client}, errors.Wrap(err, "cannot create new Azure Resource Group client")
}

// external is a createsyncdeleter using the Azure Groups API.
type external struct {
	client resourcegroup.GroupsClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	r, ok := mg.(*v1alpha3.ResourceGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotResourceGroup)
	}

	res, err := e.client.CheckExistence(ctx, meta.GetExternalName(r))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckResourceGroup)
	}

	if res.Response.StatusCode == http.StatusNotFound {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	g, err := e.client.Get(ctx, meta.GetExternalName(r))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetResourceGroup)
	}
	if g.Properties != nil {
		r.Status.ProvisioningState = v1alpha3.ProvisioningState(to.String(g.Properties.ProvisioningState))
	}

	r.SetConditions(runtimev1alpha1.Available())
	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	r, ok := mg.(*v1alpha3.ResourceGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotResourceGroup)
	}

	r.Status.SetConditions(runtimev1alpha1.Creating())
	_, err := e.client.CreateOrUpdate(ctx, meta.GetExternalName(r), resourcegroup.NewParameters(r))
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateResourceGroup)
}

func (e *external) Update(_ context.Context, _ resource.Managed) (managed.ExternalUpdate, error) {
	// TODO(negz): Support updates, if applicable.
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	r, ok := mg.(*v1alpha3.ResourceGroup)
	if !ok {
		return errors.New(errNotResourceGroup)
	}

	// Calling delete on a resource group that is already deleting will succeed,
	// but seems to prolong the deletion process, potentially resulting in a
	// resource group that never actually gets deleted.
	if r.Status.ProvisioningState == v1alpha3.ProvisioningStateDeleting {
		return nil
	}

	r.Status.SetConditions(runtimev1alpha1.Deleting())
	_, err := e.client.Delete(ctx, meta.GetExternalName(r))
	return errors.Wrap(err, errDeleteResourceGroup)
}
