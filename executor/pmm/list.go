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

package pmm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/percona/dcos-mongo-tools/common/command"
	log "github.com/sirupsen/logrus"
)

var (
	ErrMsgNoServices = "No services under monitoring."
)

type ListClientService struct {
	Name     string `json:"Name"`
	Running  bool   `json:"Running"`
	Type     string `json:"Type"`
	Port     string `json:"Port"`
	Password string `json:"Password,omitempty"`
	DSN      string `json:"DSN"`
	Options  string `json:"Options,omitempty"`
	SSL      string `json:"SSL,omitempty"`
}

type ListClient struct {
	ClientName    string               `json:"ClientName"`
	ServerAddress string               `json:"ServerAddress"`
	Services      []*ListClientService `json:"Services"`
	Err           string               `json:"Err,omitempty"`
}

func (pl *ListClient) getError() error {
	if pl.Err != "" {
		return fmt.Errorf("PMM error: %s\n", strings.TrimSpace(pl.Err))
	}
	return nil
}

func (pl *ListClient) hasError() bool {
	err := pl.getError()
	return err != nil && err.Error() != ErrMsgNoServices
}

func (pl *ListClient) hasService(serviceName string) bool {
	for _, service := range pl.Services {
		if service.Type == serviceName && service.Running {
			return true
		}
	}
	return false
}

func (p *PMM) list() (*ListClient, error) {
	log.Info("Listing PMM services")

	cmd, err := command.New(
		pmmAdminCommand,
		[]string{"list", "--json", "--config-file=" + p.configFile},
		p.user,
		p.group,
	)
	if err != nil {
		return nil, err
	}

	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	list := &ListClient{}
	err = json.Unmarshal(bytes, list)
	if err != nil {
		return nil, err
	}
	if list.hasError() {
		return nil, list.getError()
	}

	return list, nil
}
