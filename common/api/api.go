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
	"time"
)

// APIVersion is the version of the DC/OS SDK API
var APIVersion = "v1"

// Config is a struct of configuration options for the API
type Config struct {
	Host    string
	Timeout time.Duration
	Secure  bool
}

// Client is an interface describing a DC/OS SDK API Client
type Client interface {
	GetPodURL() string
	GetPods() (*Pods, error)
	GetPodTasks(podName string) ([]PodTask, error)
	GetEndpoints() (*Endpoints, error)
	GetEndpoint(endpointName string) (*Endpoint, error)
}
