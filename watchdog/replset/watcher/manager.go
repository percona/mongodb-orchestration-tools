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

package watcher

import (
	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"github.com/percona/dcos-mongo-tools/watchdog/replset"
	log "github.com/sirupsen/logrus"
)

type Manager struct {
	config    *config.Config
	watchers  map[string]*Watcher
	stopChans map[string]chan bool
}

func NewManager(config *config.Config) *Manager {
	return &Manager{
		config:    config,
		watchers:  make(map[string]*Watcher),
		stopChans: make(map[string]chan bool),
	}
}

func (m *Manager) HasWatcher(rs *replset.Replset) bool {
	if _, ok := m.watchers[rs.Name]; ok {
		return true
	}
	return false
}

func (m *Manager) Watch(rs *replset.Replset) {
	if !m.HasWatcher(rs) {
		log.WithFields(log.Fields{
			"replset": rs.Name,
		}).Info("Starting replset watcher")
		m.stopChans[rs.Name] = make(chan bool)
		m.watchers[rs.Name] = New(
			rs,
			m.config,
			m.stopChans[rs.Name],
		)
		go m.watchers[rs.Name].Run()
	}
}

func (m *Manager) Stop() {
	for _, val := range m.stopChans {
		val <- true
	}
}
