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

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/common/tool"
	"github.com/percona/dcos-mongo-tools/healthcheck"
	"github.com/percona/pmgo"
	log "github.com/sirupsen/logrus"
)

var (
	GitCommit string
	GitBranch string
)

func main() {
	app, _ := tool.New("Performs DC/OS health and readiness checks for MongoDB", GitCommit, GitBranch)

	health := app.Command("health", "Run DCOS health check")
	readiness := app.Command("readiness", "Run DCOS readiness check").Default()
	cnf := db.NewConfig(
		app,
		common.EnvMongoDBClusterMonitorUser,
		common.EnvMongoDBClusterMonitorPassword,
	)

	command, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("Cannot parse command line: %s", err)
	}
	if _, err := os.Stat(cnf.DialInfo.Password); err == nil {
		log.Infof("Loading db password from %s", cnf.DialInfo.Password)
		str := common.StringFromFile(cnf.DialInfo.Password)
		if str != nil {
			cnf.DialInfo.Password = *str
		}
	}

	session, err := db.GetSession(cnf)
	if err != nil {
		log.Fatalf("Error connecting to mongodb: %s", err)
		return
	}
	defer session.Close()

	switch command {
	case health.FullCommand():
		log.Debug("Running health check")
		state, memberState, err := healthcheck.HealthCheck(session, healthcheck.OkMemberStates)
		if err != nil {
			log.Debug(err.Error())
			session.Close()
			os.Exit(state.ExitCode())
		}
		log.Debugf("Member passed health check with replication state: %s", memberState)
	case readiness.FullCommand():
		log.Debug("Running readiness check")
		state, err := healthcheck.ReadinessCheck(pmgo.NewSessionManager(session))
		if err != nil {
			log.Debug(err.Error())
			session.Close()
			os.Exit(state.ExitCode())
		}
		log.Debug("Member passed readiness check")
	}
}
