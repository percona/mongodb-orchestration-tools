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
	DefaultHTTPHostPrefix = "api"
	DefaultHTTPHostSuffix = "marathon.l4lb.thisdcos.directory"
	DefaultHTTPTimeout    = "5s"
)

// HTTPScheme is the scheme type to be used for HTTP calls
type HTTPScheme string

const (
	HTTPSchemePlain  HTTPScheme = "http://"
	HTTPSchemeSecure HTTPScheme = "https://"
)

// String returns a string representation of the HTTPScheme
func (s HTTPScheme) String() string {
	return string(s)
}

// ClientHTTP is an HTTP client for the DC/OS SDK API
type ClientHTTP struct {
	FrameworkName string
	config        *Config
	scheme        HTTPScheme
	client        *http.Client
}

// New creates a new ClientHTTP struct configured for use with the DC/OS SDK API
func New(frameworkName string, config *Config) *ClientHTTP {
	c := &ClientHTTP{
		FrameworkName: frameworkName,
		config:        config,
		scheme:        HTTPSchemePlain,
		client: &http.Client{
			Timeout: config.Timeout,
		},
	}
	if config.Secure {
		c.scheme = HTTPSchemeSecure
	}
	return c
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

func (c *ClientHTTP) getBaseURL() string {
	return c.config.HostPrefix + "." + c.FrameworkName + "." + c.config.HostSuffix

}
