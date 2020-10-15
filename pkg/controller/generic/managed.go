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

package generic

import (
	"context"
	"fmt"

	"github.com/Azure/k8s-infra/pkg/zips"
	"github.com/go-logr/logr"

	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/database/v1beta1"
	azure "github.com/crossplane/provider-azure/pkg/clients"
)

// Setup adds a controller that reconciles managed resources generated through the ASO codegen pipeline.
// note that unlike most Setup() methods, this takes a gvk in order to support putting arbitrary types
// under controller management.
func Setup(mgr ctrl.Manager, l logging.Logger, gvk schema.GroupVersionKind) error {
	name := managed.ControllerName(gvk.GroupKind().String())

	cl, ok := l.(logr.Logger)
	if !ok {
		return fmt.Errorf("Was not able to type assert Logger argument to logr.Logger for gvk=%s", gvk.String())
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1beta1.MySQLServer{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(gvk),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient(), logger: cl}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
	// using the concrete type to satisfy the requirement of zips.AzureTemplateClient
	logger logr.Logger
}

// Connect creates an ExternalClient instance capable of reconciling the given resource.Managed
// in the case of the generic Azure reconciler, this involves creating an Applier which is responsible
// for interacting with ARM to perform CRUD of managed resources.
func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azure.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	subscriptionID := creds[azure.CredentialsKeySubscriptionID]

	rawClient, err := zips.NewClient(auth)
	if err != nil {
		return nil, err
	}

	applier := zips.Applier(&zips.AzureTemplateClient{
		RawClient:      rawClient,
		Logger:         c.logger,
		SubscriptionID: subscriptionID,
	})
	return &external{kube: c.client, applier: applier}, nil
}

type external struct {
	kube    client.Client
	applier zips.Applier
}

// Observe is not yet implemented
func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	return managed.ExternalObservation{}, nil
}

// Create is not yet implemented
func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	return managed.ExternalCreation{}, nil
}

// Update is not yet implemented
func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	return managed.ExternalUpdate{}, nil
}

// Delete is not yet implemented
func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	return nil
}
