package pmm

import (
	"time"

	"github.com/alecthomas/kingpin"
	"github.com/mesosphere/dcos-mongo/mongodb_tools/common"
)

var (
	DefaultDelayStart = "15s"
)

type ConfigMongoDB struct {
	ClusterName string
}

type Config struct {
	DB                         *common.DBConfig
	ConfigDir                  string
	Enabled                    bool
	EnableQueryAnalytics       bool
	ServerAddress              string
	ClientName                 string
	ServerSSL                  bool
	ServerInsecureSSL          bool
	MongoDB                    *ConfigMongoDB
	DelayStart                 time.Duration
	LinuxMetricsExporterPort   uint
	MongoDBMetricsExporterPort uint
}

func NewConfig(dbConfig *common.DBConfig) *Config {
	cnf := &Config{
		DB:      dbConfig,
		MongoDB: &ConfigMongoDB{},
	}
	kingpin.Flag(
		"pmm.configDir",
		"Directory containing the PMM client config file (pmm.yml), defaults to "+common.EnvMesosSandbox+" env var",
	).Envar(common.EnvMesosSandbox).StringVar(&cnf.ConfigDir)
	kingpin.Flag(
		"pmm.enable",
		"Enable Percona PMM monitoring for OS and MongoDB, defaults to "+common.EnvPMMEnabled+" env var",
	).Envar(common.EnvPMMEnabled).BoolVar(&cnf.Enabled)
	kingpin.Flag(
		"pmm.enableQueryAnalytics",
		"Enable Percona PMM query analytics (QAN) client/agent, defaults to "+common.EnvPMMEnableQueryAnalytics+" env var",
	).Envar(common.EnvPMMEnableQueryAnalytics).BoolVar(&cnf.EnableQueryAnalytics)
	kingpin.Flag(
		"pmm.serverAddress",
		"Percona PMM server address, defaults to "+common.EnvPMMServerAddress+" env var",
	).Envar(common.EnvPMMServerAddress).StringVar(&cnf.ServerAddress)
	kingpin.Flag(
		"pmm.clientName",
		"Percona PMM client address, defaults to "+common.EnvTaskName+" env var",
	).Envar(common.EnvTaskName).StringVar(&cnf.ClientName)
	kingpin.Flag(
		"pmm.delayStart",
		"Amount of time to delay start/install of Percona PMM client, defaults to "+common.EnvPMMDelayStart+" env var",
	).Default(DefaultDelayStart).Envar(common.EnvPMMDelayStart).DurationVar(&cnf.DelayStart)
	kingpin.Flag(
		"pmm.serverSSL",
		"Enable SSL communication between Percona PMM client and server, defaults to "+common.EnvPMMServerSSL+" env var",
	).Envar(common.EnvPMMServerSSL).BoolVar(&cnf.ServerSSL)
	kingpin.Flag(
		"pmm.serverInsecureSSL",
		"Enable insecure SSL communication between Percona PMM client and server, defaults to "+common.EnvPMMServerInsecureSSL+" env var",
	).Envar(common.EnvPMMServerInsecureSSL).BoolVar(&cnf.ServerInsecureSSL)
	kingpin.Flag(
		"pmm.linuxMetricsExporterPort",
		"Port number for bind Percona PMM Linux Metrics exporter to, defaults to "+common.EnvPMMLinuxMetricsExporterPort+" env var",
	).Envar(common.EnvPMMLinuxMetricsExporterPort).UintVar(&cnf.LinuxMetricsExporterPort)
	kingpin.Flag(
		"pmm.mongodbMetricsExporterPort",
		"Port number for bind Percona PMM MongoDB Metrics exporter to, defaults to "+common.EnvPMMMongoDBMetricsExporterPort+" env var",
	).Envar(common.EnvPMMMongoDBMetricsExporterPort).UintVar(&cnf.MongoDBMetricsExporterPort)
	kingpin.Flag(
		"pmm.mongodb.clusterName",
		"Percona PMM client mongodb cluster name, defaults to "+common.EnvFrameworkName+" env var",
	).Envar(common.EnvFrameworkName).StringVar(&cnf.MongoDB.ClusterName)
	return cnf
}
