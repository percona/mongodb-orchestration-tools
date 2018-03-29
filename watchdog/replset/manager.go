package replset

import (
	"github.com/percona/dcos-mongo-tools/watchdog/config"
)

type Manager struct {
	config   *config.Config
	replsets map[string]*Replset
}

func NewManager(config *config.Config) *Manager {
	return &Manager{
		config:   config,
		replsets: make(map[string]*Replset),
	}
}

func (m *Manager) HasReplset(name string) bool {
	if _, ok := m.replsets[name]; ok {
		return true
	}
	return false
}

func (m *Manager) Add(name string) {
	if !m.HasReplset(name) {
		m.replsets[name] = New(m.config, name)
	}
}

func (m *Manager) GetAll() []*Replset {
	rss := make([]*Replset, 0)
	for _, rs := range m.replsets {
		rss = append(rss, rs)
	}
	return rss
}

func (m *Manager) Get(name string) *Replset {
	if m.HasReplset(name) {
		return m.replsets[name]
	}
	return nil
}

func (m *Manager) HasMember(mongod *Mongod) bool {
	name := mongod.Replset
	if m.HasReplset(name) {
		return m.Get(name).HasMember(mongod.Name())
	}
	return false
}

func (m *Manager) UpdateMember(mongod *Mongod) {
	name := mongod.Replset
	if !m.HasReplset(name) {
		m.Add(name)
	}
	m.Get(name).UpdateMember(mongod)
}

func (m *Manager) RemoveMember(mongod *Mongod) {
	name := mongod.Replset
	if m.HasReplset(name) {
		m.Get(name).RemoveMember(mongod)
	}
}
