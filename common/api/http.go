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

type ClientHTTP struct {
	FrameworkName string
	config        *Config
	scheme        HttpScheme
	client        *http.Client
}

func New(frameworkName string, config *Config) *ClientHTTP {
	c := &ClientHTTP{
		FrameworkName: frameworkName,
		config:        config,
		scheme:        HttpSchemePlain,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
	if config.Secure {
		c.scheme = HttpSchemeSecure
	}
	return c
}

func (c *ClientHTTP) getBaseURL() string {
	return c.config.HostPrefix + "." + c.FrameworkName + "." + c.config.HostSuffix
}

func (c *ClientHTTP) GetPodURL() string {
	return c.scheme.String() + c.getBaseURL() + "/" + APIVersion + "/pod"
}

func (c *ClientHTTP) GetPods() (*Pods, error) {
	pods := &Pods{}
	err := c.get(c.GetPodURL(), pods)
	return pods, err
}

func (c *ClientHTTP) GetPodTasks(podName string) ([]*PodTask, error) {
	podURL := c.GetPodURL() + "/" + podName + "/info"
	var tasks []*PodTask
	err := c.get(podURL, &tasks)
	return tasks, err
}

func (c *ClientHTTP) getEndpointsURL() string {
	return c.scheme.String() + c.getBaseURL() + "/" + APIVersion + "/endpoints"
}

func (c *ClientHTTP) GetEndpoints() (*Endpoints, error) {
	endpoints := &Endpoints{}
	err := c.get(c.getEndpointsURL(), endpoints)
	return endpoints, err
}

func (c *ClientHTTP) GetEndpoint(endpointName string) (*Endpoint, error) {
	endpointURL := c.getEndpointsURL() + "/" + endpointName
	endpoint := &Endpoint{}
	err := c.get(endpointURL, endpoint)
	return endpoint, err
}

func (c *ClientHTTP) get(url string, out interface{}) error {
	resp, err := c.client.Get(url)
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
