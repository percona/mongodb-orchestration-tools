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
	"errors"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	DefaultHTTPTimeout   = "5s"
	DefaultSchedulerHost = "api.percona-mongo.marathon.l4lb.thisdcos.directory"
	ErrEmptyBody         = errors.New("got empty body")
	ErrNonSuccessCode    = errors.New("got non-success code")
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
}

// New creates a new ClientHTTP struct configured for use with the DC/OS SDK API
func New(frameworkName string, config *Config) *ClientHTTP {
	c := &ClientHTTP{
		FrameworkName: frameworkName,
		config:        config,
		scheme:        HTTPSchemePlain,
	}
	if config.Secure {
		c.scheme = HTTPSchemeSecure
	}
	return c
}

func (c *ClientHTTP) get(url string, out interface{}) error {
	req := fasthttp.AcquireRequest()
	req.SetRequestURI(url)
	req.Header.SetContentType("application/json")

	resp := fasthttp.AcquireResponse()
	client := &fasthttp.Client{}
	timeout := time.Now().Add(c.config.Timeout)
	client.DoDeadline(req, resp, timeout)

	if resp.StatusCode() != 200 {
		return ErrNonSuccessCode
	} else if len(resp.Body()) > 0 {
		return json.Unmarshal(resp.Body(), out)
	}
	return ErrEmptyBody
}
