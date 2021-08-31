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

package secret

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault"
	"github.com/Azure/azure-sdk-for-go/services/keyvault/v7.0/keyvault/keyvaultapi"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/keyvault/v1alpha1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	secretclients "github.com/crossplane/provider-azure/pkg/clients/keyvault/secret"
)

const (
	errNotSecret            = "the custom resource is not a Secret instance"
	errUpdateSecretCRFailed = "cannot update Secret custom resource instance"
	errCheckUpToDate        = "unable to determine if external resource is up to date"

	errConnectFailed        = "cannot connect to Azure API"
	errGetFailed            = "cannot get Key Vault Secret from Azure API"
	errListAccessKeysFailed = "cannot get access key list"
	errCreateFailed         = "cannot create the Key Vault Secret"
	errUpdateFailed         = "cannot update the Key Vault Secret"
	errDeleteFailed         = "cannot delete the Key Vault Secret"
)

// SetupSecret adds a controller that reconciles KeyVaultSecret resources.
func SetupSecret(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.KeyVaultSecretGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.KeyVaultSecret{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.KeyVaultSecretGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connector struct {
	kube client.Client
}

func (c connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	_, auth, err := azure.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errConnectFailed)
	}
	cl := keyvault.New()
	cl.Authorizer = auth
	return &external{kube: c.kube, client: cl}, nil
}

type external struct {
	kube   client.Client
	client keyvaultapi.BaseClientAPI
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.KeyVaultSecret)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSecret)
	}

	secret, err := c.client.GetSecret(ctx, cr.Spec.ForProvider.VaultBaseURL, cr.Spec.ForProvider.Name, "" /* latest */)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(resource.Ignore(azure.IsNotFound, err), errGetFailed)
	}

	lateInit := false
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	secretclients.LateInitialize(&cr.Spec.ForProvider, secret)
	if !cmp.Equal(currentSpec, &cr.Spec.ForProvider) {
		if err := c.kube.Update(ctx, cr); err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errUpdateSecretCRFailed)
		}
		lateInit = true
	}

	cr.Status.SetConditions(xpv1.Available())
	cr.Status.AtProvider = secretclients.GenerateObservation(secret)

	isUpToDate, err := secretclients.IsUpToDate(cr.Spec.ForProvider, &secret)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errCheckUpToDate)
	}

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        isUpToDate,
		ResourceLateInitialized: lateInit,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.KeyVaultSecret)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSecret)
	}

	cr.Status.SetConditions(xpv1.Creating())

	_, err := c.client.SetSecret(ctx, cr.Spec.ForProvider.VaultBaseURL, cr.Spec.ForProvider.Name, keyvault.SecretSetParameters{
		Value:            azure.ToStringPtr(cr.Spec.ForProvider.Value),
		Tags:             azure.ToStringPtrMap(cr.Spec.ForProvider.Tags),
		ContentType:      cr.Spec.ForProvider.ContentType,
		SecretAttributes: secretclients.GenerateAttributes(cr.Spec.ForProvider.SecretAttributes),
	})

	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
	}

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.KeyVaultSecret)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSecret)
	}
	_, err := c.client.SetSecret(ctx, cr.Spec.ForProvider.VaultBaseURL, cr.Spec.ForProvider.Name, keyvault.SecretSetParameters{
		Value:            azure.ToStringPtr(cr.Spec.ForProvider.Value),
		Tags:             azure.ToStringPtrMap(cr.Spec.ForProvider.Tags),
		ContentType:      cr.Spec.ForProvider.ContentType,
		SecretAttributes: secretclients.GenerateAttributes(cr.Spec.ForProvider.SecretAttributes),
	})

	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.KeyVaultSecret)
	if !ok {
		return errors.New(errNotSecret)
	}
	cr.Status.SetConditions(xpv1.Deleting())

	_, err := c.client.DeleteSecret(ctx, cr.Spec.ForProvider.VaultBaseURL, cr.Spec.ForProvider.Name)

	return err
}
