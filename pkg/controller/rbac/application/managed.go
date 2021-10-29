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

package application

import (
	"context"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac"
	"github.com/Azure/azure-sdk-for-go/services/graphrbac/1.6/graphrbac/graphrbacapi"
	"github.com/Azure/go-autorest/autorest/date"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/password"
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/rbac/v1alpha1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	azureclients "github.com/crossplane/provider-azure/pkg/clients"
)

// Error strings.
const (
	errNotApplication    = "managed resource is not an Application"
	errCreateApplication = "cannot create Application"
	errGetApplication    = "cannot get Application"
	errDeleteApplication = "cannot delete Application"
)

// Setup adds a controller that reconciles Application.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.ApplicationKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.Application{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.ApplicationGroupVersionKind),
			// Override default initializers in case to remove NewNameAsExternalName Initializer
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(poll),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, _, err := azureclients.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	ta, err := azure.NewADGraphResourceIDAuthorizer(creds)
	if err != nil {
		return nil, err
	}
	ac := graphrbac.NewApplicationsClient(creds[azure.CredentialsKeyTenantID])
	ac.Authorizer = ta
	return &external{c: ac}, nil
}

type external struct {
	c graphrbacapi.ApplicationsClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	s, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotApplication)
	}
	if meta.GetExternalName(s) == "" {
		return managed.ExternalObservation{}, nil
	}

	az, err := e.c.Get(ctx, meta.GetExternalName(s))
	if azureclients.IsNotFound(err) {
		return managed.ExternalObservation{}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetApplication)
	}

	s.Status.ApplicationID = azure.ToString(az.AppID)
	s.SetConditions(xpv1.Available())
	// TODO: drift detection
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	s, ok := mg.(*v1alpha1.Application)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotApplication)
	}
	pw, err := password.Generate()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateApplication)
	}
	keyID, err := uuid.NewRandom()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateApplication)
	}
	pc := newPasswordCredential(keyID.String(), pw)
	p := graphrbac.ApplicationCreateParameters{
		AvailableToOtherTenants: s.Spec.ForProvider.AvailableToOtherTenants,
		DisplayName:             s.Spec.ForProvider.DisplayName,
		Homepage:                s.Spec.ForProvider.Homepage,
		IdentifierUris:          azure.ToStringArrayPtr(s.Spec.ForProvider.IdentifierURIs),
		PasswordCredentials:     &[]graphrbac.PasswordCredential{pc},
	}
	rsp, err := e.c.Create(ctx, p)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateApplication)
	}
	meta.SetExternalName(s, azure.ToString(rsp.ObjectID))
	return managed.ExternalCreation{
		ConnectionDetails: map[string][]byte{
			xpv1.ResourceCredentialsSecretPasswordKey: []byte(pw),
			"keyID": []byte(keyID.String()),
		},
	}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// TODO: support updates
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	s, ok := mg.(*v1alpha1.Application)
	if !ok {
		return errors.New(errNotApplication)
	}
	_, err := e.c.Delete(ctx, meta.GetExternalName(s))
	if azureclients.IsNotFound(err) {
		return nil
	}
	return errors.Wrap(resource.IgnoreNotFound(err), errDeleteApplication)
}

const (
	// TODO: creds autoupdate
	appCredsValidYears = 5
)

func newPasswordCredential(keyID, secret string) graphrbac.PasswordCredential {
	return graphrbac.PasswordCredential{
		StartDate: &date.Time{Time: time.Now()},
		EndDate:   &date.Time{Time: time.Now().AddDate(appCredsValidYears, 0, 0)},
		KeyID:     to.StringPtr(keyID),
		Value:     to.StringPtr(secret),
	}
}
