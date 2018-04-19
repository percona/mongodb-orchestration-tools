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

package metrics

import (
	"github.com/percona/dcos-mongo-tools/common/command"
	log "github.com/sirupsen/logrus"
)

const (
	mgoStatsdConfigUpdateInterval = "0"
)

type Metrics struct {
	config  *Config
	running bool
}

func New(config *Config) *Metrics {
	return &Metrics{
		config: config,
	}
}

func (m *Metrics) Name() string {
	return "DC/OS Metrics"
}

func (m *Metrics) DoRun() bool {
	return m.config.Enabled
}

func (m *Metrics) IsRunning() bool {
	return m.running
}

func (m *Metrics) Run() error {
	if m.DoRun() == false {
		log.Warn("DC/OS Metrics client executor disabled! Skipping start")
		return nil
	}

	cmd, err := command.New(
		m.config.MgoStatsdBin,
		[]string{
			"-configUpdateInterval", mgoStatsdConfigUpdateInterval,
			"-config", m.config.MgoStatsdConfigFile,
		},
		m.config.User,
		m.config.Group,
	)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"binary": m.config.MgoStatsdBin,
		"config": m.config.MgoStatsdConfigFile,
	}).Info("Starting DC/OS Metrics client executor")

	m.running = true
	err = cmd.Start()
	if err != nil {
		m.running = false
		return err
	}

	log.Info("Completed DC/OS Metrics client executor")

	return nil
}
