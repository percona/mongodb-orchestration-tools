package executor

import (
	"os"
	"path/filepath"
	"time"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
)

const (
	NodeTypeMongod                = "mongod"
	NodeTypeMongos                = "mongos"
	DefaultBinDir                 = "/usr/bin"
	DefaultTmpDirFallback         = "/tmp"
	DefaultMongoConfigDirFallback = "/etc"
	DefaultUser                   = "mongodb"
	DefaultGroup                  = "root"
	DefaultDelayBackgroundJob     = "10"
	DefaultConnectTries           = "3"
	DefaultConnectRetrySleep      = "3s"
)

type Config struct {
	DB                 *common.DBConfig
	PMM                *pmm.Config
	Metrics            *metrics.Config
	Tool               *common.ToolConfig
	NodeType           string
	FrameworkName      string
	ConfigDir          string
	BinDir             string
	TmpDir             string
	User               string
	Group              string
	DelayBackgroundJob time.Duration
	ConnectTries       uint
	ConnectRetrySleep  time.Duration
}

func MesosSandboxPathOrFallback(path string, fallback string) string {
	mesosSandbox := os.Getenv(common.EnvMesosSandbox)
	if mesosSandbox != "" {
		if _, err := os.Stat(mesosSandbox); err == nil {
			return filepath.Join(mesosSandbox, path)
		}
	}
	return fallback
}
