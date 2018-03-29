package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/percona/dcos-mongo-tools/common"
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
		DB: common.NewDBConfig(
			common.EnvMongoDBClusterMonitorUser,
			common.EnvMongoDBClusterMonitorPassword,
		),
	}
	command := kingpin.Parse()

	if config.Tool.PrintVersion {
		config.Tool.PrintVersionAndExit()
	}

	common.SetupLogger(config.Tool)

	session, err := common.GetSession(config.DB)
	if err != nil {
		log.Fatalf("Error connecting to mongodb: %s", err)
		return
	}
	defer session.Close()

	switch command {
	case health.FullCommand():
		log.Debug("Running health check")
		exitCode, err := healthcheck.HealthCheck(session)
		if err != nil {
			log.Debug(err.Error())
			session.Close()
			os.Exit(int(exitCode))
		}
		log.Debug("Member passed health check")
	case readiness.FullCommand():
		log.Debug("Running readiness check")
		exitCode, err := healthcheck.ReadinessCheck(session)
		if err != nil {
			log.Debug(err.Error())
			session.Close()
			os.Exit(int(exitCode))
		}
		log.Debug("Member passed readiness check")
	}
}
