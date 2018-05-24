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
	testAPI       = &APIHttp{}
	testAPIConfig = &Config{
		HostPrefix: DefaultHostPrefix,
		HostSuffix: DefaultHostSuffix,
		Timeout:    time.Second,
		Secure:     true,
	}
)

func TestAPINew(t *gotesting.T) {
	testAPI = New(common.DefaultFrameworkName, testAPIConfig)
	assert.Equal(t, testAPIConfig, testAPI.config, "api.config is incorrect")
	assert.Equal(t, common.DefaultFrameworkName, testAPI.FrameworkName, "api.FrameworkName is incorrect")
	assert.Equal(t, testAPIConfig.Timeout, testAPI.client.Timeout, "api.client.Timeout is incorrect")
	assert.Equal(t, HttpSchemeSecure, testAPI.scheme, "api.scheme is incorrect")

	testAPIConfig.Secure = false
	testAPI = New(common.DefaultFrameworkName, testAPIConfig)
	assert.Equal(t, HttpSchemePlain, testAPI.scheme, "api.scheme is incorrect")
}

func TestAPIGetBaseURL(t *gotesting.T) {
	expected := DefaultHostPrefix + "." + common.DefaultFrameworkName + "." + DefaultHostSuffix
	assert.Equal(t, expected, testAPI.getBaseURL(), "api.getBaseURL() is incorrect")
}

func TestAPIGetPodURL(t *gotesting.T) {
	assert.Equal(t, testAPI.GetPodURL(), testAPI.scheme.String()+testAPI.getBaseURL()+"/"+APIVersion+"/pod", "api.GetPodURL() is incorrect")
}

func TestAPIGetEndpointsURL(t *gotesting.T) {
	assert.Equal(t, testAPI.getEndpointsURL(), testAPI.scheme.String()+testAPI.getBaseURL()+"/"+APIVersion+"/endpoints", "api.GetEndpointsURL() is incorrect")
}
