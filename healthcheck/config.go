package healthcheck

import (
	"github.com/percona/dcos-mongo-tools/common"
)

type Config struct {
	Tool *common.ToolConfig
	DB   *common.DBConfig
}
