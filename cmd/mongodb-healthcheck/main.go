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
	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/healthcheck"
	log "github.com/sirupsen/logrus"
)

var (
	health    = kingpin.Command("health", "Run DCOS health check")
	readiness = kingpin.Command("readiness", "Run DCOS readiness check").Default()
)

func main() {
	config := &healthcheck.Config{
		Tool: common.NewToolConfig(os.Args[0]),
		DB: db.NewConfig(
			common.EnvMongoDBClusterMonitorUser,
			common.EnvMongoDBClusterMonitorPassword,
		),
	}
	command := kingpin.Parse()

	if config.Tool.PrintVersion {
		config.Tool.PrintVersionAndExit()
	}

	common.SetupLogger(config.Tool)

	session, err := db.GetSession(config.DB)
	if err != nil {
		log.Fatalf("Error connecting to mongodb: %s", err)
		return
	}
	defer session.Close()

	switch command {
	case health.FullCommand():
		log.Debug("Running health check")
		state, err := healthcheck.HealthCheck(session)
		if err != nil {
			log.Debug(err.Error())
			session.Close()
			os.Exit(int(state))
		}
		log.Debug("Member passed health check")
	case readiness.FullCommand():
		log.Debug("Running readiness check")
		state, err := healthcheck.ReadinessCheck(session)
		if err != nil {
			log.Debug(err.Error())
			session.Close()
			os.Exit(int(state))
		}
		log.Debug("Member passed readiness check")
	}
}
