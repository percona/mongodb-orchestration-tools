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
