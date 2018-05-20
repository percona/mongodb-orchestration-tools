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

package mock

import (
	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/common/testing"
)

var EndpointName = "mongo-port"

type API struct{}

func New() *API {
	return &API{}
}

func (a *API) GetPodUrl() string {
	return "http://localhost/v1/pod"
}

func (a *API) GetPods() (*api.Pods, error) {
	return &api.Pods{}, nil
}

func (a *API) GetPodTasks(podName string) ([]api.PodTask, error) {
	return []api.PodTask{}, nil
}

func (a *API) GetEndpointsUrl() string {
	return "http://localhost/v1/endpoints"
}

func (a *API) GetEndpoints() (*api.Endpoints, error) {
	return &api.Endpoints{EndpointName}, nil
}

func (a *API) GetEndpoint(endpointName string) (*api.Endpoint, error) {
	if endpointName == EndpointName {
		return &api.Endpoint{
			Address: []string{testing.MongodbHost + ":" + testing.MongodbPrimaryPort},
			Dns:     []string{testing.MongodbHostname + ":" + testing.MongodbPrimaryPort},
		}, nil
	}
	return &api.Endpoint{}, nil
}
