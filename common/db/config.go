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

package db

import (
	"strings"
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/percona/dcos-mongo-tools/common"
	"gopkg.in/mgo.v2"
)

var (
	DefaultMongoDBHost            = "localhost"
	DefaultMongoDBPort            = "27017"
	DefaultMongoDBAuthDB          = "admin"
	DefaultMongoDBTimeout         = "5s"
	DefaultMongoDBTimeoutDuration = time.Duration(5) * time.Second
)

type Config struct {
	DialInfo *mgo.DialInfo
	SSL      *SSLConfig
}

func NewConfig(envUser string, envPassword string) *Config {
	db := &Config{
		DialInfo: &mgo.DialInfo{},
		SSL:      &SSLConfig{},
	}
	kingpin.Flag(
		"address",
		"mongodb server address (hostname:port)",
	).Default(DefaultMongoDBHost + ":" + DefaultMongoDBPort).StringsVar(&db.DialInfo.Addrs)
	kingpin.Flag(
		"replset",
		"mongodb replica set name, overridden by env var "+common.EnvMongoDBReplset,
	).Envar(common.EnvMongoDBReplset).StringVar(&db.DialInfo.ReplicaSetName)
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
		"ssl",
		"enable SSL secured mongodb connection, overridden by env var "+common.EnvMongoDBNetSSLEnabled,
	).Envar(common.EnvMongoDBNetSSLEnabled).BoolVar(&db.SSL.Enabled)
	kingpin.Flag(
		"sslPEMKeyFile",
		"path to client SSL Certificate file (including key, in PEM format), overriden by env var "+common.EnvMongoDBNetSSLPEMKeyFile,
	).Envar(common.EnvMongoDBNetSSLPEMKeyFile).ExistingFileVar(&db.SSL.PEMKeyFile)
	kingpin.Flag(
		"sslCAFile",
		"path to SSL Certificate Authority file (in PEM format), overridden by env var "+common.EnvMongoDBNetSSLCAFile,
	).Envar(common.EnvMongoDBNetSSLCAFile).ExistingFileVar(&db.SSL.CAFile)
	kingpin.Flag(
		"sslInsecure",
		"skip validation of the SSL certificate and hostname, overridden by env var "+common.EnvMongoDBNetSSLInsecure,
	).Envar(common.EnvMongoDBNetSSLInsecure).BoolVar(&db.SSL.Insecure)
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

func (cnf *Config) IsLocalhostSession() bool {
	if cnf.DialInfo != nil && cnf.DialInfo.Direct && len(cnf.DialInfo.Addrs) == 1 {
		split := strings.SplitN(cnf.DialInfo.Addrs[0], ":", 2)
		return split[0] == "localhost"
	}
	return false
}

func (cnf *Config) Uri() string {
	extra := ""
	if cnf.DialInfo.ReplicaSetName != "" {
		extra = "?replicaSet=" + cnf.DialInfo.ReplicaSetName
	}
	hosts := strings.Join(cnf.DialInfo.Addrs, ",")
	return "mongodb://" + cnf.DialInfo.Username + ":" + cnf.DialInfo.Password + "@" + hosts + extra
}
