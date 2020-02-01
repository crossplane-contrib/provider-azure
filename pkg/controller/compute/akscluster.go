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
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/clientcmd"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplaneio/crossplane-runtime/pkg/resource"

	computev1alpha3 "github.com/crossplaneio/stack-azure/apis/compute/v1alpha3"
	azurev1alpha3 "github.com/crossplaneio/stack-azure/apis/v1alpha3"
	azureclients "github.com/crossplaneio/stack-azure/pkg/clients"
	"github.com/crossplaneio/stack-azure/pkg/clients/compute"
)

const (
	controllerName = "aks.compute.azure.crossplane.io"
	finalizer      = "finalizer." + controllerName
	spSecretKey    = "clientSecret"
	adAppNameFmt   = "%s-crossplane-aks-app"
)

// Amounts of time we wait before requeuing a reconcile.
const (
	aLongWait = 60 * time.Second
)

// Error strings
const (
	errUpdateManagedStatus = "cannot update managed resource status"
)

var (
	ctx           = context.TODO()
	result        = reconcile.Result{}
	resultRequeue = reconcile.Result{Requeue: true}
)

// Reconciler reconciles a AKSCluster object
type Reconciler struct {
	client.Client
	newClientFn        func(creds []byte) (*azureclients.Client, error)
	aksSetupAPIFactory compute.AKSSetupAPIFactory
	publisher          managed.ConnectionPublisher
	resolver           managed.ReferenceResolver

	log logging.Logger
}

// SetupAKSCluster adds a controller that reconciles AKS clusters.
func SetupAKSCluster(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(computev1alpha3.AKSClusterKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&computev1alpha3.AKSCluster{}).
		Complete(&Reconciler{
			Client:             mgr.GetClient(),
			newClientFn:        azureclients.NewClient,
			aksSetupAPIFactory: &compute.AKSSetupClientFactory{},
			publisher:          managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme()),
			resolver:           managed.NewAPIReferenceResolver(mgr.GetClient()),
			log:                l.WithValues("controller", name)})
}

// Reconcile reads that state of the cluster for a AKSCluster object and makes changes based on the state read
// and what is in its spec.
func (r *Reconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) { // nolint:gocyclo
	// NOTE(soorena776): This method is a little over our cyclomatic complexity
	// goal, but keeping it as is for now, as we will eventually use the Managed
	// Reconciler for all types

	log := r.log.WithValues("request", request)
	log.Debug("Reconciling")

	// Fetch the CRD instance
	instance := &computev1alpha3.AKSCluster{}
	err := r.Get(ctx, request.NamespacedName, instance)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Create AKS Client
	aksClient, err := r.connect(instance)
	if err != nil {
		return r.fail(instance, err)
	}

	if !resource.IsConditionTrue(instance.GetCondition(runtimev1alpha1.TypeReferencesResolved)) {
		if err := r.resolver.ResolveReferences(ctx, instance); err != nil {
			condition := runtimev1alpha1.ReconcileError(err)
			if managed.IsReferencesAccessError(err) {
				condition = runtimev1alpha1.ReferenceResolutionBlocked(err)
			}

			instance.Status.SetConditions(condition)
			return reconcile.Result{RequeueAfter: aLongWait}, errors.Wrap(r.Update(ctx, instance), errUpdateManagedStatus)
		}

		// Add ReferenceResolutionSuccess to the conditions
		instance.Status.SetConditions(runtimev1alpha1.ReferenceResolutionSuccess())
	}

	// Check for deletion
	if instance.DeletionTimestamp != nil {
		log.Debug("AKS cluster has been deleted, running finalizer now")
		return r.delete(instance, aksClient)
	}

	// TODO(negz): Move finalizer creation into the create method?
	// Add finalizer
	meta.AddFinalizer(instance, finalizer)
	if err := r.Update(ctx, instance); err != nil {
		return resultRequeue, err
	}

	if instance.Status.RunningOperation != "" {
		// there is a running operation on the instance, wait for it to complete
		return r.waitForCompletion(instance, aksClient)
	}

	// Create cluster instance
	if !r.created(instance) {
		return r.create(instance, aksClient)
	}

	// Sync cluster instance status with cluster status
	return r.sync(instance, aksClient)
}

