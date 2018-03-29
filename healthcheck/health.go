package healthcheck

import (
	"fmt"

	"github.com/timvaillancourt/go-mongodb-replset/status"
	"gopkg.in/mgo.v2"
)

var (
	okStates = []status.MemberState{
		status.MemberStatePrimary,
		status.MemberStateSecondary,
		status.MemberStateRecovering,
		status.MemberStateStartup2,
	}
)

func HealthCheck(session *mgo.Session) (ExitCode, error) {
	rs_status, err := status.New(session)
	if err != nil {
		return ExitCodeFailed, fmt.Errorf("Error getting replica set status: %s", err)
	}

	member := rs_status.GetSelf()
	if member == nil || member.Health != status.MemberHealthUp {
		return ExitCodeFailed, fmt.Errorf("Member is not healthy")
	}

	for _, state := range okStates {
		if member.State == state {
			return ExitCodeOk, nil
		}
	}

	return ExitCodeFailed, fmt.Errorf("Member has bad state: %d", member.State)
}
