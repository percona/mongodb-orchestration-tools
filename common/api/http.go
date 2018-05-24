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
)

var (
	DefaultHostPrefix = "api"
	DefaultHostSuffix = "marathon.l4lb.thisdcos.directory"
	DefaultTimeout    = "5s"
)

type HttpScheme string

const (
	HttpSchemePlain  HttpScheme = "http://"
	HttpSchemeSecure HttpScheme = "https://"
)

func (s HttpScheme) String() string {
	return string(s)
}

type ApiHttp struct {
	FrameworkName string
	config        *Config
	scheme        HttpScheme
	client        *http.Client
}

func New(frameworkName string, config *Config) *ApiHttp {
	a := &ApiHttp{
		FrameworkName: frameworkName,
		config:        config,
		scheme:        HttpSchemePlain,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
	if config.Secure {
		a.scheme = HttpSchemeSecure
	}
	return a
}

func (a *ApiHttp) getBaseURL() string {
	return a.config.HostPrefix + "." + a.FrameworkName + "." + a.config.HostSuffix
}

func (a *ApiHttp) GetPodURL() string {
	return a.scheme.String() + a.getBaseURL() + "/" + APIVersion + "/pod"
}

func (a *ApiHttp) GetPods() (*Pods, error) {
	pods := &Pods{}
	err := a.get(a.GetPodURL(), pods)
	return pods, err
}

func (a *ApiHttp) GetPodTasks(podName string) ([]*PodTask, error) {
	podURL := a.GetPodURL() + "/" + podName + "/info"
	var tasks []*PodTask
	err := a.get(podURL, &tasks)
	return tasks, err
}

func (a *ApiHttp) getEndpointsURL() string {
	return a.scheme.String() + a.getBaseURL() + "/" + APIVersion + "/endpoints"
}

func (a *ApiHttp) GetEndpoints() (*Endpoints, error) {
	endpoints := &Endpoints{}
	err := a.get(a.getEndpointsURL(), endpoints)
	return endpoints, err
}

func (a *ApiHttp) GetEndpoint(endpointName string) (*Endpoint, error) {
	endpointURL := a.getEndpointsURL() + "/" + endpointName
	endpoint := &Endpoint{}
	err := a.get(endpointURL, endpoint)
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
