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
package SecurityGroup

import (
	"context"
	azurenetwork "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network"
	"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-06-01/network/networkapi"
	securitygroup "github.com/crossplane/provider-azure/pkg/clients/network"

	//"github.com/crossplane/provider-azure/pkg/clients/network"

	//"github.com/Azure/azure-sdk-for-go/services/network/mgmt/2019-12-01/network/networkapi"
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/logging"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/crossplane/provider-azure/apis/network/v1alpha3"
	azureclients "github.com/crossplane/provider-azure/pkg/clients"
	"github.com/pkg/errors"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Error strings.
const (
	errNotSecurityGroup    = "managed resource is not an SecurityGroup"
	errCreateSecurityGroup = "cannot create SecurityGroup"
	errUpdateSecurityGroup = "cannot update SecurityGroup"
	errGetSecurityGroup    = "cannot get SecurityGroup"
	errDeleteSecurityGroup = "cannot delete SecurityGroup"
)

// Setup adds a controller that reconciles Security Group.
func Setup(mgr ctrl.Manager, l logging.Logger) error {
	name := managed.ControllerName(v1alpha3.SecurityGroupGroupKind)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		For(&v1alpha3.SecurityGroup{}).
		Complete(managed.NewReconciler(mgr,
			resource.ManagedKind(v1alpha3.SecurityGroupGroupVersionKind),
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
	creds, auth, err := azureclients.GetAuthInfo(ctx, c.client, mg)
	if err != nil {
		return nil, err
	}
	cl := azurenetwork.NewSecurityGroupsClient(creds[azureclients.CredentialsKeySubscriptionID])
	cl.Authorizer = auth
	return &external{client: cl}, nil
}

type external struct {
	client networkapi.SecurityGroupsClientAPI
	//client azurenetwork.SecurityGroupsClient
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	v, ok := mg.(*v1alpha3.SecurityGroup)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSecurityGroup)
	}
	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), "")
	if azureclients.IsNotFound(err) {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errGetSecurityGroup)
	}

	securitygroup.UpdateSecurityGroupStatusFromAzure(v, az)

	v.SetConditions(runtimev1alpha1.Available())

	o := managed.ExternalObservation{
		ResourceExists:    true,
		ConnectionDetails: managed.ConnectionDetails{},
	}
	return o, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	v, ok := mg.(*v1alpha3.SecurityGroup)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSecurityGroup)
	}
	v.Status.SetConditions(runtimev1alpha1.Creating())

	sg := securitygroup.NewSecurityGroupParameters(v)

	if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), sg); err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateSecurityGroup)
	}

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	v, ok := mg.(*v1alpha3.SecurityGroup)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSecurityGroup)
	}
	az, err := e.client.Get(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), "")
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errGetSecurityGroup)
	}
	if securitygroup.SecurityGroupNeedsUpdate(v, az) {
		sg := securitygroup.NewSecurityGroupParameters(v)
		if _, err := e.client.CreateOrUpdate(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v), sg); err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateSecurityGroup)
		}
	}
	return managed.ExternalUpdate{}, nil
}
func (e *external) Delete(ctx context.Context, mg resource.Managed) error {
	v, ok := mg.(*v1alpha3.SecurityGroup)
	if !ok {
		return errors.New(errNotSecurityGroup)
	}

	mg.SetConditions(runtimev1alpha1.Deleting())

	_, err := e.client.Delete(ctx, v.Spec.ResourceGroupName, meta.GetExternalName(v))
	return errors.Wrap(resource.Ignore(azureclients.IsNotFound, err), errDeleteSecurityGroup)
}
