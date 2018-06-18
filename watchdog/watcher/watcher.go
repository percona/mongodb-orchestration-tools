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
	"sync"
	"time"

	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"github.com/percona/dcos-mongo-tools/watchdog/replset"
	log "github.com/sirupsen/logrus"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	"gopkg.in/mgo.v2"
)

var (
	directReadPreference               = mgo.Monotonic
	replsetReadPreference              = mgo.PrimaryPreferred
	waitForMongodAvailableRetries uint = 10
)

type Watcher struct {
	sync.Mutex
	config            *config.Config
	masterSession     *mgo.Session
	mongodAddQueue    chan []*replset.Mongod
	mongodRemoveQueue chan []*rsConfig.Member
	replset           *replset.Replset
	state             *replset.State
	configManager     rsConfig.Manager
	stop              *chan bool
}

func New(rs *replset.Replset, config *config.Config, stop *chan bool) *Watcher {
	return &Watcher{
		config:            config,
		replset:           rs,
		stop:              stop,
		state:             replset.NewState(rs.Name),
		mongodAddQueue:    make(chan []*replset.Mongod),
		mongodRemoveQueue: make(chan []*rsConfig.Member),
	}
}

func (rw *Watcher) getReplsetSession() *mgo.Session {
	rw.Lock()
	defer rw.Unlock()

	return rw.masterSession
}

func (rw *Watcher) connectReplsetSession() error {
	dbCnf := rw.replset.GetReplsetDBConfig(rw.config.SSL)
	session, err := db.WaitForSession(dbCnf, 30, rw.config.ReplsetPoll)
	if err != nil {
		return err
	}
	session.SetMode(replsetReadPreference, true)

	rw.Lock()
	defer rw.Unlock()

	if rw.masterSession != nil {
		log.Infof("Reconnecting to %s/%s", rw.replset.Name, dbCnf.DialInfo.Addrs)
		rw.masterSession.Close()
	}
	rw.masterSession = session
	rw.configManager = rsConfig.New(rw.masterSession)

	return nil
}

func (rw *Watcher) reconnectReplsetSession() {
	err := rw.connectReplsetSession()
	if err != nil {
		log.WithFields(log.Fields{
			"replset": rw.replset.Name,
			"error":   err,
		}).Error("Error reconnecting mongodb session to replset")
	}
}

func (rw *Watcher) logReplsetState() {
	status := rw.state.GetStatus()
	if status == nil {
		return
	}
	primary := status.Primary()
	member := rw.replset.GetMember(primary.Name)

	log.WithFields(log.Fields{
		"replset":    rw.replset.Name,
		"host":       primary.Name,
		"task":       member.Task.Name(),
		"task_state": member.Task.State(),
	}).Info("Replset Primary")

	for _, secondary := range status.Secondaries() {
		member = rw.replset.GetMember(secondary.Name)
		log.WithFields(log.Fields{
			"replset":    rw.replset.Name,
			"host":       secondary.Name,
			"task":       member.Task.Name(),
			"task_state": member.Task.State(),
		}).Info("Replset Secondary")
	}
}

func (rw *Watcher) getMongodsNotInReplsetConfig() []*replset.Mongod {
	notInReplset := make([]*replset.Mongod, 0)
	replsetConfig := rw.state.GetConfig()
	if rw.state != nil && replsetConfig != nil {
		for _, member := range rw.replset.GetMembers() {
			cnfMember := replsetConfig.GetMember(member.Name())
			if cnfMember == nil {
				notInReplset = append(notInReplset, member)
			}
		}
	}
	return notInReplset
}

func (rw *Watcher) getOrphanedMembersFromReplsetConfig() []*rsConfig.Member {
	orphanedMembers := make([]*rsConfig.Member, 0)
	replsetConfig := rw.state.GetConfig()
	if rw.state != nil && replsetConfig != nil {
		for _, member := range replsetConfig.Members {
			if rw.replset.HasMember(member.Host) != true {
				orphanedMembers = append(orphanedMembers, member)
			}
		}
	}
	return orphanedMembers
}

func (rw *Watcher) waitForAvailable(member replset.Member) error {
	session, err := db.WaitForSession(
		member.DBConfig(rw.config.SSL),
		waitForMongodAvailableRetries,
		rw.config.ReplsetPoll,
	)
	if err != nil {
		return err
	}
	session.Close()
	return nil
}

func (rw *Watcher) replsetConfigAdder(add <-chan []*replset.Mongod) {
	for members := range add {
		mongods := make([]*replset.Mongod, 0)
		for _, mongod := range members {
			err := rw.waitForAvailable(mongod)
			if err != nil {
				log.WithFields(log.Fields{
					"host":    mongod.Name(),
					"retries": waitForMongodAvailableRetries,
				}).Error(err)
				continue
			}
			log.WithFields(log.Fields{
				"replset": rw.replset.Name,
				"host":    mongod.Name(),
			}).Info("Mongod not present in replset config, adding it to replset")
			mongods = append(mongods, mongod)
		}
		rw.state.AddConfigMembers(rw.getReplsetSession(), rw.configManager, mongods)
		//rw.reconnectReplsetSession()
	}
}

func (rw *Watcher) replsetConfigRemover(remove <-chan []*rsConfig.Member) {
	for members := range remove {
		if rw.state == nil {
			continue
		}
		rw.state.RemoveConfigMembers(rw.getReplsetSession(), rw.configManager, members)
		//rw.reconnectReplsetSession()
	}
}

func (rw *Watcher) UpdateMongod(mongod *replset.Mongod) {
	rw.Lock()
	defer rw.Unlock()

	fields := log.Fields{
		"replset": rw.replset.Name,
		"name":    mongod.Task.Name(),
		"state":   string(mongod.Task.State()),
		"host":    mongod.Name(),
	}

	if rw.replset.HasMember(mongod.Name()) {
		if mongod.Task.IsRemovedMongod() {
			log.WithFields(fields).Info("Removing completed mongod task")
			rw.replset.RemoveMember(mongod)
		} else {
			log.WithFields(fields).Info("Updating running mongod task")
			rw.replset.UpdateMember(mongod)
		}
	} else if mongod.Task.HasState() {
		log.WithFields(fields).Info("Adding new mongod task")
		rw.replset.UpdateMember(mongod)
	}
}

func (rw *Watcher) Run() {
	log.WithFields(log.Fields{
		"replset":  rw.replset.Name,
		"interval": rw.config.ReplsetPoll,
	}).Info("Watching replset")

	go rw.replsetConfigAdder(rw.mongodAddQueue)
	go rw.replsetConfigRemover(rw.mongodRemoveQueue)

	err := rw.connectReplsetSession()
	if err != nil {
		log.WithFields(log.Fields{
			"replset": rw.replset.Name,
			"error":   err,
		}).Error("Error getting mongodb session to replset")
		return
	}

	ticker := time.NewTicker(rw.config.ReplsetPoll)
	for {
		select {
		case <-ticker.C:
			err := rw.state.Fetch(rw.getReplsetSession(), rw.configManager)
			if err != nil {
				log.Errorf("Error fetching replset state: %s", err)
				continue
			}
			//if rw.state.GetStatus() != nil {
			rw.mongodAddQueue <- rw.getMongodsNotInReplsetConfig()
			rw.mongodRemoveQueue <- rw.getOrphanedMembersFromReplsetConfig()
			rw.logReplsetState()
			//}
		case <-*rw.stop:
			log.WithFields(log.Fields{
				"replset": rw.replset.Name,
			}).Info("Stopping watcher for replset")
			ticker.Stop()
			return
		}
	}
}
