package metrics

import (
	"github.com/percona/dcos-mongo-tools/common"
)

const (
	DefaultUser         = "nobody"
	DefaultGroup        = "nogroup"
	DefaultIntervalSecs = "10"
)

type Config struct {
	DB                  *common.DBConfig
	Enabled             bool
	User                string
	Group               string
	IntervalSecs        uint
	MgoStatsdBin        string
	MgoStatsdConfigFile string
}
