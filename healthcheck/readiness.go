package healthcheck

import (
	"fmt"

	"gopkg.in/mgo.v2"
)

func ReadinessCheck(session *mgo.Session) (ExitCode, error) {
	err := session.Ping()
	if err != nil {
		return ExitCodeFailed, fmt.Errorf("Failed to get successful ping: %s", err)
	}
	return ExitCodeOk, nil
}
