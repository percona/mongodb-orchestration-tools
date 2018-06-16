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
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	rsStatus "github.com/timvaillancourt/go-mongodb-replset/status"
	"gopkg.in/mgo.v2"
)

const (
	frameworkTagName = "dcosFramework"
)

// State is a struct reflecting the state of a MongoDB Replica Set
type State struct {
	sync.Mutex
	Replset  string
	config   *rsConfig.Config
	status   *rsStatus.Status
	doUpdate bool
}

// NewState returns a new State struct
func NewState(replset string) *State {
	return &State{
		Replset: replset,
	}
}

func (s *State) fetchConfig(configManager rsConfig.Manager) error {
	err := configManager.Load()
	if err != nil {
		return err
	}

	s.config = configManager.Get()
	return nil
}

func (s *State) fetchStatus(session *mgo.Session) error {
	status, err := rsStatus.New(session)
	if err != nil {
		return err
	}

	s.status = status
	return nil
}

// Fetch gets the current MongoDB Replica Set status and config while locking the State
func (s *State) Fetch(session *mgo.Session, configManager rsConfig.Manager) error {
	s.Lock()
	defer s.Unlock()

	log.WithFields(log.Fields{
		"replset": s.Replset,
	}).Info("Updating replset config and status")

	err := s.fetchConfig(configManager)
	if err != nil {
		return err
	}

	return s.fetchStatus(session)
}

func (s *State) updateConfig(configManager rsConfig.Manager) error {
	if s.doUpdate == false {
		return nil
	}

	configManager.IncrVersion()
	config := configManager.Get()
	log.WithFields(log.Fields{
		"replset":        s.Replset,
		"config_version": config.Version,
	}).Info("Writing new replset config")
	fmt.Println(config)

	err := configManager.Save()
	if err != nil {
		return err
	}
	s.doUpdate = false

	return nil
}

// AddConfigMembers adds members to the MongoDB Replica Set config
func (s *State) AddConfigMembers(session *mgo.Session, configManager rsConfig.Manager, members []*Mongod) {
	if len(members) == 0 {
		return
	}

	s.Lock()
	defer s.Unlock()

	err := s.fetchConfig(session)
	if err != nil {
		log.Errorf("Error fetching config while adding members: '%s'", err.Error())
		return
	}

	for _, mongod := range members {
		member := rsConfig.NewMember(mongod.Name())
		member.Tags = &rsConfig.ReplsetTags{
			frameworkTagName: mongod.FrameworkName,
		}
		if mongod.IsBackupNode() {
			member.Hidden = true
			member.Priority = 0
			member.Tags = &rsConfig.ReplsetTags{
				"backup":         "true",
				frameworkTagName: mongod.FrameworkName,
			}
			member.Votes = 0
		}
		configManager.AddMember(member)
		s.doUpdate = true
	}

	s.updateConfig()
}

// RemoveConfigMembers removes members from the MongoDB Replica Set config
func (s *State) RemoveConfigMembers(session *mgo.Session, configManager rsConfig.Manager, members []*rsConfig.Member) {
	if len(members) == 0 {
		return
	}

	s.Lock()
	defer s.Unlock()

	err := s.fetchConfig(session)
	if err != nil {
		log.Errorf("Error fetching config while removing members: '%s'", err.Error())
		return
	}

	for _, member := range members {
		configManager.RemoveMember(member)
		s.doUpdate = true
	}

	s.updateConfig()
}
