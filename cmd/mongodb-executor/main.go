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
	"path/filepath"

	"github.com/alecthomas/kingpin"
	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/executor"
	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/mongodb"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
	log "github.com/sirupsen/logrus"
)

var (
	mongod = kingpin.Command("mongod", "run a mongod instance")
	mongos = kingpin.Command("mongos", "run a mongos instance")
)

func handleMongoDB(cnf *executor.Config) {
	kingpin.Flag(
		"mongodb.configDir",
		"path to mongodb instance config file, defaults to $"+common.EnvMesosSandbox+" if available, otherwise "+mongodb.DefaultConfigDirFallback,
	).Default(mongodb.DefaultConfigDirFallback).Envar(common.EnvMesosSandbox).StringVar(&cnf.MongoDB.ConfigDir)
	kingpin.Flag(
		"mongodb.binDir",
		"path to mongodb binary directory",
	).Default(mongodb.DefaultBinDir).StringVar(&cnf.MongoDB.BinDir)
	kingpin.Flag(
		"mongodb.tmpDir",
		"path to mongodb temporary directory, defaults to $"+common.EnvMesosSandbox+"/tmp if available, otherwise "+mongodb.DefaultTmpDirFallback,
	).Default(executor.MesosSandboxPathOrFallback(
		"tmp",
		mongodb.DefaultTmpDirFallback,
	)).StringVar(&cnf.MongoDB.TmpDir)
	kingpin.Flag(
		"mongodb.user",
		"user to run mongodb instance as",
	).Default(mongodb.DefaultUser).StringVar(&cnf.MongoDB.User)
	kingpin.Flag(
		"mongodb.group",
		"group to run mongodb instance as",
	).Default(mongodb.DefaultGroup).StringVar(&cnf.MongoDB.Group)
}

func handleMetrics(cnf *executor.Config) {
	kingpin.Flag(
		"metrics.enable",
		"Enable DC/OS Metrics monitoring for MongoDB, defaults to "+common.EnvMetricsEnabled+" env var",
	).Envar(common.EnvMetricsEnabled).BoolVar(&cnf.Metrics.Enabled)
	kingpin.Flag(
		"metrics.user",
		"The user to run the mgo-statsd process as",
	).Default(metrics.DefaultUser).StringVar(&cnf.Metrics.User)
	kingpin.Flag(
		"metrics.group",
		"The group to run the mgo-statsd process as",
	).Default(metrics.DefaultGroup).StringVar(&cnf.Metrics.Group)
	kingpin.Flag(
		"metrics.intervalSecs",
		"The frequency (in seconds) to send metrics to DC/OS Metrics service, defaults to "+common.EnvMetricsIntervalSecs+" env var",
	).Default(metrics.DefaultIntervalSecs).Envar(common.EnvMetricsIntervalSecs).UintVar(&cnf.Metrics.IntervalSecs)
	kingpin.Flag(
		"metrics.mgoStatsdBin",
		"Path to the mgo-statsd binary, defaults to $MESOS_SANDBOX/mgo-statsd, otherwise $GOPATH/bin/mgo-statsd",
	).Default(executor.MesosSandboxPathOrFallback(
		"mgo-statsd",
		filepath.Join(os.Getenv("GOPATH"), "bin", "mgo-statsd"),
	)).StringVar(&cnf.Metrics.MgoStatsdBin)
	kingpin.Flag(
		"metrics.mgoStatsdConfigFile",
		"Path to the mgo-statsd config file, defaults to $MESOS_SANDBOX/mgo-statsd.ini, otherwise ./mgo-statsd.ini",
	).Default(executor.MesosSandboxPathOrFallback(
		"mgo-statsd.ini",
		"mgo-statsd.ini",
	)).StringVar(&cnf.Metrics.MgoStatsdConfigFile)
}

