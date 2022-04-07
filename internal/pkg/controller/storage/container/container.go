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

package container

import (
	"context"
	"reflect"
	"time"

	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	azure "github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients"
	"github.com/crossplane-contrib/provider-jet-azure/internal/pkg/clients/storage"

	"github.com/crossplane-contrib/provider-jet-azure/apis/classic/storage/v1alpha3"
)

const (
	controllerName = "container.storage.azure.crossplane.io"
	finalizer      = "finalizer." + controllerName

	reconcileTimeout = 2 * time.Minute
)

// Error strings
const (
	errAcctSecretNil = "account does not have a connection secret"
)

var (
	resultRequeue = reconcile.Result{Requeue: true}
)

// Reconciler reconciles an Azure storage container
type Reconciler struct {
	client.Client
	syncdeleterMaker
	managed.ReferenceResolver
	managed.Initializer

	poll time.Duration

	log logging.Logger
}

// Setup adds a controller that reconciles Containers.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha3.ContainerGroupKind)

	r := &Reconciler{
		Client:           mgr.GetClient(),
		syncdeleterMaker: &containerSyncdeleterMaker{mgr.GetClient()},
		Initializer:      managed.NewNameAsExternalName(mgr.GetClient()),
		poll:             o.PollInterval,
		log:              o.Logger.WithValues("controller", name),
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha3.Container{}).
		Complete(r)
}

// Reconcile reads that state of the cluster for a Provider acct and makes changes based on the state read
// and what is in the Provider.Spec
func (r *Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	r.log.Debug("Reconciling", "request", request)

	ctx, cancel := context.WithTimeout(ctx, reconcileTimeout)
	defer cancel()

	c := &v1alpha3.Container{}
	if err := r.Get(ctx, request.NamespacedName, c); err != nil {
		return reconcile.Result{}, resource.Ignore(kerrors.IsNotFound, err)
	}
	if err := r.Initialize(ctx, c); err != nil {
		return reconcile.Result{}, err
	}

	sd, err := r.newSyncdeleter(ctx, c, r.poll)
	if err != nil {
		c.Status.SetConditions(xpv1.ReconcileError(err))
		return resultRequeue, r.Status().Update(ctx, c)
	}

	// Check for deletion
	if c.DeletionTimestamp != nil {
		return sd.delete(ctx)
	}

	return sd.sync(ctx)
}

type syncdeleterMaker interface {
	newSyncdeleter(context.Context, *v1alpha3.Container, time.Duration) (syncdeleter, error)
}

type containerSyncdeleterMaker struct {
	client.Client
}

func (m *containerSyncdeleterMaker) newSyncdeleter(ctx context.Context, c *v1alpha3.Container, poll time.Duration) (syncdeleter, error) { // nolint:gocyclo
	nn := types.NamespacedName{}
	switch {
	case c.GetProviderConfigReference() != nil && c.GetProviderConfigReference().Name != "":
		nn.Name = c.GetProviderConfigReference().Name
	case c.GetProviderReference() != nil && c.GetProviderReference().Name != "":
		nn.Name = c.GetProviderReference().Name
	default:
		return nil, errors.New("neither providerConfigRef nor providerRef is given")
	}
	// Storage containers use a storage account as their 'provider', not a
	// typical Azure provider.
	acct := &v1alpha3.Account{}
	if err := m.Get(ctx, nn, acct); err != nil {
		// For storage account not found errors - check if we are on deletion path
		// if so - remove finalizer from this container object
		if kerrors.IsNotFound(err) && c.DeletionTimestamp != nil {
			meta.RemoveFinalizer(c, finalizer)
			if err := m.Client.Update(ctx, c); err != nil {
				return nil, errors.Wrapf(err, "failed to update after removing finalizer")
			}
		}
		return nil, errors.Wrapf(err, "failed to retrieve storage account: %s", nn.Name)
	}

	if acct.GetWriteConnectionSecretToReference() == nil {
		return nil, errors.New(errAcctSecretNil)
	}

	// Retrieve storage account secret
	s := &corev1.Secret{}
	n := types.NamespacedName{
		Namespace: acct.Spec.WriteConnectionSecretToReference.Namespace,
		Name:      acct.Spec.WriteConnectionSecretToReference.Name,
	}
	if err := m.Get(ctx, n, s); err != nil {
		return nil, errors.Wrapf(err, "failed to retrieve storage account secret: %s", n)
	}

	accountName := string(s.Data[xpv1.ResourceCredentialsSecretUserKey])
	accountPassword := string(s.Data[xpv1.ResourceCredentialsSecretPasswordKey])
	containerName := meta.GetExternalName(c)

	ch, err := storage.NewContainerHandle(accountName, accountPassword, containerName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create client handle: %s, storage account: %s", containerName, accountName)
	}

	// set owner reference on the container to storage account, thus
	// if the account is delete - container is garbage collected as well
	or := meta.AsOwner(meta.TypedReferenceTo(acct, v1alpha3.AccountGroupVersionKind))
	or.BlockOwnerDeletion = to.BoolPtr(true)
	meta.AddOwnerReference(c, or)

	return &containerSyncdeleter{
		createupdater: &containerCreateUpdater{
			ContainerOperations: ch,
			kube:                m.Client,
			container:           c,
			poll:                poll,
		},
		ContainerOperations: ch,
		kube:                m.Client,
		container:           c,
	}, nil
}

