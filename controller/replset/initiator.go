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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/controller"
	"github.com/percona/dcos-mongo-tools/controller/user"
	log "github.com/sirupsen/logrus"
	rs_config "github.com/timvaillancourt/go-mongodb-replset/config"
	"gopkg.in/mgo.v2"
)

const (
	ErrMsgDNSNotReady = "No host described in new configuration 1 for replica set rs maps to this node"
)

type Initiator struct {
	config        *controller.Config
	hostname      string
	connectTries  uint
	replInitTries uint
	usersTries    uint
}

func NewInitiator(config *controller.Config) *Initiator {
	return &Initiator{
		config: config,
	}
}

func (i *Initiator) InitReplset(session *mgo.Session) error {
	rsCnfMan := rs_config.New(session)
	if rsCnfMan.IsInitiated() {
		return errors.New("Replset should not be initiated already! Exiting")
	}

	config := rs_config.NewConfig(i.config.Replset)
	member := rs_config.NewMember(i.config.ReplsetInit.PrimaryAddr)
	member.Tags = &rs_config.ReplsetTags{
		"dcosFramework": i.config.FrameworkName,
	}
	config.AddMember(member)
	rsCnfMan.Set(config)

	log.Info("Initiating replset")
	fmt.Println(config)

	for i.replInitTries <= i.config.ReplsetInit.MaxReplTries {
		err := rsCnfMan.Initiate()
		if err == nil {
			log.WithFields(log.Fields{
				"version": config.Version,
			}).Info("Initiated replset with config:")
			break
		}
		if err.Error() != ErrMsgDNSNotReady {
			log.WithFields(log.Fields{
				"replset": i.config.Replset,
				"error":   err,
			}).Error("Error initiating replset! Retrying")
		}
		time.Sleep(i.config.ReplsetInit.RetrySleep)
		i.replInitTries += 1
	}
	if i.replInitTries >= i.config.ReplsetInit.MaxReplTries {
		return errors.New("Could not init replset")
	}

	return nil
}

func (i *Initiator) InitAdminUser(session *mgo.Session) error {
	for i.usersTries <= i.config.ReplsetInit.MaxAddUsersTries {
		err := user.UpdateUser(session, user.UserAdmin, "admin")
		if err != nil {
			if err.Error() == "not master" {
				log.WithFields(log.Fields{
					"sleep":     i.config.ReplsetInit.RetrySleep,
					"retry":     i.usersTries,
					"max_retry": i.config.ReplsetInit.MaxAddUsersTries,
				}).Warn("Failed to add admin as host is not Primary yet, retrying")
				time.Sleep(i.config.ReplsetInit.RetrySleep)
				i.usersTries += 1
				continue
			}
			log.Errorf("Error adding admin user: %s", err)
		} else {
			break
		}
		time.Sleep(i.config.ReplsetInit.RetrySleep)
		i.usersTries += 1
	}
	if i.usersTries >= i.config.ReplsetInit.MaxAddUsersTries {
		return errors.New("Could not add admin user")
	}
	return nil
}

func (i *Initiator) InitUsers(session *mgo.Session) error {
	for i.usersTries <= i.config.ReplsetInit.MaxAddUsersTries {
		err := user.UpdateUsers(session, user.SystemUsers, "admin")
		if err != nil {
			log.Errorf("Error adding users: %s", err)
		} else {
			break
		}
		time.Sleep(i.config.ReplsetInit.RetrySleep)
		i.usersTries += 1
	}
	if i.usersTries >= i.config.ReplsetInit.MaxAddUsersTries {
		return errors.New("Could not add users")
	}
	return nil
}

func (i *Initiator) Run() error {
	log.WithFields(log.Fields{
		"framework": i.config.FrameworkName,
	}).Info("Mongod replset initiator started")

	log.WithFields(log.Fields{
		"sleep": i.config.ReplsetInit.Delay,
	}).Info("Waiting to start initiation")
	time.Sleep(i.config.ReplsetInit.Delay)

	split := strings.SplitN(i.config.ReplsetInit.PrimaryAddr, ":", 2)
	localhostNoAuthSession, err := db.WaitForSession(
		&db.Config{
			DialInfo: &mgo.DialInfo{
				Addrs:    []string{"localhost:" + split[1]},
				Direct:   true,
				FailFast: true,
				Timeout:  db.DefaultMongoDBTimeoutDuration,
			},
			SSL: i.config.DB.SSL,
		},
		i.config.ReplsetInit.MaxConnectTries,
		i.config.ReplsetInit.RetrySleep,
	)
	if err != nil {
		return err
	}
	defer localhostNoAuthSession.Close()

	log.WithFields(log.Fields{
		"host":    "localhost",
		"auth":    false,
		"replset": "",
	}).Info("Connected to MongoDB")

	err = i.InitReplset(localhostNoAuthSession)
	if err != nil {
		return err
	}
	err = i.InitAdminUser(localhostNoAuthSession)
	if err != nil {
		return err
	}

	log.Info("Closing localhost connection, reconnecting with a replset+auth connection")
	localhostNoAuthSession.Close()

	replsetAuthSession, err := db.WaitForSession(
		&db.Config{
			DialInfo: &mgo.DialInfo{
				Addrs:          []string{i.config.ReplsetInit.PrimaryAddr},
				Username:       i.config.UserAdminUser,
				Password:       i.config.UserAdminPassword,
				ReplicaSetName: i.config.Replset,
				Direct:         false,
				FailFast:       true,
				Timeout:        db.DefaultMongoDBTimeoutDuration,
			},
			SSL: i.config.DB.SSL,
		},
		i.config.ReplsetInit.MaxConnectTries,
		i.config.ReplsetInit.RetrySleep,
	)
	if err != nil {
		return err
	}
	defer replsetAuthSession.Close()
	log.WithFields(log.Fields{
		"host":    i.config.ReplsetInit.PrimaryAddr,
		"auth":    true,
		"replset": i.config.Replset,
	}).Info("Connected to MongoDB")

	err = i.InitUsers(replsetAuthSession)
	if err != nil {
		return err
	}

	log.Info("Mongod replset initiator complete")
	return nil
}
