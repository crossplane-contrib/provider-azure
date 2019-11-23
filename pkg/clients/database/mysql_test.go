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

package database

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	runtimev1alpha1 "github.com/crossplaneio/crossplane-runtime/apis/core/v1alpha1"
	"github.com/crossplaneio/crossplane-runtime/pkg/test"

	"github.com/crossplaneio/stack-azure/apis/database/v1alpha3"
	"github.com/crossplaneio/stack-azure/apis/database/v1beta1"
	azure "github.com/crossplaneio/stack-azure/pkg/clients"
)

const (
	uid           = types.UID("definitely-a-uuid")
	vnetRuleName  = "myvnetrule"
	serverName    = "myserver"
	rgName        = "myrg"
	vnetSubnetID  = "a/very/important/subnet"
	ignoreMissing = true

	id           = "very-cool-id"
	resourceType = "very-cool-type"
	credentials  = `
		{
			"clientId": "cool-id",
			"clientSecret": "cool-secret",
			"tenantId": "cool-tenant",
			"subscriptionId": "cool-subscription",
			"activeDirectoryEndpointUrl": "cool-aad-url",
			"resourceManagerEndpointUrl": "cool-rm-url",
			"activeDirectoryGraphResourceId": "cool-graph-id"
		}
	`
)

var (
	ctx = context.Background()
)

func TestNewMySQLVirtualNetworkRulesClient(t *testing.T) {
	cases := []struct {
		name       string
		r          []byte
		returnsErr bool
	}{
		{
			name: "Successful",
			r:    []byte(credentials),
		},
		{
			name:       "Unsuccessful",
			r:          []byte("invalid"),
			returnsErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := NewMySQLVirtualNetworkRulesClient(ctx, tc.r)

			if tc.returnsErr != (err != nil) {
				t.Errorf("NewMySQLVirtualNetworkRulesClient(...) error: want: %t got: %t", tc.returnsErr, err != nil)
			}

			if _, ok := got.(MySQLVirtualNetworkRulesClient); !ok && !tc.returnsErr {
				t.Error("NewMySQLVirtualNetworkRulesClient(...): got does not satisfy MySQLVirtualNetworkRulesClient interface")
			}
		})
	}
}

func TestNewMySQLVirtualNetworkRuleParameters(t *testing.T) {
	cases := []struct {
		name string
		r    *v1alpha3.MySQLServerVirtualNetworkRule
		want mysql.VirtualNetworkRule
	}{
		{
			name: "Successful",
			r: &v1alpha3.MySQLServerVirtualNetworkRule{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.MySQLVirtualNetworkRuleSpec{
					Name:              vnetRuleName,
					ServerName:        serverName,
					ResourceGroupName: rgName,
					VirtualNetworkRuleProperties: v1alpha3.VirtualNetworkRuleProperties{
						VirtualNetworkSubnetID:           vnetSubnetID,
						IgnoreMissingVnetServiceEndpoint: ignoreMissing,
					},
				},
			},
			want: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           to.StringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: to.BoolPtr(ignoreMissing),
				},
			},
		},
		{
			name: "SuccessfulPartial",
			r: &v1alpha3.MySQLServerVirtualNetworkRule{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.MySQLVirtualNetworkRuleSpec{
					Name:              vnetRuleName,
					ServerName:        serverName,
					ResourceGroupName: rgName,
					VirtualNetworkRuleProperties: v1alpha3.VirtualNetworkRuleProperties{
						VirtualNetworkSubnetID: vnetSubnetID,
					},
				},
			},
			want: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           to.StringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: to.BoolPtr(false),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NewMySQLVirtualNetworkRuleParameters(tc.r)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("NewMySQLVirtualNetworkRuleParameters(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestMySQLServerVirtualNetworkRuleNeedsUpdate(t *testing.T) {
	cases := []struct {
		name string
		kube *v1alpha3.MySQLServerVirtualNetworkRule
		az   mysql.VirtualNetworkRule
		want bool
	}{
		{
			name: "NoUpdateNeeded",
			kube: &v1alpha3.MySQLServerVirtualNetworkRule{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.MySQLVirtualNetworkRuleSpec{
					Name:              vnetRuleName,
					ServerName:        serverName,
					ResourceGroupName: rgName,
					VirtualNetworkRuleProperties: v1alpha3.VirtualNetworkRuleProperties{
						VirtualNetworkSubnetID:           vnetSubnetID,
						IgnoreMissingVnetServiceEndpoint: ignoreMissing,
					},
				},
			},
			az: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
				},
			},
			want: false,
		},
		{
			name: "UpdateNeededVirtualNetworkSubnetID",
			kube: &v1alpha3.MySQLServerVirtualNetworkRule{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.MySQLVirtualNetworkRuleSpec{
					Name:              vnetRuleName,
					ServerName:        serverName,
					ResourceGroupName: rgName,
					VirtualNetworkRuleProperties: v1alpha3.VirtualNetworkRuleProperties{
						VirtualNetworkSubnetID:           vnetSubnetID,
						IgnoreMissingVnetServiceEndpoint: ignoreMissing,
					},
				},
			},
			az: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr("some/other/subnet"),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
				},
			},
			want: true,
		},
		{
			name: "UpdateNeededIgnoreMissingVnetServiceEndpoint",
			kube: &v1alpha3.MySQLServerVirtualNetworkRule{
				ObjectMeta: metav1.ObjectMeta{UID: uid},
				Spec: v1alpha3.MySQLVirtualNetworkRuleSpec{
					Name:              vnetRuleName,
					ServerName:        serverName,
					ResourceGroupName: rgName,
					VirtualNetworkRuleProperties: v1alpha3.VirtualNetworkRuleProperties{
						VirtualNetworkSubnetID:           vnetSubnetID,
						IgnoreMissingVnetServiceEndpoint: ignoreMissing,
					},
				},
			},
			az: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(!ignoreMissing),
				},
			},
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MySQLServerVirtualNetworkRuleNeedsUpdate(tc.kube, tc.az)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("MySQLServerVirtualNetworkRuleNeedsUpdate(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestUpdateMySQLVirtualNetworkRuleStatusFromAzure(t *testing.T) {

	mockCondition := runtimev1alpha1.Condition{Message: "mockMessage"}
	resourceStatus := runtimev1alpha1.ResourceStatus{
		ConditionedStatus: runtimev1alpha1.ConditionedStatus{
			Conditions: []runtimev1alpha1.Condition{mockCondition},
		},
	}

	cases := []struct {
		name string
		r    mysql.VirtualNetworkRule
		want v1alpha3.VirtualNetworkRuleStatus
	}{
		{
			name: "SuccessfulFull",
			r: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				ID:   azure.ToStringPtr(id),
				Type: azure.ToStringPtr(resourceType),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
					State:                            mysql.Ready,
				},
			},
			want: v1alpha3.VirtualNetworkRuleStatus{
				State: "Ready",
				ID:    id,
				Type:  resourceType,
			},
		},
		{
			name: "SuccessfulPartial",
			r: mysql.VirtualNetworkRule{
				Name: azure.ToStringPtr(vnetRuleName),
				ID:   azure.ToStringPtr(id),
				VirtualNetworkRuleProperties: &mysql.VirtualNetworkRuleProperties{
					VirtualNetworkSubnetID:           azure.ToStringPtr(vnetSubnetID),
					IgnoreMissingVnetServiceEndpoint: azure.ToBoolPtr(ignoreMissing),
					State:                            mysql.Ready,
				},
			},
			want: v1alpha3.VirtualNetworkRuleStatus{
				State: "Ready",
				ID:    id,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v := &v1alpha3.MySQLServerVirtualNetworkRule{
				Status: v1alpha3.VirtualNetworkRuleStatus{
					ResourceStatus: resourceStatus,
				},
			}

			UpdateMySQLVirtualNetworkRuleStatusFromAzure(v, tc.r)

			// make sure that internal resource status hasn't changed
			if diff := cmp.Diff(mockCondition, v.Status.ResourceStatus.Conditions[0]); diff != "" {
				t.Errorf("UpdateMySQLVirtualNetworkRuleStatusFromAzure(...): -want, +got\n%s", diff)
			}

			// make sure that other resource parameters are updated
			tc.want.ResourceStatus = resourceStatus
			if diff := cmp.Diff(tc.want, v.Status); diff != "" {
				t.Errorf("UpdateMySQLVirtualNetworkRuleStatusFromAzure(...): -want, +got\n%s", diff)
			}
		})
	}
}

