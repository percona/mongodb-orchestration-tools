package executor

import (
	"os"
	"path/filepath"
	"time"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/mongodb"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
)

const (
	NodeTypeMongod             = "mongod"
	NodeTypeMongos             = "mongos"
	DefaultDelayBackgroundJob  = "15s"
	DefaultConnectRetrySleep   = "5s"
	DefaultMetricsIntervalSecs = "10"
)

type Config struct {
	DB                 *common.DBConfig
	MongoDB            *mongodb.Config
	PMM                *pmm.Config
	Metrics            *metrics.Config
	Tool               *common.ToolConfig
	NodeType           string
	FrameworkName      string
	DelayBackgroundJob time.Duration
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