func handlePmm(cnf *executor.Config) {
	kingpin.Flag(
		"pmm.configDir",
		"Directory containing the PMM client config file (pmm.yml), defaults to "+common.EnvMesosSandbox+" env var",
	).Envar(common.EnvMesosSandbox).StringVar(&cnf.PMM.ConfigDir)
	kingpin.Flag(
		"pmm.enable",
		"Enable Percona PMM monitoring for OS and MongoDB, defaults to "+common.EnvPMMEnabled+" env var",
	).Envar(common.EnvPMMEnabled).BoolVar(&cnf.PMM.Enabled)
	kingpin.Flag(
		"pmm.enableQueryAnalytics",
		"Enable Percona PMM query analytics (QAN) client/agent, defaults to "+common.EnvPMMEnableQueryAnalytics+" env var",
	).Envar(common.EnvPMMEnableQueryAnalytics).BoolVar(&cnf.PMM.EnableQueryAnalytics)
	kingpin.Flag(
		"pmm.serverAddress",
		"Percona PMM server address, defaults to "+common.EnvPMMServerAddress+" env var",
	).Envar(common.EnvPMMServerAddress).StringVar(&cnf.PMM.ServerAddress)
	kingpin.Flag(
		"pmm.clientName",
		"Percona PMM client address, defaults to "+common.EnvTaskName+" env var",
	).Envar(common.EnvTaskName).StringVar(&cnf.PMM.ClientName)
	kingpin.Flag(
		"pmm.serverSSL",
		"Enable SSL communication between Percona PMM client and server, defaults to "+common.EnvPMMServerSSL+" env var",
	).Envar(common.EnvPMMServerSSL).BoolVar(&cnf.PMM.ServerSSL)
	kingpin.Flag(
		"pmm.serverInsecureSSL",
		"Enable insecure SSL communication between Percona PMM client and server, defaults to "+common.EnvPMMServerInsecureSSL+" env var",
	).Envar(common.EnvPMMServerInsecureSSL).BoolVar(&cnf.PMM.ServerInsecureSSL)
	kingpin.Flag(
		"pmm.linuxMetricsExporterPort",
		"Port number for bind Percona PMM Linux Metrics exporter to, defaults to "+common.EnvPMMLinuxMetricsExporterPort+" env var",
	).Envar(common.EnvPMMLinuxMetricsExporterPort).UintVar(&cnf.PMM.LinuxMetricsExporterPort)
	kingpin.Flag(
		"pmm.mongodbMetricsExporterPort",
		"Port number for bind Percona PMM MongoDB Metrics exporter to, defaults to "+common.EnvPMMMongoDBMetricsExporterPort+" env var",
	).Envar(common.EnvPMMMongoDBMetricsExporterPort).UintVar(&cnf.PMM.MongoDBMetricsExporterPort)
	kingpin.Flag(
		"pmm.mongodb.clusterName",
		"Percona PMM client mongodb cluster name, defaults to "+common.EnvFrameworkName+" env var",
	).Envar(common.EnvFrameworkName).StringVar(&cnf.PMM.MongoDB.ClusterName)
}

func main() {
	dbConfig := db.NewConfig(
		common.EnvMongoDBClusterMonitorUser,
		common.EnvMongoDBClusterMonitorPassword,
	)
	cnf := &executor.Config{
		DB:      dbConfig,
		MongoDB: &mongodb.Config{},
		Metrics: &metrics.Config{
			DB: dbConfig,
		},
		PMM: &pmm.Config{
			DB:      dbConfig,
			MongoDB: &pmm.ConfigMongoDB{},
		},
		Tool: common.NewToolConfig(os.Args[0]),
	}

	kingpin.Flag(
		"framework",
		"dcos framework name, overridden by env var "+common.EnvFrameworkName,
	).Default(common.DefaultFrameworkName).Envar(common.EnvFrameworkName).StringVar(&cnf.FrameworkName)
	kingpin.Flag(
		"connectRetrySleep",
		"duration to wait between retries of the connection/ping to mongodb",
	).Default(executor.DefaultConnectRetrySleep).DurationVar(&cnf.ConnectRetrySleep)
	kingpin.Flag(
		"delayBackgroundJobs",
		"Amount of time to delay running of executor background jobs",
	).Default(executor.DefaultDelayBackgroundJob).DurationVar(&cnf.DelayBackgroundJob)

	handleMongoDB(cnf)
	handleMetrics(cnf)
	handlePmm(cnf)

	cnf.NodeType = kingpin.Parse()
	common.SetupLogger(cnf.Tool)
	e := executor.New(cnf)

	if cnf.Tool.PrintVersion {
		cnf.Tool.PrintVersionAndExit()
	}

	switch cnf.NodeType {
	case executor.NodeTypeMongod:
		mongod := mongodb.NewMongod(cnf.MongoDB, cnf.NodeType)
		err := e.Run(mongod)
		if err != nil {
			log.Errorf("Failed to start mongod: %s", err)
			return
		}
	case executor.NodeTypeMongos:
		log.Error("mongos nodes are not supported yet!")
		return
	default:
		log.Error("did not start anything, this is unexpected")
		return
	}
}
