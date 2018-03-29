package executor

import (
	"os"
	"path/filepath"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
)

const (
	NodeTypeMongod                string = "mongod"
	NodeTypeMongos                string = "mongos"
	DefaultBinDir                 string = "/usr/bin"
	DefaultTmpDirFallback         string = "/tmp"
	DefaultMongoConfigDirFallback string = "/etc"
	DefaultUser                   string = "mongodb"
	DefaultGroup                  string = "root"
)

type Config struct {
	DB            *common.DBConfig
	PMM           *pmm.Config
	Tool          *common.ToolConfig
	NodeType      string
	FrameworkName string
	ConfigDir     string
	BinDir        string
	TmpDir        string
	User          string
	Group         string
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
