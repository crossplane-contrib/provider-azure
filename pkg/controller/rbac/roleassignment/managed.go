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

package roleassignment

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization"
	"github.com/Azure/azure-sdk-for-go/services/preview/authorization/mgmt/2018-01-01-preview/authorization/authorizationapi"
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
	"github.com/crossplane/crossplane-runtime/pkg/ratelimiter"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/rbac/v1alpha1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	azureclients "github.com/crossplane/provider-azure/pkg/clients"
)

// Error strings.
const (
	errNotRoleAssignment                = "managed resource is not a RoleAssignment"
	errCreateRoleAssignment             = "cannot create RoleAssignment"
	errRoleAssignmentUpdateNotSupported = "RoleAssignment updates not supported"
	errGetRoleAssignment                = "cannot get RoleAssignment"
	errDeleteRoleAssignment             = "cannot delete RoleAssignment"
)

// Setup adds a controller that reconciles RoleAssignment.
func Setup(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter, poll time.Duration) error {
	name := managed.ControllerName(v1alpha1.RoleAssignmentKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha1.RoleAssignment{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.RoleAssignmentGroupVersionKind),
			// Override default initializers in case to remove NewNameAsExternalName Initializer
			managed.WithInitializers(),
			managed.WithConnectionPublishers(),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azureclients.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	subID := creds[azure.CredentialsKeySubscriptionID]
	rac := authorization.NewRoleAssignmentsClient(subID)
	rac.Authorizer = auth
	_ = rac.AddToUserAgent(azure.UserAgent)
	return &external{c: rac, subID: subID}, nil
}

type external struct {
	c     authorizationapi.RoleAssignmentsClientAPI
	subID string
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	s, ok := mg.(*v1alpha1.RoleAssignment)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRoleAssignment)
	}
	name := ""
	filter := fmt.Sprintf("principalId eq '%s'", s.Spec.ForProvider.PrincipalID)
	l, err := e.c.ListForScopeComplete(ctx, s.Spec.ForProvider.Scope, filter)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetRoleAssignment)
	}
	if l.NotDone() {
		err := l.NextWithContext(ctx)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errGetRoleAssignment)
		}
		name = azure.ToString(l.Value().Name)
	} else {
		// Not exist
		return managed.ExternalObservation{}, nil
	}
	meta.SetExternalName(s, name)
	s.SetConditions(xpv1.Available())
	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	s, ok := mg.(*v1alpha1.RoleAssignment)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRoleAssignment)
	}
	p := authorization.RoleAssignmentCreateParameters{RoleAssignmentProperties: &authorization.RoleAssignmentProperties{
		RoleDefinitionID: azure.ToStringPtr(fmt.Sprintf("/subscriptions/%s%s", e.subID, s.Spec.ForProvider.RoleID)),
		PrincipalID:      azure.ToStringPtr(s.Spec.ForProvider.PrincipalID),
	}}
	uuidName, err := uuid.NewRandom()
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateRoleAssignment)
	}
	name := uuidName.String()
	_, err = e.c.Create(ctx, s.Spec.ForProvider.Scope, name, p)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateRoleAssignment)
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// RoleAssignments updates not supported by sdk
	return managed.ExternalUpdate{}, errors.New(errRoleAssignmentUpdateNotSupported)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	s, ok := mg.(*v1alpha1.RoleAssignment)
	if !ok {
		return errors.New(errNotRoleAssignment)
	}
	_, err := e.c.Delete(ctx, s.Spec.ForProvider.Scope, meta.GetExternalName(s))
	if azure.IsNotFound(err) {
		return nil
	}
	return errors.Wrap(err, errDeleteRoleAssignment)
}
