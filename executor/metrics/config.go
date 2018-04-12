package metrics

import (
	"github.com/percona/dcos-mongo-tools/common"
)

type Config struct {
	DB                  *common.DBConfig
	Enabled             bool
	IntervalSecs        uint
	MgoStatsdBin        string
	MgoStatsdConfigFile string
}
