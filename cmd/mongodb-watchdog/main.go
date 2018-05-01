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
)

func main() {
	cnf := &config.Config{
		API:  &api.Config{},
		Tool: common.NewToolConfig(os.Args[0]),
	}
	kingpin.Flag(
		"framework",
		"API framework name, this flag or env var "+common.EnvFrameworkName+" is required",
	).Default(common.DefaultFrameworkName).Envar(common.EnvFrameworkName).StringVar(&cnf.FrameworkName)
	kingpin.Flag(
		"username",
		"MongoDB clusterAdmin username, this flag or env var "+common.EnvMongoDBClusterAdminUser+" is required",
	).Envar(common.EnvMongoDBClusterAdminUser).Required().StringVar(&cnf.Username)
	kingpin.Flag(
		"password",
		"MongoDB clusterAdmin password, this flag or env var "+common.EnvMongoDBClusterAdminPassword+" is required",
	).Envar(common.EnvMongoDBClusterAdminPassword).Required().StringVar(&cnf.Password)
	kingpin.Flag(
		"apiPoll",
		"Frequency of DC/OS SDK API polls, overridden by env var WATCHDOG_API_POLL",
	).Default(config.DefaultAPIPoll).Envar("WATCHDOG_API_POLL").DurationVar(&cnf.APIPoll)
	kingpin.Flag(
		"apiTimeout",
		"DC/OS SDK API timeout, overridden by env var WATCHDOG_API_TIMEOUT",
	).Default(api.DefaultTimeout).Envar("WATCHDOG_API_TIMEOUT").DurationVar(&cnf.API.Timeout)
	kingpin.Flag(
		"replsetPoll",
		"Frequency of replset state polls, overridden by env var WATCHDOG_REPLSET_POLL",
	).Default(config.DefaultReplsetPoll).Envar("WATCHDOG_REPLSET_POLL").DurationVar(&cnf.ReplsetPoll)
	kingpin.Flag(
		"replsetTimeout",
		"MongoDB connect timeout, should be less than 'replsetPoll', overridden by env var WATCHDOG_REPLSET_TIMEOUT",
	).Default(config.DefaultReplsetTimeout).Envar("WATCHDOG_REPLSET_TIMEOUT").DurationVar(&cnf.ReplsetTimeout)
	kingpin.Flag(
		"replsetConfUpdatePoll",
		"Frequency of replica set config state updates, overridden by env var WATCHDOG_REPLSET_CONF_UPDATE_POLL",
	).Default(config.DefaultReplsetConfUpdatePoll).Envar("WATCHDOG_REPLSET_CONF_UPDATE_POLL").DurationVar(&cnf.ReplsetConfUpdatePoll)
	kingpin.Flag(
		"delayWatcherStart",
		"Amount of time to delay the start of replset watchers, overridden by env var WATCHDOG_DELAY_WATCHER_START",
	).Default(config.DefaultDelayWatcher).Envar("WATCHDOG_DELAY_WATCHER_START").DurationVar(&cnf.DelayWatcher)
	kingpin.Flag(
		"apiHostPrefix",
		"DC/OS SDK API hostname prefix, used to construct the DCOS API hostname",
	).Default(api.DefaultHostPrefix).StringVar(&cnf.API.HostPrefix)
	kingpin.Flag(
		"apiHostSuffix",
		"DC/OS SDK API hostname suffix, used to construct the DCOS API hostname",
	).Default(api.DefaultHostSuffix).StringVar(&cnf.API.HostSuffix)

	cnf.SSL = db.NewSSLConfig()
	kingpin.Parse()

	if cnf.Tool.PrintVersion {
		cnf.Tool.PrintVersionAndExit()
	}

	common.SetupLogger(cnf.Tool)

	watchdog.New(cnf).Run()
}
