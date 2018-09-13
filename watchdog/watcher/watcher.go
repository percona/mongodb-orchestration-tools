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
	"errors"
	"sync"
	"time"

	"github.com/percona/dcos-mongo-tools/internal/db"
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
	config        *config.Config
	masterSession *mgo.Session
	dbConfig      *db.Config
	replset       *replset.Replset
	state         *replset.State
	configManager rsConfig.Manager
	quit          *chan bool
	running       bool
}

func New(rs *replset.Replset, config *config.Config, quit *chan bool) *Watcher {
	return &Watcher{
		config:  config,
		replset: rs,
		state:   replset.NewState(rs.Name),
		quit:    quit,
	}
}

func (rw *Watcher) getReplsetSession() *mgo.Session {
	if rw.masterSession == nil || rw.masterSession.Ping() != nil {
		err := rw.connectReplsetSession()
		if err != nil {
			return nil
		}
	}
	return rw.masterSession
}

func (rw *Watcher) connectReplsetSession() error {
	var session *mgo.Session
	for {
		ticker := time.NewTicker(rw.config.ReplsetPoll)
		select {
		case <-ticker.C:
			rw.dbConfig = rw.replset.GetReplsetDBConfig(rw.config.SSL)
			if len(rw.dbConfig.DialInfo.Addrs) >= 1 {
				var err error
				if session == nil {
					session, err = db.GetSession(rw.dbConfig)
				}
				if err == nil && session != nil {
					session.SetMode(replsetReadPreference, true)
					err = session.Ping()
					if err == nil {
						ticker.Stop()
						break
					}
				}

				log.WithFields(log.Fields{
					"addrs":   rw.dbConfig.DialInfo.Addrs,
					"replset": rw.replset.Name,
					"ssl":     rw.dbConfig.SSL.Enabled,
				}).Errorf("Error connecting to mongodb replset: %s", err)

				if session != nil {
					session.Close()
				}
			}
		case <-*rw.quit:
			return errors.New("received quit")
		}
		break
	}

	rw.Lock()
	defer rw.Unlock()

	if rw.masterSession != nil {
		log.WithFields(log.Fields{
			"addrs":   rw.dbConfig.DialInfo.Addrs,
			"replset": rw.replset.Name,
			"ssl":     rw.dbConfig.SSL.Enabled,
		}).Info("Reconnecting to mongodb replset")
		rw.masterSession.Close()
	}
	rw.masterSession = session

	return nil
}

func (rw *Watcher) reconnectReplsetSession() {
	err := rw.connectReplsetSession()
	if err != nil {
		log.WithFields(log.Fields{
			"addrs":   rw.dbConfig.DialInfo.Addrs,
			"replset": rw.replset.Name,
			"ssl":     rw.dbConfig.SSL.Enabled,
			"error":   err,
		}).Error("Error reconnecting mongodb replset session")
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

func (rw *Watcher) waitForMongodAvailable(member replset.Member) error {
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

func (rw *Watcher) replsetConfigAdder(add []*replset.Mongod) {
	mongods := make([]*replset.Mongod, 0)
	for _, mongod := range add {
		err := rw.waitForMongodAvailable(mongod)
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
	if len(mongods) == 0 {
		return
	}

	session := rw.getReplsetSession()
	if session != nil {
		rw.state.AddConfigMembers(session, rsConfig.New(session), mongods)
	}
	rw.reconnectReplsetSession()
}

func (rw *Watcher) replsetConfigRemover(remove []*rsConfig.Member) {
	if rw.state == nil || len(remove) == 0 {
		return
	}

	session := rw.getReplsetSession()
	if session != nil {
		rw.state.RemoveConfigMembers(session, rsConfig.New(session), remove)
	}
	rw.reconnectReplsetSession()
}

func (rw *Watcher) UpdateMongod(mongod *replset.Mongod) {
	fields := log.Fields{
		"replset": rw.replset.Name,
		"name":    mongod.Task.Name(),
		"host":    mongod.Name(),
		"state":   string(mongod.Task.State()),
	}

	if rw.replset.HasMember(mongod.Name()) {
		if mongod.Task.IsRemovedMongod() {
			log.WithFields(fields).Info("Removing completed mongod task")
			rw.replset.RemoveMember(mongod)
		} else if mongod.Task.IsRunning() {
			log.WithFields(fields).Info("Updating running mongod task")
			rw.replset.UpdateMember(mongod)
		}
	} else if mongod.Task.HasState() && mongod.Task.IsRunning() {
		log.WithFields(fields).Info("Adding new mongod task")
		rw.replset.UpdateMember(mongod)
	}
}

func (rw *Watcher) setRunning(running bool) {
	rw.Lock()
	defer rw.Unlock()
	rw.running = running
}

func (rw *Watcher) IsRunning() bool {
	rw.Lock()
	defer rw.Unlock()
	return rw.running
}

func (rw *Watcher) Run() {
	log.WithFields(log.Fields{
		"replset":  rw.replset.Name,
		"interval": rw.config.ReplsetPoll,
	}).Info("Watching replset")

	err := rw.connectReplsetSession()
	if err != nil {
		log.WithError(err).Fatal("Cannot connect to replset")
	}

	rw.setRunning(true)

	ticker := time.NewTicker(rw.config.ReplsetPoll)
	for {
		select {
		case <-ticker.C:
			session := rw.getReplsetSession()
			if session == nil {
				continue
			}
			err := rw.state.Fetch(session, rsConfig.New(session))
			if err != nil {
				log.Errorf("Error fetching replset state: %s", err)
				rw.reconnectReplsetSession()
				continue
			}
			if rw.state.GetStatus() != nil {
				rw.replsetConfigAdder(rw.getMongodsNotInReplsetConfig())
				rw.replsetConfigRemover(rw.getOrphanedMembersFromReplsetConfig())
				rw.logReplsetState()
			}
		case <-*rw.quit:
			log.WithFields(log.Fields{
				"replset": rw.replset.Name,
			}).Info("Stopping watcher for replset")

			ticker.Stop()
			rw.setRunning(false)
			return
		}
	}
}