func (r *Reconciler) connect(instance *computev1alpha3.AKSCluster) (*compute.AKSSetupClient, error) {
	p := &azurev1alpha3.Provider{}
	if err := r.Get(ctx, meta.NamespacedNameOf(instance.Spec.ProviderReference), p); err != nil {
		return nil, errors.Wrap(err, "failed to get provider")
	}

	s := &v1.Secret{}
	n := types.NamespacedName{Namespace: p.Spec.CredentialsSecretRef.Namespace, Name: p.Spec.CredentialsSecretRef.Name}
	if err := r.Get(ctx, n, s); err != nil {
		return nil, errors.Wrap(err, "failed to get provider secret")
	}

	c, err := r.newClientFn(s.Data[p.Spec.CredentialsSecretRef.Key])
	if err != nil {
		return nil, errors.Wrap(err, "failed to create Azure client")
	}

	return r.aksSetupAPIFactory.CreateSetupClient(c)
}

// TODO(negz): This method's cyclomatic complexity is a little high. Consider
// refactoring to reduce said complexity if you touch it.
// nolint:gocyclo
func (r *Reconciler) create(instance *computev1alpha3.AKSCluster, aksClient *compute.AKSSetupClient) (reconcile.Result, error) {
	instance.Status.SetConditions(runtimev1alpha1.Creating())
	// create or fetch the secret for the AD application and its service principal the cluster will use for Azure APIs
	spSecret, err := r.servicePrincipalSecret(instance)
	if err != nil {
		return r.fail(instance, errors.Wrapf(err, "failed to get service principal secret for AKS cluster %s", instance.Name))
	}

	// create the AD application that the cluster will use for the Azure APIs
	appParams := azureclients.ApplicationParameters{
		Name:          fmt.Sprintf(adAppNameFmt, instance.Name),
		DNSNamePrefix: instance.Spec.DNSNamePrefix,
		Location:      instance.Spec.Location,
		ObjectID:      instance.Status.ApplicationObjectID,
		ClientSecret:  spSecret,
	}
	r.log.Debug("starting create of app for AKS cluster")
	app, err := aksClient.ApplicationAPI.CreateApplication(ctx, appParams)
	if err != nil {
		return r.fail(instance, errors.Wrapf(err, "failed to create app for AKS cluster %s", instance.Name))
	}

	if instance.Status.ApplicationObjectID == "" {
		// save the application object ID on the CRD status now
		instance.Status.ApplicationObjectID = *app.ObjectID
		// TODO: retry this CRD update upon conflict
		r.Update(ctx, instance) // nolint:errcheck
	}

	// create the service principal for the AD application
	r.log.Debug("starting create of service principal for AKS cluster")
	sp, err := aksClient.ServicePrincipalAPI.CreateServicePrincipal(ctx, instance.Status.ServicePrincipalID, *app.AppID)
	if err != nil {
		return r.fail(instance, errors.Wrapf(err, "failed to create service principal for AKS cluster %s", instance.Name))
	}

	if instance.Status.ServicePrincipalID == "" {
		// save the service principal ID on the CRD status now
		instance.Status.ServicePrincipalID = *sp.ObjectID
		// TODO: retry this CRD update upon conflict
		r.Update(ctx, instance) // nolint:errcheck
	}

	// create the role assignment for the service principal if subnet defined
	if instance.Spec.VnetSubnetID != "" {
		r.log.Debug("starting create of role assignment for service principal for AKS cluster")
		name := string(instance.GetUID())
		_, err := aksClient.RoleAssignmentsAPI.CreateRoleAssignment(ctx, instance.Status.ServicePrincipalID, instance.Spec.VnetSubnetID, name)
		if err != nil {
			return r.fail(instance, errors.Wrapf(err, "failed to create role assignment for service principal for AKS cluster %s", instance.Name))
		}
	}

	// start the creation of the AKS cluster
	r.log.Debug("starting create of AKS cluster")
	clusterName := compute.SanitizeClusterName(instance.Name)
	createOp, err := aksClient.AKSClusterAPI.CreateOrUpdateBegin(ctx, *instance, clusterName, *app.AppID, spSecret)
	if err != nil {
		return r.fail(instance, errors.Wrapf(err, "failed to start create operation for AKS cluster %s", instance.Name))
	}

	r.log.Debug("started create of AKS cluster", "operation", string(createOp))

	// save the create operation to the CRD status
	instance.Status.RunningOperation = string(createOp)

	// set the creating/provisioning state to the CRD status
	instance.Status.ClusterName = clusterName

	instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())

	b := wait.Backoff{
		Steps:    10,
		Duration: 500 * time.Millisecond,
		Factor:   1.0,
		Jitter:   0.1,
	}
	// wait until the important status fields we just set have become committed/consistent
	updateWaitErr := wait.ExponentialBackoff(b, func() (done bool, err error) {
		if err := r.Update(ctx, instance); err != nil {
			return false, nil
		}

		// the update went through, let's do a get to verify the fields are committed/consistent
		// TODO(negz): Is this necessary? The update call should populate the
		// instance struct with the latest view of the world.
		fetchedInstance := &computev1alpha3.AKSCluster{}
		if err := r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, fetchedInstance); err != nil {
			return false, nil
		}

		if fetchedInstance.Status.RunningOperation != "" && fetchedInstance.Status.ClusterName != "" {
			// both the running operation field and the cluster name field have been committed, we can stop retrying
			return true, nil
		}

		// the instance hasn't reached consistency yet, retry
		r.log.Debug("AKS cluster hasn't reached consistency yet, retrying", "instance", instance)
		return false, nil
	})

	return resultRequeue, updateWaitErr
}

