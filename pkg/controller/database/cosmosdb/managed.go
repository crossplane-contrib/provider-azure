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

package cosmosdb

import (
	"context"
	"net/http"

	"github.com/Azure/azure-sdk-for-go/services/cosmos-db/mgmt/2015-04-08/documentdb"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane/provider-azure/apis/database/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/database/cosmosdb"
)

// Error strings
const (
	errNotNoSQLAccount    = "managed resource is not a Database Account"
	errCreateNoSQLAccount = "cannot create Database Account"
	errGetNoSQLAccount    = "cannot get Database Account"
	errDeleteNoSQLAccount = "cannot delete Database Account"
)

// Setup adds a controller that reconciles NoSQLAccount.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.CosmosDBAccountGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.CosmosDBAccount{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.CosmosDBAccountGroupVersionKind),
			managed.WithConnectionPublishers(),
			managed.WithExternalConnecter(&connecter{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	kube client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azure.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, err
	}
	cl := documentdb.NewDatabaseAccountsClient(creds[azure.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{kube: c.kube, client: cl}, nil
}

// external is a createsyncdeleter using the Azure API.
type external struct {
	kube   client.Client
	client cosmosdb.AccountClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	r, ok := mg.(*v1alpha3.CosmosDBAccount)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotNoSQLAccount)
	}

	res, err := e.client.CheckNameExists(ctx, meta.GetExternalName(r))
	if res.IsHTTPStatus(http.StatusNotFound) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetNoSQLAccount)
	}

	account, err := e.client.Get(ctx, r.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(r))
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetNoSQLAccount)
	}
	cosmosdb.UpdateCosmosDBAccountObservation(&r.Status, account)

	switch r.Status.AtProvider.State {
	case "Succeeded":
		r.SetConditions(runtimev1alpha1.Available())
	default:
		r.SetConditions(runtimev1alpha1.Unavailable())
	}
	resourceUpToDate := cosmosdb.CheckEqualDatabaseProperties(r.Spec.ForProvider.Properties, account)
	return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: resourceUpToDate}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	r, ok := mg.(*v1alpha3.CosmosDBAccount)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotNoSQLAccount)
	}

	r.Status.SetConditions(runtimev1alpha1.Creating())
	_, err := e.client.CreateOrUpdate(ctx,
		r.Spec.ForProvider.ResourceGroupName,
		meta.GetExternalName(r),
		cosmosdb.ToDatabaseAccountCreateOrUpdate(&r.Spec))
	// TODO(artursouza): handle secrets.
	return managed.ExternalCreation{}, errors.Wrap(err, errCreateNoSQLAccount)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	_, err := e.Create(ctx, mg)
	return managed.ExternalUpdate{}, err
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	r, ok := mg.(*v1alpha3.CosmosDBAccount)
	if !ok {
		return errors.New(errNotNoSQLAccount)
	}

	r.Status.SetConditions(runtimev1alpha1.Deleting())
	_, err := e.client.Delete(ctx, r.Spec.ForProvider.ResourceGroupName, meta.GetExternalName(r))
	return errors.Wrap(err, errDeleteNoSQLAccount)
}
