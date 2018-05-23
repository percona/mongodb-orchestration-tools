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
	"time"

	log "github.com/sirupsen/logrus"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	rsStatus "github.com/timvaillancourt/go-mongodb-replset/status"
	"gopkg.in/mgo.v2"
)

const frameworkTagName = "dcosFramework"

type State struct {
	sync.Mutex
	Replset       string
	Config        *rsConfig.Config
	Status        *rsStatus.Status
	configManager *rsConfig.ConfigManager
	session       *mgo.Session
	fetcher       Fetcher
	doUpdate      bool
}

func NewState(session *mgo.Session, configManager *rsConfig.ConfigManager, fetcher Fetcher, replset string) *State {
	return &State{
		Replset:       replset,
		session:       session,
		configManager: configManager,
		fetcher:       fetcher,
	}
}

func (s *State) fetchConfig() error {
	config, err := s.fetcher.GetConfig()
	if err != nil {
		return err
	}
	s.Config = config
	return nil
}

func (s *State) fetchStatus() error {
	status, err := s.fetcher.GetStatus()
	if err != nil {
		return err
	}
	s.Status = status
	return nil
}

func (s *State) Fetch() error {
	s.Lock()
	defer s.Unlock()

	log.WithFields(log.Fields{
		"replset": s.Replset,
	}).Info("Updating replset config and status")

	err := s.fetchConfig()
	if err != nil {
		return err
	}

	return s.fetchStatus()
}

func (s *State) updateConfig() error {
	if s.doUpdate == false {
		return nil
	}

	s.configManager.IncrVersion()
	config := s.configManager.Get()
	log.WithFields(log.Fields{
		"replset":        s.Replset,
		"config_version": config.Version,
	}).Info("Writing new replset config")
	fmt.Println(config)

	err := s.configManager.Save()
	if err != nil {
		return err
	}
	s.doUpdate = false

	return nil
}

func (s *State) AddConfigMembers(mongods []*Mongod) {
	if len(mongods) == 0 {
		return
	}

	s.Lock()
	defer s.Unlock()

	err := s.fetchConfig()
	if err != nil {
		log.Errorf("Error fetching config while adding members: '%s'", err.Error())
		return
	}

	for _, mongod := range mongods {
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
		s.configManager.AddMember(member)
		s.doUpdate = true
	}
	s.updateConfig()
}

func (s *State) RemoveConfigMembers(members []*rsConfig.Member) {
	if len(members) == 0 {
		return
	}

	s.Lock()
	defer s.Unlock()

	err := s.fetchConfig()
	if err != nil {
		log.Errorf("Error fetching config while removing members: '%s'", err.Error())
		return
	}

	for _, member := range members {
		s.configManager.RemoveMember(member)
		s.doUpdate = true
	}
	s.updateConfig()
}

func (s *State) StartFetcher(stop *chan bool, interval time.Duration) {
	log.WithFields(log.Fields{
		"replset":  s.Replset,
		"interval": interval,
	}).Info("Started background replset state fetcher")

	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			s.Fetch()
		case <-*stop:
			ticker.Stop()
			break
		}
	}
}
