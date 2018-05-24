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
	"errors"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/common/testing"
)

func (a *API) GetEndpointsUrl() string {
	return "http://localhost/" + api.APIVersion + "/endpoints"
}

func (a *API) GetEndpoints() (*api.Endpoints, error) {
	if SimulateError {
		return nil, errors.New("simulating a .GetEndpoints() error")
	}
	return &api.Endpoints{common.DefaultMongoDBMongodEndpointName}, nil
}

func (a *API) GetEndpoint(endpointName string) (*api.Endpoint, error) {
	if SimulateError {
		return nil, errors.New("simulating a .GetEndpoint() error")
	}
	if !testing.Enabled() || endpointName != common.DefaultMongoDBMongodEndpointName {
		return &api.Endpoint{}, nil
	}
	return &api.Endpoint{
		Address: []string{
			testing.MongodbHost + ":" + testing.MongodbPrimaryPort,
			testing.MongodbHost + ":" + testing.MongodbSecondary1Port,
			testing.MongodbHost + ":" + testing.MongodbSecondary2Port,
		},
		Dns: []string{
			testing.MongodbHostname + ":" + testing.MongodbPrimaryPort,
			testing.MongodbHostname + ":" + testing.MongodbSecondary1Port,
			testing.MongodbHostname + ":" + testing.MongodbSecondary2Port,
		},
	}, nil
}
