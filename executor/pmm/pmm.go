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
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	pmmAdminCommand  = "pmm-admin"
	pmmAdminRunUser  = "root"
	pmmAdminRunGroup = "root"
	qanServiceName   = "mongodb:queries"
)

type PMM struct {
	config        *Config
	configFile    string
	frameworkName string
	maxRetries    uint
	retrySleep    time.Duration
	user          *user.User
	group         *user.Group
	running       bool
}

func New(config *Config, frameworkName string) (*PMM, error) {
	pmmUser, err := user.Lookup(pmmAdminRunUser)
	if err != nil {
		return nil, err
	}

	pmmGroup, err := user.LookupGroup(pmmAdminRunGroup)
	if err != nil {
		return nil, err
	}

	return &PMM{
		config:        config,
		configFile:    filepath.Join(config.ConfigDir, "pmm.yml"),
		frameworkName: frameworkName,
		maxRetries:    5,
		retrySleep:    time.Duration(5) * time.Second,
		user:          pmmUser,
		group:         pmmGroup,
	}, nil
}

func (p *PMM) Name() string {
	return "Percona PMM"
}

func (p *PMM) DoRun() bool {
	return p.config.Enabled
}

func (p *PMM) IsRunning() bool {
	return p.running
}

func (p *PMM) doStartQueryAnalytics() bool {
	return p.config.EnableQueryAnalytics
}

func (p *PMM) startMetrics() error {
	list, err := p.list()
	if err != nil {
		log.Errorf("Got error listing PMM services: %s", err)
		return err
	}

	log.WithFields(log.Fields{
		"max_retries":  p.maxRetries,
		"linux_port":   p.config.LinuxMetricsExporterPort,
		"mongodb_port": p.config.MongoDBMetricsExporterPort,
	}).Info("Starting PMM metrics services")

	services := []*Service{
		NewService(
			p.configFile,
			"linux:metrics",
			p.config.LinuxMetricsExporterPort,
			[]string{},
			p.user,
			p.group,
		),
		NewService(
			p.configFile,
			"mongodb:metrics",
			p.config.MongoDBMetricsExporterPort,
			[]string{
				"--cluster=" + p.frameworkName,
				"--uri=" + p.config.DB.Uri(),
			},
			p.user,
			p.group,
		),
	}

	for _, service := range services {
		if list != nil && list.hasService(service.Name) {
			log.Warnf("Service %s already added! Skipping", service.Name)
			continue
		}
		err := service.addWithRetry(p.maxRetries, p.retrySleep)
		if err != nil {
			log.Errorf("Could not add PMM service %s after %d retries: %s", service.Name, p.maxRetries, err)
		}
	}

	return nil
}

func (p *PMM) startQueryAnalytics() error {
	list, err := p.list()
	if err != nil {
		log.Errorf("Got error listing PMM services: %s", err)
		return err
	}
	if list != nil && list.hasService(qanServiceName) {
		log.Warnf("Service %s already added! Skipping", qanServiceName)
		return nil
	}

	service := NewService(
		p.configFile,
		qanServiceName,
		0,
		[]string{"--uri=" + p.config.DB.Uri()},
		p.user,
		p.group,
	)

	log.WithFields(log.Fields{
		"max_retries": p.maxRetries,
	}).Info("Starting PMM Query Analytics (QAN) agent service")

	return service.addWithRetry(p.maxRetries, p.retrySleep)
}

func (p *PMM) Run(quit *chan bool) {
	if p.DoRun() == false {
		log.Warn("PMM client executor disabled! Skipping start")
		return
	}

	log.WithFields(log.Fields{
		"config": p.configFile,
	}).Info("Starting PMM client executor")
	p.running = true

	err := p.repair()
	if err != nil {
		log.Errorf("Error repairing PMM services: %s", err)
		return
	}

	err = p.startMetrics()
	if err != nil {
		log.Errorf("PMM metrics services did not start: %s", err)
		return
	}

	if p.doStartQueryAnalytics() {
		err = p.startQueryAnalytics()
		if err != nil {
			log.Errorf("Could not enable PMM Query Analytics (QAN) agent service: %s", err)
			return
		}
	} else {
		log.Info("PMM Query Analytics (QAN) disabled, skipping")
	}

	p.running = false
	log.Info("Completed PMM client executor")
}
