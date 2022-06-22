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

package recordset

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/services/dns/mgmt/2018-05-01/dns"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	dnsv1alpha1 "github.com/crossplane/provider-azure/apis/dns/v1alpha1"
	"github.com/crossplane/provider-azure/apis/v1alpha1"
	azureclients "github.com/crossplane/provider-azure/pkg/clients"
	dnsclients "github.com/crossplane/provider-azure/pkg/clients/dns"
	"github.com/crossplane/provider-azure/pkg/features"
)

// Error strings.
const (
	errNotDNSRecordSet    = "managed resource is not an DNS RecordSet"
	errCreateDNSRecordSet = "cannot create DNS RecordSet"
	errUpdateDNSRecordSet = "cannot update DNS RecordSet"
	errGetDNSRecordSet    = "cannot get DNS RecordSet"
	errDeleteDNSRecordSet = "cannot delete DNS RecordSet"
)

// Setup adds a controller that reconciles DNS RecordSets.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(dnsv1alpha1.RecordSetGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&dnsv1alpha1.RecordSet{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(dnsv1alpha1.RecordSetGroupVersionKind),
			managed.WithConnectionPublishers(),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azureclients.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := dns.NewRecordSetsClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{
		client: dnsclients.NewRecordSetClient(cl),
	}, nil
}

type external struct {
	client dnsclients.RecordSetAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	r, ok := mg.(*dnsv1alpha1.RecordSet)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDNSRecordSet)
	}

	az, err := e.client.Get(ctx, r)
	if azureclients.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetDNSRecordSet)
	}

	r.Spec.ForProvider.Metadata = azureclients.LateInitializeStringMap(r.Spec.ForProvider.Metadata, az.Metadata)

	dnsclients.UpdateRecordSetStatusFromAzure(r, az)

	if r.Status.AtProvider.ProvisioningState != dnsclients.RecordSetSuccessfulState {
		return managed.ExternalObservation{ResourceExists: true, ResourceUpToDate: true}, nil
	}

	r.SetConditions(xpv1.Available())

	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: dnsclients.RecordSetIsUpToDate(&r.Spec.ForProvider, az.RecordSetProperties),
	}

	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	r, ok := mg.(*dnsv1alpha1.RecordSet)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDNSRecordSet)
	}

	return managed.ExternalCreation{}, errors.Wrap(e.client.CreateOrUpdate(ctx, r), errCreateDNSRecordSet)
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	r, ok := mg.(*dnsv1alpha1.RecordSet)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDNSRecordSet)
	}

	az, err := e.client.Get(ctx, r)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetDNSRecordSet)
	}

	dnsclients.UpdateRecordSetStatusFromAzure(r, az)

	return managed.ExternalUpdate{}, errors.Wrap(e.client.CreateOrUpdate(ctx, r), errUpdateDNSRecordSet)
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	r, ok := mg.(*dnsv1alpha1.RecordSet)
	if !ok {
		return errors.New(errNotDNSRecordSet)
	}

	az, err := e.client.Get(ctx, r)
	if err != nil {
		return errors.Wrap(err, errGetDNSRecordSet)
	}

	dnsclients.UpdateRecordSetStatusFromAzure(r, az)

	if r.Status.AtProvider.ProvisioningState == dnsclients.RecordSetDeletingState {
		return nil
	}

	err = e.client.Delete(ctx, r)

	return errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errDeleteDNSRecordSet)
}
