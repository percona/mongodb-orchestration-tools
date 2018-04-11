package pmm

import (
	"github.com/percona/dcos-mongo-tools/common"
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
	LinuxMetricsExporterPort   uint
	MongoDBMetricsExporterPort uint
}
