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
	config   *config.Config
	stop     *chan bool
	watchers map[string]*Watcher
}

func NewManager(config *config.Config, stop *chan bool) *Manager {
	return &Manager{
		config:   config,
		stop:     stop,
		watchers: make(map[string]*Watcher),
	}
}

func (m *Manager) HasWatcher(rsName string) bool {
	if _, ok := m.watchers[rsName]; ok {
		return true
	}
	return false
}

func (m *Manager) Watch(rs *replset.Replset) {
	if !m.HasWatcher(rs.Name) {
		log.WithFields(log.Fields{
			"replset": rs.Name,
		}).Info("Starting replset watcher")

		m.watchers[rs.Name] = New(
			rs,
			m.config,
			m.stop,
		)
		go m.watchers[rs.Name].Run()
	}
}

func (m *Manager) Get(rsName string) *Watcher {
	if !m.HasWatcher(rsName) {
		return nil
	}
	return m.watchers[rsName]
}
