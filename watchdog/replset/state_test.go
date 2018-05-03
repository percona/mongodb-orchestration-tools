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

package replset

import (
	gotesting "testing"

	testing "github.com/percona/dcos-mongo-tools/common/testing"
	//"github.com/timvaillancourt/go-mongodb-replset/status"
)

func TestNewState(t *gotesting.T) {
	state := NewState(nil, testing.MongodbReplsetName)
	if state.Replset != testing.MongodbReplsetName {
		t.Errorf("replset.NewState() returned State struct with incorrect 'Replset' name: %v", state.Replset)
	}
	if state.session != nil {
		t.Errorf("replset.NewState() returned State struct with a session other than nil: %v", state.session)
	}
	if state.doUpdate {
		t.Error("replset.NewState() returned State struct with 'doUpdate' set to true")
	}
}

func TestFetch(t *gotesting.T) {
	testing.DoSkipTest(t)

	session, err := testing.GetPrimarySession(t)
	if err != nil {
		t.Fatalf("Database connection error: %s", err)
	}
	defer session.Close()

	state := NewState(session, testing.MongodbReplsetName)
	err = state.Fetch()
	if err != nil {
		t.Fatalf("state.Fetch() failed with error: %v", err)
	}

	if state.Status == nil {
		t.Error("state.Status is nil")
	} else if state.Status.Set != testing.MongodbReplsetName {
		t.Errorf("state.Status.Set is not %s: %v", testing.MongodbReplsetName, state.Status.Set)
	} else if len(state.Status.Members) <= 1 {
		t.Error("state.Status.Members has <= 1 member")
	}

	if state.Config == nil {
		t.Error("state.Config is nil")
	} else if state.Config.Name != testing.MongodbReplsetName {
		t.Errorf("state.Config.Name is not %s: %v", testing.MongodbReplsetName, state.Config.Name)
	} else if len(state.Config.Members) <= 1 {
		t.Error("state.Config.Members has <= 1 member")
	}
}