func (r *Reconciler) waitForCompletion(instance *computev1alpha3.AKSCluster, aksClient *compute.AKSSetupClient) (reconcile.Result, error) {
	// check if the operation is done yet and if there was any error
	done, err := aksClient.AKSClusterAPI.CreateOrUpdateEnd([]byte(instance.Status.RunningOperation))
	if !done {
		// not done yet, check again on the next reconcile
		r.log.Debug("waiting on create of AKS cluster")
		return resultRequeue, err
	}

	// the operation is done, clear out the running operation on the CRD status
	instance.Status.RunningOperation = ""

	if err != nil {
		// the operation completed, but there was an error
		return r.fail(instance, errors.Wrapf(err, "failure result returned from create operation for AKS cluster %s", instance.Name))
	}

	r.log.Debug("AKS cluster successfully created")
	return resultRequeue, r.Update(ctx, instance)
}

func (r *Reconciler) created(instance *computev1alpha3.AKSCluster) bool {
	return instance.Status.ClusterName != ""
}

func (r *Reconciler) sync(instance *computev1alpha3.AKSCluster, aksClient *compute.AKSSetupClient) (reconcile.Result, error) {
	cluster, err := aksClient.AKSClusterAPI.Get(ctx, *instance)
	if err != nil {
		return r.fail(instance, err)
	}

	cd, err := r.connectionDetails(instance, aksClient)
	if err != nil {
		return r.fail(instance, err)
	}

	if err := r.publisher.PublishConnection(ctx, instance, cd); err != nil {
		return r.fail(instance, err)
	}

	// update resource status
	if cluster.ID != nil {
		instance.Status.ProviderID = *cluster.ID
	}
	if cluster.ProvisioningState != nil {
		instance.Status.State = *cluster.ProvisioningState
	}
	if cluster.Fqdn != nil {
		instance.Status.Endpoint = *cluster.Fqdn
	}

	instance.Status.SetConditions(runtimev1alpha1.Available(), runtimev1alpha1.ReconcileSuccess())
	resource.SetBindable(instance)
	return result, r.Update(ctx, instance)
}

// delete performs a deletion of the AKS cluster if needed
func (r *Reconciler) delete(instance *computev1alpha3.AKSCluster, aksClient *compute.AKSSetupClient) (reconcile.Result, error) { // nolint:gocyclo
	instance.Status.SetConditions(runtimev1alpha1.Deleting())
	if instance.Spec.ReclaimPolicy == runtimev1alpha1.ReclaimDelete {
		// delete the AKS cluster
		r.log.Debug("deleting AKS cluster")
		deleteFuture, err := aksClient.AKSClusterAPI.Delete(ctx, *instance)
		if err != nil && !azureclients.IsNotFound(err) {
			return r.fail(instance, errors.Wrapf(err, "failed to delete AKS cluster %s", instance.Name))
		}
		deleteFutureJSON, _ := deleteFuture.MarshalJSON()
		r.log.Debug("started delete of AKS cluster", "operation", string(deleteFutureJSON))

		// delete the role assignment if created
		if instance.Spec.VnetSubnetID != "" {
			r.log.Debug("deleting role assignment for service principal for AKS cluster")
			name := string(instance.GetUID())
			err = aksClient.RoleAssignmentsAPI.DeleteRoleAssignment(ctx, instance.Spec.VnetSubnetID, name)
			if err != nil && !azureclients.IsNotFound(err) {
				return r.fail(instance, errors.Wrap(err, "failed to delete role assignment for service principal"))
			}
		}

		// delete the service principal
		r.log.Debug("deleting service principal for AKS cluster")
		err = aksClient.ServicePrincipalAPI.DeleteServicePrincipal(ctx, instance.Status.ServicePrincipalID)
		if err != nil && !azureclients.IsNotFound(err) {
			return r.fail(instance, errors.Wrap(err, "failed to delete service principal"))
		}

		// delete the AD application
		r.log.Debug("deleting app for AKS cluster")
		err = aksClient.ApplicationAPI.DeleteApplication(ctx, instance.Status.ApplicationObjectID)
		if err != nil && !azureclients.IsNotFound(err) {
			return r.fail(instance, errors.Wrap(err, "failed to delete AD application"))
		}

		r.log.Debug("all resources deleted for AKS cluster")
	}

	meta.RemoveFinalizer(instance, finalizer)
	instance.Status.SetConditions(runtimev1alpha1.ReconcileSuccess())
	return result, r.Update(ctx, instance)
}

