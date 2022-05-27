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

package configuration

import (
	"testing"

	"github.com/Azure/azure-sdk-for-go/services/mysql/mgmt/2017-12-01/mysql"
	"github.com/google/go-cmp/cmp"

	azuredbv1beta1 "github.com/crossplane-contrib/provider-azure/apis/database/v1beta1"
)

type mysqlConfigurationModifier func(configuration *mysql.Configuration)

func mysqlConfigurationWithValue(v *string) mysqlConfigurationModifier {
	return func(configuration *mysql.Configuration) {
		configuration.Value = v
	}
}

func mysqlConfiguration(cm ...mysqlConfigurationModifier) *mysql.Configuration {
	c := &mysql.Configuration{
		ConfigurationProperties: &mysql.ConfigurationProperties{},
	}
	for _, m := range cm {
		m(c)
	}
	return c
}

func TestIsMySQLConfigurationUpToDate(t *testing.T) {
	val1, val2 := testValue1, testValue2
	type args struct {
		p  azuredbv1beta1.SQLServerConfigurationParameters
		in mysql.Configuration
	}
	tests := map[string]struct {
		args args
		want bool
	}{
		"UpToDate": {
			args: args{
				p: *sqlServerConfigurationParameters(
					sqlServerConfigurationParametersWithValue(&val1)),
				in: *mysqlConfiguration(mysqlConfigurationWithValue(&val1)),
			},
			want: true,
		},
		"NeedsUpdate": {
			args: args{
				p: *sqlServerConfigurationParameters(
					sqlServerConfigurationParametersWithValue(&val1)),
				in: *mysqlConfiguration(mysqlConfigurationWithValue(&val2)),
			},
			want: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := IsMySQLConfigurationUpToDate(tt.args.p, tt.args.in)
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("IsMySQLConfigurationUpToDate(...): -want, +got\n%s", diff)
			}
		})
	}
}
