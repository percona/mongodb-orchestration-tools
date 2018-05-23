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

package replset

import (
	"github.com/percona/dcos-mongo-tools/watchdog/config"
)

var testManagerReplset = &Replset{}

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

func (m *Manager) addReplset(rs *Replset) {
	if !m.HasReplset(rs.Name) {
		m.replsets[rs.Name] = rs
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
		m.addReplset(New(m.config, name))
	}
	m.Get(name).UpdateMember(mongod)
}

func (m *Manager) RemoveMember(mongod *Mongod) {
	name := mongod.Replset
	if m.HasReplset(name) {
		m.Get(name).RemoveMember(mongod)
	}
}
