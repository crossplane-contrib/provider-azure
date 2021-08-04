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

package vm

import (
	"context"

	computemngr "github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute"
	"github.com/Azure/azure-sdk-for-go/services/compute/mgmt/2019-12-01/compute/computeapi"
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

	"github.com/crossplane/provider-azure/apis/compute/v1alpha3"
	azure "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/crossplane/provider-azure/pkg/clients/compute"
)

// Error strings.
const (
	errNotVirtualMachine    = "managed resource is not a VirtualMachine"
	errCreateVirtualMachine = "cannot create VirtualMachine"
	errGetVirtualMachine    = "cannot get VirtualMachine"
	errDeleteVirtualMachine = "cannot delete VirtualMachine"
)

// SetupVirtualMachine adds a controller that reconciles VirtualMachine.
func SetupVirtualMachine(mgr ctrl.Manager, l logging.Logger, rl workqueue.RateLimiter) error {
	name := managed.ControllerName(v1alpha3.VirtualMachineKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(controller.Options{
			RateLimiter: ratelimiter.NewDefaultManagedRateLimiter(rl),
		}).
		For(&v1alpha3.VirtualMachine{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.VirtualMachineGroupVersionKind),
			managed.WithConnectionPublishers(),
			managed.WithExternalConnecter(&connecter{client: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithLogger(l.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name)))))
}

type connecter struct {
	client client.Client
}

func (c *connecter) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azure.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := computemngr.NewVirtualMachinesClient(creds[azure.CredentialsKeySubscriptionID])
	diskClient := computemngr.NewDisksClient(creds[azure.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	diskClient.Authorizer = auth
	return &external{client: cl, diskClient: diskClient}, nil
}

type external struct {
	client     computeapi.VirtualMachinesClientAPI
	diskClient computeapi.DisksClientAPI
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha3.VirtualMachine)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotVirtualMachine)
	}

	c, err := e.client.Get(ctx, cr.Spec.ResourceGroupName, meta.GetExternalName(cr), computemngr.InstanceView)
	if azure.IsNotFound(err) {
		if cr.Status.OSDiskName == "" {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		disk, err := e.diskClient.Get(ctx, cr.Spec.ResourceGroupName, cr.Status.OSDiskName)
		if azure.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errGetVirtualMachine)
		}
		compute.UpdateDiskStatus(cr, &disk)
		return managed.ExternalObservation{ResourceExists: true}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetVirtualMachine)
	}
	compute.UpdateVirtualMachineStatus(cr, &c)
	if cr.Status.State == "Succeeded" {
		cr.SetConditions(xpv1.Available())
	}

	// VirtualMachine are always up to date because we can't yet update them.
	o := managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
	}
	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha3.VirtualMachine)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotVirtualMachine)
	}
	cr.SetConditions(xpv1.Creating())
	if _, err := e.client.CreateOrUpdate(ctx, cr.Spec.ResourceGroupName, meta.GetExternalName(cr), compute.NewVirtualMachine(cr.Spec.VirtualMachineParameters)); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateVirtualMachine)
	}
	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	if _, ok := mg.(*v1alpha3.VirtualMachine); !ok {
		return managed.ExternalUpdate{}, errors.New(errNotVirtualMachine)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha3.VirtualMachine)
	if !ok {
		return errors.New(errNotVirtualMachine)
	}
	cr.SetConditions(xpv1.Deleting())
	_, err := e.client.Delete(ctx, cr.Spec.ResourceGroupName, meta.GetExternalName(cr))
	if azure.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, errDeleteVirtualMachine)
	}
	_, err = e.diskClient.Delete(ctx, cr.Spec.ResourceGroupName, cr.Status.OSDiskName)
	if azure.IsNotFound(err) && cr.Status.OSDiskName != "" {
		return nil
	}
	if err != nil {
		return errors.Wrap(err, errDeleteVirtualMachine)
	}
	return nil
}
