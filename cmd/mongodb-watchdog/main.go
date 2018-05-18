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

package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/watchdog"
	config "github.com/percona/dcos-mongo-tools/watchdog/config"
	log "github.com/sirupsen/logrus"
)

var (
	GitCommit string
	GitBranch string
)

func main() {
	app := kingpin.New("mongodb-watchdog", "A daemon for watching the DC/OS SDK API for MongoDB tasks and updating the MongoDB replica set state on changes")
	common.HandleAppVersion(app, GitCommit, GitBranch)

	cnf := &config.Config{
		API: &api.Config{},
	}
	app.Flag(
		"framework",
		"API framework name, this flag or env var "+common.EnvFrameworkName+" is required",
	).Default(common.DefaultFrameworkName).Envar(common.EnvFrameworkName).StringVar(&cnf.FrameworkName)
	app.Flag(
		"username",
		"MongoDB clusterAdmin username, this flag or env var "+common.EnvMongoDBClusterAdminUser+" is required",
	).Envar(common.EnvMongoDBClusterAdminUser).Required().StringVar(&cnf.Username)
	app.Flag(
		"password",
		"MongoDB clusterAdmin password, this flag or env var "+common.EnvMongoDBClusterAdminPassword+" is required",
	).Envar(common.EnvMongoDBClusterAdminPassword).Required().StringVar(&cnf.Password)
	app.Flag(
		"apiPoll",
		"Frequency of DC/OS SDK API polls, overridden by env var WATCHDOG_API_POLL",
	).Default(config.DefaultAPIPoll).Envar("WATCHDOG_API_POLL").DurationVar(&cnf.APIPoll)
	app.Flag(
		"apiTimeout",
		"DC/OS SDK API timeout, overridden by env var WATCHDOG_API_TIMEOUT",
	).Default(api.DefaultTimeout).Envar("WATCHDOG_API_TIMEOUT").DurationVar(&cnf.API.Timeout)
	app.Flag(
		"replsetPoll",
		"Frequency of replset state polls, overridden by env var WATCHDOG_REPLSET_POLL",
	).Default(config.DefaultReplsetPoll).Envar("WATCHDOG_REPLSET_POLL").DurationVar(&cnf.ReplsetPoll)
	app.Flag(
		"replsetTimeout",
		"MongoDB connect timeout, should be less than 'replsetPoll', overridden by env var WATCHDOG_REPLSET_TIMEOUT",
	).Default(config.DefaultReplsetTimeout).Envar("WATCHDOG_REPLSET_TIMEOUT").DurationVar(&cnf.ReplsetTimeout)
	app.Flag(
		"replsetConfUpdatePoll",
		"Frequency of replica set config state updates, overridden by env var WATCHDOG_REPLSET_CONF_UPDATE_POLL",
	).Default(config.DefaultReplsetConfUpdatePoll).Envar("WATCHDOG_REPLSET_CONF_UPDATE_POLL").DurationVar(&cnf.ReplsetConfUpdatePoll)
	app.Flag(
		"delayWatcherStart",
		"Amount of time to delay the start of replset watchers, overridden by env var WATCHDOG_DELAY_WATCHER_START",
	).Default(config.DefaultDelayWatcher).Envar("WATCHDOG_DELAY_WATCHER_START").DurationVar(&cnf.DelayWatcher)
	app.Flag(
		"apiHostPrefix",
		"DC/OS SDK API hostname prefix, used to construct the DCOS API hostname",
	).Default(api.DefaultHostPrefix).StringVar(&cnf.API.HostPrefix)
	app.Flag(
		"apiHostSuffix",
		"DC/OS SDK API hostname suffix, used to construct the DCOS API hostname",
	).Default(api.DefaultHostSuffix).StringVar(&cnf.API.HostSuffix)
	app.Flag(
		"apiSecure",
		"Use secure connections to DC/OS SDK API",
	).BoolVar(&cnf.API.Secure)

	cnf.SSL = db.NewSSLConfig(app)

	common.SetupLogger(app, common.GetLogFormatter(os.Args[0]), os.Stdout)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Cannot parse command line: %s", err)
	}

	watchdog.New(cnf, api.New(
		cnf.FrameworkName,
		cnf.API,
	)).Run()
}
