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

func HealthCheckLiveness(session *mgo.Session, startupDelaySeconds int64) (State, *status.MemberState, error) {
	isMasterResp := IsMasterResp{}
	err := session.Run(bson.D{{Name: "isMaster", Value: 1}}, &isMasterResp)
	if err != nil {
		return StateFailed, nil, fmt.Errorf("isMaster returned error %v", err)
	}
	if isMasterResp.Ok == 0 {
		return StateFailed, nil, errors.New(isMasterResp.Errmsg)
	}

	info, err := session.BuildInfo()
	if err != nil {
		return StateFailed, nil, fmt.Errorf("failed to get mongo build info: %v", err)
	}

	replSetStatusCommand := bson.D{{Name: "replSetGetStatus", Value: 1}}
	if info.Version < "4.2.1" {
		// https://docs.mongodb.com/manual/reference/command/replSetGetStatus/#syntax
		replSetStatusCommand = append(replSetStatusCommand, bson.DocElem{Name: "initialSync", Value: 1})
	}

	replSetGetStatusResp := ReplSetStatus{}
	err = session.Run(replSetStatusCommand, &replSetGetStatusResp)
	if err != nil {
		return StateFailed, nil, fmt.Errorf("replSetGetStatus returned error %v", err)
	}

	err = replSetGetStatusResp.CheckState(startupDelaySeconds)
	if err != nil {
		return StateFailed, &replSetGetStatusResp.MyState, err
	}

	return StateOk, &replSetGetStatusResp.MyState, nil
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
	Ok     int    `bson:"ok" json:"ok"`
	Errmsg string `bson:"errmsg,omitempty" json:"errmsg,omitempty"`
}

type IsMasterResp struct {
	IsMaster bool `bson:"ismaster" json:"ismaster"`

	Ok     int    `bson:"ok" json:"ok"`
	Errmsg string `bson:"errmsg,omitempty" json:"errmsg,omitempty"`
}

type ReplSetStatus struct {
	status.Status     `bson:",inline"`
	InitialSyncStatus InitialSyncStatus `bson:"initialSyncStatus" json:"initialSyncStatus"`
}

type InitialSyncStatus interface{}

func (rs ReplSetStatus) CheckState(startupDelaySeconds int64) error {
	if rs.Ok == 0 {
		return errors.New(rs.Errmsg)
	}

	uptime := rs.GetSelf().Uptime

	switch rs.MyState {
	case status.MemberStatePrimary, status.MemberStateSecondary, status.MemberStateArbiter:
		return nil
	case status.MemberStateStartup, status.MemberStateStartup2:
		if (rs.InitialSyncStatus == nil && uptime > 30) || (rs.InitialSyncStatus != nil && uptime > startupDelaySeconds) {
			return fmt.Errorf("state is %s and uptime is %d", rs.MyState, uptime)
		}
	case status.MemberStateRecovering:
		if uptime > startupDelaySeconds {
			return fmt.Errorf("state is %s and uptime is %d", rs.MyState, uptime)
		}
	case status.MemberStateUnknown, status.MemberStateDown, status.MemberStateRollback, status.MemberStateRemoved:
		return fmt.Errorf("invalid state %s", rs.MyState)
	default:
		return fmt.Errorf("state is unknown %s", rs.MyState)
	}

	return nil
}
