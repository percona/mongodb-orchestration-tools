package healthcheck

import (
	"github.com/mesosphere/dcos-mongo/mongodb_tools/common"
)

type Config struct {
	Tool *common.ToolConfig
	DB   *common.DBConfig
}
