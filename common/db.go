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

package common

import (
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

var (
	DefaultMongoDBHost            = "localhost"
	DefaultMongoDBPort            = "27017"
	DefaultMongoDBAuthDB          = "admin"
	DefaultMongoDBTimeout         = "5s"
	DefaultMongoDBTimeoutDuration = time.Duration(5) * time.Second

	ErrMsgAuthFailedStr string = "server returned error on SASL authentication step: Authentication failed."
)

type DBConfig struct {
	DialInfo *mgo.DialInfo
}

func NewDBConfig(envUser string, envPassword string) *DBConfig {
	db := &DBConfig{
		DialInfo: &mgo.DialInfo{},
	}
	kingpin.Flag(
		"address",
		"mongodb server address (hostname:port)",
	).Default(DefaultMongoDBHost + ":" + DefaultMongoDBPort).StringsVar(&db.DialInfo.Addrs)
	kingpin.Flag(
		"replset",
		"mongodb replica set name, overridden by env var "+EnvMongoDBReplset,
	).Envar(EnvMongoDBReplset).StringVar(&db.DialInfo.ReplicaSetName)
	kingpin.Flag(
		"timeout",
		"mongodb server timeout",
	).Default(DefaultMongoDBTimeout).DurationVar(&db.DialInfo.Timeout)
	kingpin.Flag(
		"username",
		"mongodb auth username, this flag or env var "+envUser+" is required",
	).Envar(envUser).Required().StringVar(&db.DialInfo.Username)
	kingpin.Flag(
		"password",
		"mongodb auth password, this flag or env var "+envPassword+" is required",
	).Envar(envPassword).Required().StringVar(&db.DialInfo.Password)
	kingpin.Flag(
		"authDb",
		"mongodb auth database",
	).Default(DefaultMongoDBAuthDB).StringVar(&db.DialInfo.Source)
	kingpin.Flag(
		"useDirectConnection",
		"enable direct connection",
	).Default("true").BoolVar(&db.DialInfo.Direct)
	kingpin.Flag(
		"useFailFastConnection",
		"enable fail-fast connection",
	).Default("true").BoolVar(&db.DialInfo.FailFast)
	return db
}

func (dbConfig *DBConfig) IsLocalhostSession() bool {
	if dbConfig.DialInfo != nil && dbConfig.DialInfo.Direct && len(dbConfig.DialInfo.Addrs) == 1 {
		split := strings.SplitN(dbConfig.DialInfo.Addrs[0], ":", 2)
		return split[0] == "localhost"
	}
	return false
}

func (dbConfig *DBConfig) Uri() string {
	extra := ""
	if dbConfig.DialInfo.ReplicaSetName != "" {
		extra = "?replicaSet=" + dbConfig.DialInfo.ReplicaSetName
	}
	hosts := strings.Join(dbConfig.DialInfo.Addrs, ",")
	return "mongodb://" + dbConfig.DialInfo.Username + ":" + dbConfig.DialInfo.Password + "@" + hosts + extra
}

func GetSession(dbConfig *DBConfig) (*mgo.Session, error) {
	log.WithFields(log.Fields{
		"hosts": dbConfig.DialInfo.Addrs,
	}).Debug("Connecting to mongodb")

	if dbConfig.DialInfo.Username != "" && dbConfig.DialInfo.Password != "" {
		log.WithFields(log.Fields{
			"user":   dbConfig.DialInfo.Username,
			"source": dbConfig.DialInfo.Source,
		}).Debug("Enabling authentication for session")
	}

	session, err := mgo.DialWithInfo(dbConfig.DialInfo)
	if err != nil && err.Error() == ErrMsgAuthFailedStr {
		log.Debug("Authentication failed, retrying with authentication disabled")
		dbConfig.DialInfo.Username = ""
		dbConfig.DialInfo.Password = ""
		session, err = mgo.DialWithInfo(dbConfig.DialInfo)
	}
	if err != nil {
		return nil, err
	}

	session.SetMode(mgo.Monotonic, true)
	return session, err
}

func WaitForSession(dbConfig *DBConfig, maxRetries uint, sleepDuration time.Duration) (*mgo.Session, error) {
	var err error
	var tries uint
	for tries <= maxRetries || maxRetries == 0 {
		session, err := GetSession(dbConfig)
		if err == nil && session.Ping() == nil {
			return session, err
		}
		time.Sleep(sleepDuration)
		tries += 1
	}
	return nil, err
}
