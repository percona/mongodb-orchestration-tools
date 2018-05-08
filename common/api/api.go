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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

var (
	DefaultHostPrefix = "api"
	DefaultHostSuffix = "marathon.l4lb.thisdcos.directory"
	DefaultTimeout    = "5s"
)

type ApiScheme string

const (
	ApiSchemePlain  ApiScheme = "http://"
	ApiSchemeSecure ApiScheme = "https://"
)

func (s ApiScheme) String() string {
	return string(s)
}

type Config struct {
	HostPrefix string
	HostSuffix string
	Timeout    time.Duration
	Secure     bool
}

type Api interface {
	GetBaseUrl() string
	GetPodUrl() string
	GetPods() (*Pods, error)
	GetPodTasks(podName string) ([]PodTask, error)
	GetEndpointsUrl() string
	GetEndpoints() (*Endpoints, error)
	GetEndpoint(endpointName string) (*Endpoint, error)
}

type ApiHttp struct {
	FrameworkName string
	config        *Config
	scheme        ApiScheme
	client        *http.Client
}

func New(frameworkName string, config *Config) *ApiHttp {
	a := &ApiHttp{
		FrameworkName: frameworkName,
		config:        config,
		scheme:        ApiSchemePlain,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
	if config.Secure {
		a.scheme = ApiSchemeSecure
	}
	return a
}

func (a *ApiHttp) GetBaseUrl() string {
	return a.config.HostPrefix + "." + a.FrameworkName + "." + a.config.HostSuffix
}

func (a *ApiHttp) GetPodUrl() string {
	return a.scheme.String() + a.GetBaseUrl() + "/v1/pod"
}

func (a *ApiHttp) GetPods() (*Pods, error) {
	pods := &Pods{}
	err := a.get(a.GetPodUrl(), pods)
	return pods, err
}

func (a *ApiHttp) GetPodTasks(podName string) ([]PodTask, error) {
	podUrl := a.GetPodUrl() + "/" + podName + "/info"
	var tasks []PodTask
	err := a.get(podUrl, &tasks)
	return tasks, err
}

func (a *ApiHttp) GetEndpointsUrl() string {
	return a.scheme.String() + a.GetBaseUrl() + "/v1/endpoints"
}

func (a *ApiHttp) GetEndpoints() (*Endpoints, error) {
	endpoints := &Endpoints{}
	err := a.get(a.GetEndpointsUrl(), endpoints)
	return endpoints, err
}

func (a *ApiHttp) GetEndpoint(endpointName string) (*Endpoint, error) {
	endpointUrl := a.GetEndpointsUrl() + "/" + endpointName
	endpoint := &Endpoint{}
	err := a.get(endpointUrl, endpoint)
	return endpoint, err
}

func (a *ApiHttp) get(url string, out interface{}) error {
	resp, err := a.client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, out)
}
