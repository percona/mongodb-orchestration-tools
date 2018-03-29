package common

import (
	"fmt"

	"github.com/percona/dcos-mongo-tools"
)

func PrintVersion(progName string) {
	fmt.Printf("%s, tools version %s\n",
		progName,
		tools.Version,
	)
}

func DoStop(stop *chan bool) bool {
	select {
	case doStop := <-*stop:
		return doStop
	default:
		return false
	}
}
