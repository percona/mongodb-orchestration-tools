package healthcheck

import (
	"fmt"

	"gopkg.in/mgo.v2"
)

func ReadinessCheck(session *mgo.Session) (State, error) {
	err := session.Ping()
	if err != nil {
		return StateFailed, fmt.Errorf("Failed to get successful ping: %s", err)
	}
	return StateOk, nil
}