// fail - helper function to set fail condition with reason and message
func (r *Reconciler) fail(instance *computev1alpha3.AKSCluster, err error) (reconcile.Result, error) {
	instance.Status.SetConditions(runtimev1alpha1.ReconcileError(err))
	return resultRequeue, r.Update(ctx, instance)
}

func (r *Reconciler) servicePrincipalSecret(instance *computev1alpha3.AKSCluster) (string, error) {
	s := &v1.Secret{}
	n := types.NamespacedName{
		Namespace: instance.Spec.WriteServicePrincipalSecretTo.Namespace,
		Name:      instance.Spec.WriteServicePrincipalSecretTo.Name,
	}
	err := r.Get(ctx, n, s)
	if resource.IgnoreNotFound(err) != nil {
		return "", errors.Wrap(err, "cannot get service principal secret")
	}

	// We successfully read the existing service principal secret, so return its
	// contents.
	if err == nil {
		// TODO(negz): Ensure we're the owner of this secret before trusting it.
		return string(s.Data[spSecretKey]), nil
	}

	// service principal secret must not exist yet, generate a new one
	newSPSecretValue, err := uuid.NewRandom()
	if err != nil {
		return "", errors.Wrap(err, "failed to generate client secret")
	}

	// save the service principal secret
	ref := meta.AsController(meta.ReferenceTo(instance, computev1alpha3.AKSClusterGroupVersionKind))
	spSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       instance.Spec.WriteServicePrincipalSecretTo.Namespace,
			Name:            instance.Spec.WriteServicePrincipalSecretTo.Name,
			OwnerReferences: []metav1.OwnerReference{ref},
		},
		Data: map[string][]byte{spSecretKey: []byte(newSPSecretValue.String())},
	}

	if err := r.Create(ctx, spSecret); err != nil {
		return "", errors.Wrap(err, "failed to create service principal secret")
	}

	return newSPSecretValue.String(), nil
}

func (r *Reconciler) connectionDetails(instance *computev1alpha3.AKSCluster, client *compute.AKSSetupClient) (managed.ConnectionDetails, error) {
	creds, err := client.ListClusterAdminCredentials(ctx, *instance)
	if err != nil {
		return nil, err
	}

	// TODO(negz): It's not clear in what case this would contain more than one kubeconfig file.
	// https://docs.microsoft.com/en-us/rest/api/aks/managedclusters/listclusteradmincredentials#credentialresults
	if creds.Kubeconfigs == nil || len(*creds.Kubeconfigs) == 0 || (*creds.Kubeconfigs)[0].Value == nil {
		return nil, errors.Errorf("zero kubeconfig credentials returned")
	}
	// Azure's generated Godoc claims Value is a 'base64 encoded kubeconfig'.
	// This is true on the wire, but not true in the actual struct because
	// encoding/json automatically base64 encodes and decodes byte slices.
	kcfg, err := clientcmd.Load(*(*creds.Kubeconfigs)[0].Value)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse kubeconfig file")
	}

	kctx, ok := kcfg.Contexts[instance.Status.ClusterName]
	if !ok {
		return nil, errors.Errorf("context configuration is not found for cluster: %s", instance.Status.ClusterName)
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
		runtimev1alpha1.ResourceCredentialsSecretKubeconfigKey: *(*creds.Kubeconfigs)[0].Value,
	}, nil
}
