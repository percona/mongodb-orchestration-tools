package main

import (
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/mesosphere/dcos-mongo/mongodb_tools/common"
	"github.com/mesosphere/dcos-mongo/mongodb_tools/executor"
	"github.com/mesosphere/dcos-mongo/mongodb_tools/executor/pmm"
	log "github.com/sirupsen/logrus"
)

var (
	mongod = kingpin.Command("mongod", "run a mongod instance")
	mongos = kingpin.Command("mongos", "run a mongos instance")
)

func main() {
	dbConfig := common.NewDBConfig(
		common.EnvMongoDBClusterMonitorUser,
		common.EnvMongoDBClusterMonitorPassword,
	)
	cnf := &executor.Config{
		DB:   dbConfig,
		PMM:  pmm.NewConfig(dbConfig),
		Tool: common.NewToolConfig(os.Args[0]),
	}

	kingpin.Flag(
		"framework",
		"dcos framework name, overridden by env var "+common.EnvFrameworkName,
	).Default(common.DefaultFrameworkName).Envar(common.EnvFrameworkName).StringVar(&cnf.FrameworkName)
	kingpin.Flag(
		"configDir",
		"path to mongodb instance config file, defaults to $"+common.EnvMesosSandbox+" if available, otherwise "+executor.DefaultMongoConfigDirFallback,
	).Default(executor.DefaultMongoConfigDirFallback).Envar(common.EnvMesosSandbox).StringVar(&cnf.ConfigDir)
	kingpin.Flag(
		"binDir",
		"path to mongodb binary directory",
	).Default(executor.DefaultBinDir).StringVar(&cnf.BinDir)
	kingpin.Flag(
		"tmpDir",
		"path to mongodb temporary directory, defaults to $"+common.EnvMesosSandbox+"/tmp if available, otherwise "+executor.DefaultTmpDirFallback,
	).Default(executor.MesosSandboxPathOrFallback(
		"tmp",
		executor.DefaultTmpDirFallback,
	)).StringVar(&cnf.TmpDir)
	kingpin.Flag(
		"user",
		"user to run mongodb instance as",
	).Default(executor.DefaultUser).StringVar(&cnf.User)
	kingpin.Flag(
		"group",
		"group to run mongodb instance as",
	).Default(executor.DefaultGroup).StringVar(&cnf.Group)

	cnf.NodeType = kingpin.Parse()

	if cnf.Tool.PrintVersion {
		cnf.Tool.PrintVersionAndExit()
	}

	common.SetupLogger(cnf.Tool)

	err := executor.New(cnf).Start()
	if err != nil {
		log.Fatalf("Failed with error: %s", err)
		os.Exit(1)
	}
}
