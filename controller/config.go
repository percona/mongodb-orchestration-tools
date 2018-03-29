package controller

import (
	"time"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/api"
)

var (
	DefaultPodName              = "mongo"
	DefaultInitDelay            = "15s"
	DefaultRetrySleep           = "3s"
	DefaultMaxConnectTries      = "30"
	DefaultInitMaxReplTries     = "60"
	DefaultInitMaxAddUsersTries = "60"
)

type ConfigReplsetInit struct {
	PrimaryAddr      string
	MongoDBPort      string
	Delay            time.Duration
	MaxConnectTries  uint
	MaxReplTries     uint
	MaxAddUsersTries uint
	RetrySleep       time.Duration
}

type ConfigUser struct {
	API             *api.Config
	Endpoint        *api.Endpoint
	EndpointName    string
	Database        string
	Username        string
	File            string
	MaxConnectTries uint
	RetrySleep      time.Duration
}

type Config struct {
	DB                *common.DBConfig
	Tool              *common.ToolConfig
	FrameworkName     string
	Replset           string
	UserAdminUser     string
	UserAdminPassword string
	ReplsetInit       *ConfigReplsetInit
	User              *ConfigUser
}