func TestFetchAsyncOperation(t *testing.T) {
	inprogressStatus := "inprogress"
	inProgressResponse := fmt.Sprintf(`{"status": "%s"}`, inprogressStatus)

	errorStatus := "Failed"
	errorResponse := fmt.Sprintf(`{"status": "%s"}`, errorStatus)
	errorMessage := fmt.Sprintf(`Code="Failed" Message="The async operation failed." AdditionalInfo=[{"status":"%s"}]`, errorStatus)

	pollingURL := "https://crossplane.io"

	type args struct {
		sender autorest.Sender
		as     *v1beta1.AsyncOperation
	}
	type want struct {
		op  *v1beta1.AsyncOperation
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"Should skip when there is no operation": {},
		"In progress": {
			args: args{
				as: &v1beta1.AsyncOperation{
					Method:     http.MethodPut,
					PollingURL: pollingURL,
				},
				sender: autorest.SenderFunc(func(req *http.Request) (*http.Response, error) {
					req.URL, _ = url.Parse("https://crossplane.io/resource1")
					return &http.Response{
						Request:       req,
						StatusCode:    http.StatusAccepted,
						Body:          ioutil.NopCloser(strings.NewReader(inProgressResponse)),
						ContentLength: int64(len([]byte(inProgressResponse))),
					}, nil
				}),
			},
			want: want{
				op: &v1beta1.AsyncOperation{
					Method:     http.MethodPut,
					PollingURL: pollingURL,
					Status:     inprogressStatus,
				},
			},
		},
		"Failure": {
			args: args{
				as: &v1beta1.AsyncOperation{
					Method:     http.MethodPut,
					PollingURL: pollingURL,
				},
				sender: autorest.SenderFunc(func(req *http.Request) (*http.Response, error) {
					req.URL, _ = url.Parse("https://crossplane.io/resource1")
					return &http.Response{
						Request:       req,
						StatusCode:    http.StatusOK,
						Body:          ioutil.NopCloser(strings.NewReader(errorResponse)),
						ContentLength: int64(len([]byte(inProgressResponse))),
					}, nil
				}),
			},
			want: want{
				op: &v1beta1.AsyncOperation{
					Method:       http.MethodPut,
					PollingURL:   pollingURL,
					Status:       errorStatus,
					ErrorMessage: errorMessage,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mina := name
			fmt.Println(mina)
			err := FetchAsyncOperation(context.Background(), tc.args.sender, tc.args.as)
			if diff := cmp.Diff(tc.want.err, err, test.EquateErrors()); diff != "" {
				t.Errorf("FetchAsyncOperation(...): -want error, +got error:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.op, tc.args.as); diff != "" {
				t.Errorf("FetchAsyncOperation(...): -want, +got:\n%s", diff)
			}
		})
	}

}
