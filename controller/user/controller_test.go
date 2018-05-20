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

package user

import (
	gotesting "testing"

	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

type MockApi struct{}

func (a *MockApi) GetPodUrl() string {
	return "http://localhost/v1/pod"
}

func (a *MockApi) GetPods() (*api.Pods, error) {
	return &api.Pods{}, nil
}

func (a *MockApi) GetPodTasks(podName string) ([]api.PodTask, error) {
	return []api.PodTask{}, nil
}

func (a *MockApi) GetEndpointsUrl() string {
	return "http://localhost/v1/endpoints"
}

func (a *MockApi) GetEndpoints() (*api.Endpoints, error) {
	return &api.Endpoints{"mongo-port"}, nil
}

func (a *MockApi) GetEndpoint(endpointName string) (*api.Endpoint, error) {
	if endpointName == "mongo-port" {
		return &api.Endpoint{
			Address: []string{testing.MongodbHost + ":" + testing.MongodbPrimaryPort},
			Dns:     []string{testing.MongodbHostname + ":" + testing.MongodbPrimaryPort},
		}, nil
	}
	return &api.Endpoint{}, nil
}

func TestControllerUserNew(t *gotesting.T) {
	testing.DoSkipTest(t)

	var err error
	testController, err = NewController(testControllerConfig, &MockApi{})
	assert.NoError(t, err, ".NewController() should not return an error")
	assert.NotNil(t, testController, ".NewController() should return a Controller that is not nil")
	assert.NotNil(t, testController.session, ".NewController() should return a Controller with a session field that is not nil")
	assert.NoError(t, testController.session.Ping(), ".NewController() should return a Controller with a session that is pingable")
	assert.Equal(t, mgo.Primary, testController.session.Mode(), ".NewController() should return a Controller with a session that is in mgo.Primary mode")
}

func TestControllerUserClose(t *gotesting.T) {
	testing.DoSkipTest(t)

	testController.Close()
	assert.Nil(t, testController.session, "Controller session should not nil after .Close()")
}
