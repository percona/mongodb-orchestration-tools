// Copyright 2018 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

func getSelfMemberState(rs_status *status.Status) *status.MemberState {
	member := rs_status.GetSelf()
	if member == nil || member.Health != status.MemberHealthUp {
		return nil
	}
	return &member.State
}

func isStateOk(memberState *status.MemberState) bool {
	for _, state := range okStates {
		if *memberState == state {
			return true
		}
	}
	return false
}

func HealthCheck(session *mgo.Session) (State, error) {
	rs_status, err := status.New(session)
	if err != nil {
		return StateFailed, fmt.Errorf("Error getting replica set status: %s", err)
	}

	state := getSelfMemberState(rs_status)
	if state == nil {
		return StateFailed, fmt.Errorf("Found no member state for self in replica set status")
	}
	if isStateOk(state) {
		return StateOk, nil
	}

	return StateFailed, fmt.Errorf("Member has bad state: %d", state)
}
