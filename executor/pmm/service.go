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
	"os/user"
	"strconv"
	"strings"
	"time"

	"github.com/percona/dcos-mongo-tools/common/command"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	Name       string
	configFile string
	port       uint
	args       []string
	user       *user.User
	group      *user.Group
}

func NewService(configFile string, name string, port uint, args []string, runUser *user.User, runGroup *user.Group) *Service {
	return &Service{
		Name:       name,
		configFile: configFile,
		port:       port,
		args:       args,
		user:       runUser,
		group:      runGroup,
	}
}

func (s *Service) add() error {
	args := append(
		[]string{"add"},
		s.Name,
		"--config-file="+s.configFile,
	)
	if int(s.port) > 0 {
		args = append(args, "--service-port="+strconv.Itoa(int(s.port)))
	}
	args = append(args, s.args...)

	cmd, err := command.New(
		pmmAdminCommand,
		args,
		s.user,
		s.group,
	)
	if err != nil {
		return err
	}

	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	if err != nil {
		log.Errorf("Failed to add PMM service %s! Error: '%s'", s.Name, trimmed)
		return err
	}
	log.Infof("Added PMM service %s, pmm-admin out: '%s'", s.Name, trimmed)

	return nil
}

func (s *Service) addWithRetry(maxRetries uint, retrySleep time.Duration) error {
	var err error
	var tries uint
	for tries <= maxRetries {
		err = s.add()
		if err == nil {
			return nil
		}
		log.Errorf("Received error adding PMM service %s: %s, retrying", s.Name, err)
		time.Sleep(retrySleep)
		tries += 1
	}
	return err
}

func (p *PMM) repair() error {
	log.Info("Repairing all PMM client services")

	cmd, err := command.New(
		pmmAdminCommand,
		[]string{"repair", "--config-file=" + p.configFile},
		p.user,
		p.group,
	)
	if err != nil {
		return err
	}

	return cmd.Run()
}
