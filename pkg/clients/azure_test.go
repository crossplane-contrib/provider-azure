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

package azure

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Azure/go-autorest/autorest"
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"

	"github.com/crossplane/crossplane-runtime/pkg/test"

	"github.com/crossplane-contrib/provider-azure/apis/v1alpha3"
)

const (
	authData = `{
		"clientId": "0f32e96b-b9a4-49ce-a857-243a33b20e5c",
		"clientSecret": "49d8cab5-d47a-4d1a-9133-5c5db29c345d",
		"subscriptionId": "bf1b0e59-93da-42e0-82c6-5a1d94227911",
		"tenantId": "302de427-dba9-4452-8583-a4268e46de6b",
		"activeDirectoryEndpointUrl": "https://login.microsoftonline.com",
		"resourceManagerEndpointUrl": "https://management.azure.com/",
		"activeDirectoryGraphResourceId": "https://graph.windows.net/",
		"sqlManagementEndpointUrl": "https://management.core.windows.net:8443/",
		"galleryEndpointUrl": "https://gallery.azure.com/",
		"managementEndpointUrl": "https://management.core.windows.net/"
}`
)

func TestNewClient(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	client, err := NewClient([]byte(authData))
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(client).NotTo(gomega.BeNil())
	g.Expect(client.SubscriptionID).To(gomega.Equal("bf1b0e59-93da-42e0-82c6-5a1d94227911"))
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
		as     *v1alpha3.AsyncOperation
	}
	type want struct {
		op  *v1alpha3.AsyncOperation
		err error
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"NoOperation": {},
		"InProgress": {
			args: args{
				as: &v1alpha3.AsyncOperation{
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
				op: &v1alpha3.AsyncOperation{
					Method:     http.MethodPut,
					PollingURL: pollingURL,
					Status:     inprogressStatus,
				},
			},
		},
		"Failure": {
			args: args{
				as: &v1alpha3.AsyncOperation{
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
				op: &v1alpha3.AsyncOperation{
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

func TestIsNotFound(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	cases := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{autorest.DetailedError{}, false},
		{autorest.DetailedError{StatusCode: http.StatusNotFound}, true},
	}

	for _, tt := range cases {
		actual := IsNotFound(tt.err)
		g.Expect(actual).To(gomega.Equal(tt.expected))
	}
}

func TestStringHelpers(t *testing.T) {
	t.Run("ToStringMap", func(t *testing.T) {
		original := make(map[string]*string)

		original["a"] = nil
		original["b"] = ToStringPtr("hello")
		original["c"] = ToStringPtr("")

		result := ToStringMap(original)

		if diff := cmp.Diff("", result["a"], test.EquateErrors()); diff != "" {
			t.Errorf("ToStringMap(...): -want error, +got error:\n%s", diff)
		}
		if diff := cmp.Diff("hello", result["b"], test.EquateErrors()); diff != "" {
			t.Errorf("ToStringMap(...): -want error, +got error:\n%s", diff)
		}
		if diff := cmp.Diff("", result["c"], test.EquateErrors()); diff != "" {
			t.Errorf("ToStringMap(...): -want error, +got error:\n%s", diff)
		}
	})

	t.Run("ToStringPtrMap", func(t *testing.T) {
		original := make(map[string]string)

		original["a"] = ""
		original["b"] = "hello"

		result := ToStringPtrMap(original)

		if diff := cmp.Diff("", *result["a"], test.EquateErrors()); diff != "" {
			t.Errorf("ToStringPtrMap(...): -want error, +got error:\n%s", diff)
		}
		if diff := cmp.Diff("hello", *result["b"], test.EquateErrors()); diff != "" {
			t.Errorf("ToStringPtrMap(...): -want error, +got error:\n%s", diff)
		}
	})

	t.Run("ToString", func(t *testing.T) {
		hello := "hello"
		empty := ""

		if diff := cmp.Diff("", ToString(nil), test.EquateErrors()); diff != "" {
			t.Errorf("ToString(...): -want error, +got error:\n%s", diff)
		}
		if diff := cmp.Diff("hello", ToString(&hello), test.EquateErrors()); diff != "" {
			t.Errorf("ToString(...): -want error, +got error:\n%s", diff)
		}
		if diff := cmp.Diff("", ToString(&empty), test.EquateErrors()); diff != "" {
			t.Errorf("ToString(...): -want error, +got error:\n%s", diff)
		}
	})

	t.Run("ToStringPtr", func(t *testing.T) {
		if diff := cmp.Diff("", *ToStringPtr("", FieldRequired), test.EquateErrors()); diff != "" {
			t.Errorf("ToStringPtr(...): -want error, +got error:\n%s", diff)
		}
		if diff := cmp.Diff("hello", *ToStringPtr("hello"), test.EquateErrors()); diff != "" {
			t.Errorf("ToStringPtr(...): -want error, +got error:\n%s", diff)
		}
		if diff := cmp.Diff((*string)(nil), ToStringPtr(""), test.EquateErrors()); diff != "" {
			t.Errorf("ToStringPtr(...): -want error, +got error:\n%s", diff)
		}
	})
}
