// Copyright 2018 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	gotesting "testing"
	"time"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/stretchr/testify/assert"
)

var (
	testApi       = &ApiHttp{}
	testApiConfig = &Config{
		HostPrefix: DefaultHostPrefix,
		HostSuffix: DefaultHostSuffix,
		Timeout:    time.Second,
		Secure:     true,
	}
)

func TestApiNew(t *gotesting.T) {
	testApi = New(common.DefaultFrameworkName, testApiConfig)
	assert.Equal(t, testApiConfig, testApi.config, "api.config is incorrect")
	assert.Equal(t, common.DefaultFrameworkName, testApi.FrameworkName, "api.FrameworkName is incorrect")
	assert.Equal(t, testApiConfig.Timeout, testApi.client.Timeout, "api.client.Timeout is incorrect")
	assert.Equal(t, HttpSchemeSecure, testApi.scheme, "api.scheme is incorrect")

	testApiConfig.Secure = false
	testApi = New(common.DefaultFrameworkName, testApiConfig)
	assert.Equal(t, HttpSchemePlain, testApi.scheme, "api.scheme is incorrect")
}

func TestApiGetBaseURL(t *gotesting.T) {
	expected := DefaultHostPrefix + "." + common.DefaultFrameworkName + "." + DefaultHostSuffix
	assert.Equal(t, expected, testApi.getBaseURL(), "api.getBaseURL() is incorrect")
}

func TestApiGetPodURL(t *gotesting.T) {
	assert.Equal(t, testApi.GetPodURL(), testApi.scheme.String()+testApi.getBaseURL()+"/"+APIVersion+"/pod", "api.GetPodURL() is incorrect")
}

func TestApiGetEndpointsURL(t *gotesting.T) {
	assert.Equal(t, testApi.getEndpointsURL(), testApi.scheme.String()+testApi.getBaseURL()+"/"+APIVersion+"/endpoints", "api.GetEndpointsURL() is incorrect")
}
