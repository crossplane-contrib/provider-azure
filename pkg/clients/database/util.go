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
	"github.com/Azure/go-autorest/autorest/date"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-azure/apis/database/v1beta1"
)

// Get a pointer to a CreateMode
func pointerFromCreateMode(createMode v1beta1.CreateMode) *v1beta1.CreateMode {
	result := createMode
	return &result
}

// Get a CreateMode from a pointer-to-CreateMode
func pointerToCreateMode(mode *v1beta1.CreateMode) v1beta1.CreateMode {
	if nil == mode {
		return v1beta1.CreateModeDefault
	}
	return *mode
}

// Convert a possibly nil metav1.Time to a possibly nil date.Time
func safeDate(time *metav1.Time) *date.Time {
	if time == nil {
		return nil
	}
	return &date.Time{Time: time.Time}
}