type deleter interface {
	delete(context.Context) (reconcile.Result, error)
}

type syncer interface {
	sync(context.Context) (reconcile.Result, error)
}

type creator interface {
	create(context.Context) (reconcile.Result, error)
}

type updater interface {
	update(context.Context, *azblob.PublicAccessType, azblob.Metadata) (reconcile.Result, error)
}

type syncdeleter interface {
	deleter
	syncer
}

type containerSyncdeleter struct {
	createupdater
	storage.ContainerOperations
	kube      client.Client
	container *v1alpha3.Container
}

func (csd *containerSyncdeleter) delete(ctx context.Context) (reconcile.Result, error) {
	csd.container.Status.SetConditions(xpv1.Deleting())
	if csd.container.Spec.DeletionPolicy == xpv1.DeletionDelete {
		if err := csd.Delete(ctx); err != nil && !azure.IsNotFound(err) {
			csd.container.Status.SetConditions(xpv1.ReconcileError(err))
			return resultRequeue, csd.kube.Status().Update(ctx, csd.container)
		}
	}

	// NOTE(negz): We don't update the conditioned status here because assuming
	// no other finalizers need to be cleaned up the object should cease to
	// exist after we update it.
	meta.RemoveFinalizer(csd.container, finalizer)
	return reconcile.Result{}, csd.kube.Update(ctx, csd.container)
}

func (csd *containerSyncdeleter) sync(ctx context.Context) (reconcile.Result, error) {
	access, meta, err := csd.Get(ctx)
	if err != nil && !storage.IsNotFoundError(err) {
		csd.container.Status.SetConditions(xpv1.ReconcileError(err))
		return resultRequeue, csd.kube.Status().Update(ctx, csd.container)
	}

	if access == nil {
		return csd.create(ctx)
	}

	return csd.update(ctx, access, meta)
}

type createupdater interface {
	creator
	updater
}

// containerCreateUpdater implementation of createupdater interface
type containerCreateUpdater struct {
	storage.ContainerOperations
	kube      client.Client
	container *v1alpha3.Container
	poll      time.Duration
}

var _ createupdater = &containerCreateUpdater{}

func (ccu *containerCreateUpdater) create(ctx context.Context) (reconcile.Result, error) {
	container := ccu.container
	container.Status.SetConditions(xpv1.Creating())

	meta.AddFinalizer(container, finalizer)
	if err := ccu.kube.Update(ctx, container); err != nil {
		return resultRequeue, errors.Wrapf(err, "failed to update container spec")
	}

	spec := container.Spec
	if err := ccu.Create(ctx, spec.PublicAccessType, spec.Metadata); err != nil {
		container.Status.SetConditions(xpv1.ReconcileError(err))
		return resultRequeue, ccu.kube.Status().Update(ctx, container)
	}

	container.Status.SetConditions(xpv1.Available(), xpv1.ReconcileSuccess())
	return reconcile.Result{}, ccu.kube.Status().Update(ctx, ccu.container)
}

func (ccu *containerCreateUpdater) update(ctx context.Context, accessType *azblob.PublicAccessType, meta azblob.Metadata) (reconcile.Result, error) {
	container := ccu.container
	spec := container.Spec

	if !reflect.DeepEqual(*accessType, spec.PublicAccessType) || !reflect.DeepEqual(meta, spec.Metadata) {
		if err := ccu.Update(ctx, spec.PublicAccessType, spec.Metadata); err != nil {
			container.Status.SetConditions(xpv1.ReconcileError(err))
			return resultRequeue, ccu.kube.Status().Update(ctx, container)
		}
	}

	container.Status.SetConditions(xpv1.Available(), xpv1.ReconcileSuccess())
	return reconcile.Result{RequeueAfter: ccu.poll}, ccu.kube.Status().Update(ctx, ccu.container)
}
