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
	"errors"
	"fmt"

	"github.com/timvaillancourt/go-mongodb-replset/status"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// OkMemberStates is a slice of acceptable replication member states
var OkMemberStates = []status.MemberState{
	status.MemberStatePrimary,
	status.MemberStateSecondary,
	status.MemberStateRecovering,
	status.MemberStateArbiter,
	status.MemberStateStartup2,
	status.MemberStateRollback,
}

// getSelfMemberState returns the replication state of the local MongoDB member
func getSelfMemberState(rsStatus *status.Status) *status.MemberState {
	member := rsStatus.GetSelf()
	if member == nil || member.Health != status.MemberHealthUp {
		return nil
	}
	return &member.State
}

// isStateOk checks if a replication member state matches one of the acceptable member states in 'OkMemberStates'
func isStateOk(memberState *status.MemberState, okMemberStates []status.MemberState) bool {
	for _, state := range okMemberStates {
		if *memberState == state {
			return true
		}
	}
	return false
}

// HealthCheck checks the replication member state of the local MongoDB member
func HealthCheck(session *mgo.Session, okMemberStates []status.MemberState) (State, *status.MemberState, error) {
	if err := checkServerStatus(session); err != nil {
		return StateFailed, nil, err
	}

	rsStatus, err := status.New(session)
	if err != nil {
		return StateFailed, nil, fmt.Errorf("error getting replica set status: %s", err)
	}

	state := getSelfMemberState(rsStatus)
	if state == nil {
		return StateFailed, state, fmt.Errorf("found no member state for self in replica set status")
	}
	if isStateOk(state, okMemberStates) {
		return StateOk, state, nil
	}

	return StateFailed, state, fmt.Errorf("member has unhealthy replication state: %s", state)
}

func checkServerStatus(session *mgo.Session) error {
	status := &ServerStatus{}
	err := session.DB("admin").Run(bson.D{{Name: "serverStatus", Value: 1}}, status)
	if err != nil {
		return err
	}
	if status.Ok == 0 {
		return errors.New(status.Errmsg)
	}
	return nil
}

type ServerStatus struct {
	Errmsg string `bson:"errmsg,omitempty" json:"errmsg,omitempty"`
	Ok     int    `bson:"ok" json:"ok"`
}
