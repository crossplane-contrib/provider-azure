/*
Copyright 2022 The Crossplane Authors.

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

package containers

import (
	"context"
	"encoding/base64"

	"github.com/Azure/azure-sdk-for-go/services/authorization/mgmt/2015-07-01/authorization"
	"github.com/Azure/azure-sdk-for-go/services/redhatopenshift/mgmt/2022-04-01/redhatopenshift"
	"github.com/Azure/azure-sdk-for-go/services/redhatopenshift/mgmt/2022-04-01/redhatopenshift/redhatopenshiftapi"
	"github.com/crossplane-contrib/provider-azure/apis/containers/v1alpha1"
	apisv1alpha1 "github.com/crossplane-contrib/provider-azure/apis/v1alpha1"
	azure "github.com/crossplane-contrib/provider-azure/pkg/clients"
	openshiftclient "github.com/crossplane-contrib/provider-azure/pkg/clients/containers"
	"github.com/crossplane-contrib/provider-azure/pkg/features"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	errNotOpenshift                 = "managed resource is not a Openshift custom resource"
	errTrackPCUsage                 = "cannot track ProviderConfig usage"
	errGetPC                        = "cannot get ProviderConfig"
	errGetCreds                     = "cannot get credentials"
	errGetFailed                    = "cannot get Openshift cluster from Azure API"
	errNewClient                    = "cannot create new Service"
	errConnectFailed                = "cannot connect to Azure API"
	errUpdateOpenShiftClusterFailed = "cannot update Openshift cluster in Azure API"
	errCreateFailed                 = "cannot create the Openshift cluster"
	errUpdateFailed                 = "cannot update the Openshift cluster"
	errDeleteFailed                 = "cannot delete the Openshift cluster"
	errListAccessKeysFailed         = "cannot get credentials list"
)

// Setup adds a controller that reconciles Openshift managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.OpenshiftGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), apisv1alpha1.StoreConfigGroupVersionKind))
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		For(&v1alpha1.Openshift{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha1.OpenshiftGroupVersionKind),
			managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
			managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
			managed.WithPollInterval(o.PollInterval),
			managed.WithLogger(o.Logger.WithValues("controller", name)),
			managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
			managed.WithConnectionPublishers(cps...)))
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube client.Client
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	creds, auth, err := azure.GetAuthInfo(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errConnectFailed)
	}
	cl := redhatopenshift.NewOpenShiftClustersClient(creds[azure.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	rac := authorization.NewRoleAssignmentsClient(creds[azure.CredentialsKeySubscriptionID])
	rac.Authorizer = auth
	return &external{kube: c.kube, redhatClient: cl, roleClient: rac}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	kube       client.Client
	redhatClient     redhatopenshiftapi.OpenShiftClustersClientAPI
	roleClient authorization.RoleAssignmentsClient
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Openshift)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotOpenshift)
	}
	openShiftCluster, err := c.redhatClient.Get(ctx, cr.Spec.ForProvider.ResourceGroupNameRef.Name, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errors.Wrap(resource.Ignore(azure.IsNotFound, err), errGetFailed)
	}

	openshiftclient.LateInitialize(&cr.Spec.ForProvider, openShiftCluster)
	if err := c.kube.Update(ctx, cr); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errUpdateOpenShiftClusterFailed)
	}
	cr.Status.AtProvider = openshiftclient.GenerateObservation(openShiftCluster)
	var conn managed.ConnectionDetails
	switch cr.Status.AtProvider.ProvisioningState {
	case openshiftclient.ProvisioningStateSucceeded:
		kubeConfig, err := c.redhatClient.ListAdminCredentials(ctx, cr.Spec.ForProvider.ResourceGroupNameRef.Name, meta.GetExternalName(cr))
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errListAccessKeysFailed)
		}
		kubeconfigDecoded, err := base64.StdEncoding.DecodeString(*kubeConfig.Kubeconfig)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errGetCreds)
		}
		creds, err := c.redhatClient.ListCredentials(ctx, cr.Spec.ForProvider.ResourceGroupNameRef.Name, meta.GetExternalName(cr))
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, errListAccessKeysFailed)
		}
		conn = managed.ConnectionDetails{
			xpv1.ResourceCredentialsSecretKubeconfigKey: []byte(kubeconfigDecoded),
			xpv1.ResourceCredentialsSecretUserKey:       []byte(*creds.KubeadminUsername),
			xpv1.ResourceCredentialsSecretPasswordKey:   []byte(*creds.KubeadminPassword),
		}
		cr.Status.SetConditions(xpv1.Available())
	case openshiftclient.ProvisioningStateCreating:
		cr.Status.SetConditions(xpv1.Creating())
	case openshiftclient.ProvisioningStateDeleting:
		cr.Status.SetConditions(xpv1.Deleting())
	default:
		cr.Status.SetConditions(xpv1.Unavailable())
	}
	return managed.ExternalObservation{
		ResourceExists:    true,
		ResourceUpToDate:  true,
		ConnectionDetails: conn,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Openshift)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotOpenshift)
	}
	cr, err := openshiftclient.ExtractSecrets(ctx, c.kube, cr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGetCreds)
	}

	err = openshiftclient.EnsureResourceGroup(ctx, cr.Spec.ForProvider.ServicePrincipalProfile.PrincipalID ,openshiftclient.ExtractScopeFromSubnetID(cr.Spec.ForProvider.MasterProfile.SubnetID), c.roleClient )
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGetCreds)
	}

	err = openshiftclient.EnsureResourceGroup(ctx, cr.Spec.ForProvider.ServicePrincipalProfile.AzureRedHatOpenShiftRPPrincipalID ,openshiftclient.ExtractScopeFromSubnetID(cr.Spec.ForProvider.MasterProfile.SubnetID), c.roleClient )
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errGetCreds)
	}

	_, err = c.redhatClient.CreateOrUpdate(ctx, cr.Spec.ForProvider.ResourceGroupNameRef.Name, meta.GetExternalName(cr), openshiftclient.NewCreateParameters(ctx, cr, c.kube))
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errNewClient)
	}

	return managed.ExternalCreation{}, errors.Wrap(err, errCreateFailed)
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Openshift)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotOpenshift)
	}

	if cr.Status.AtProvider.ProvisioningState != openshiftclient.ProvisioningStateSucceeded {
		return managed.ExternalUpdate{}, nil
	}

	openShiftCluster, err := c.redhatClient.Get(ctx, cr.Spec.ForProvider.ResourceGroupNameRef.Name, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetFailed)
	}

	_, err = c.redhatClient.Update(
		ctx,
		cr.Spec.ForProvider.ResourceGroupNameRef.Name,
		meta.GetExternalName(cr),
		openshiftclient.NewUpdateParameters(cr.Spec.ForProvider, openShiftCluster))
	return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateFailed)
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Openshift)
	if !ok {
		return errors.New(errNotOpenshift)
	}
	cr.Status.SetConditions(xpv1.Deleting())
	if cr.Status.AtProvider.ProvisioningState == openshiftclient.ProvisioningStateDeleting {
		return nil
	}
	_, err := c.redhatClient.Delete(ctx, cr.Spec.ForProvider.ResourceGroupNameRef.Name, meta.GetExternalName(cr))
	return errors.Wrap(resource.Ignore(azure.IsNotFound, err), errDeleteFailed)
}
