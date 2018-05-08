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

	"github.com/stretchr/testify/assert"
)

var (
	testApi       = &ApiHttp{}
	testApiConfig = &Config{
		HostPrefix: "api",
		HostSuffix: "test.com",
		Timeout:    time.Second,
		Secure:     true,
	}
	testApiFrameworkName = "percona-mongo"
)

func TestApiNew(t *gotesting.T) {
	testApi = New(testApiFrameworkName, testApiConfig)
	assert.Equal(t, testApiConfig, testApi.config, "api.config is incorrect")
	assert.Equal(t, testApiFrameworkName, testApi.FrameworkName, "api.FrameworkName is incorrect")
	assert.Equal(t, testApiConfig.Timeout, testApi.client.Timeout, "api.client.Timeout is incorrect")
	assert.Equal(t, HttpSchemeSecure, testApi.scheme, "api.scheme is incorrect")

	testApiConfig.Secure = false
	testApi = New(testApiFrameworkName, testApiConfig)
	assert.Equal(t, HttpSchemePlain, testApi.scheme, "api.scheme is incorrect")
}

func TestApiGetBaseUrl(t *gotesting.T) {
	assert.Equal(t, testApi.GetBaseUrl(), "api.percona-mongo.test.com", "api.GetBaseUrl() is incorrect")
}

func TestApiGetPodUrl(t *gotesting.T) {
	assert.Equal(t, testApi.GetPodUrl(), "http://"+testApi.GetBaseUrl()+"/v1/pod", "api.GetPodUrl() is incorrect")
}

func TestApiGetEndpointsUrl(t *gotesting.T) {
	assert.Equal(t, testApi.GetEndpointsUrl(), "http://"+testApi.GetBaseUrl()+"/v1/endpoints", "api.GetEndpointsUrl() is incorrect")
}
