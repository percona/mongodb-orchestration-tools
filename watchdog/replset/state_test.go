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
	"strconv"
	"strings"
	gotesting "testing"

	testing "github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
)

func TestNewState(t *gotesting.T) {
	state := NewState(nil, testing.MongodbReplsetName)
	assert.Equal(t, state.Replset, testing.MongodbReplsetName, "replset.NewState() returned State struct with incorrect 'Replset' name")
	assert.Nil(t, state.session, "replset.NewState() returned State struct with a session other than nil")
	assert.False(t, state.doUpdate, "replset.NewState() returned State struct with 'doUpdate' set to true")
}

func TestFetchConfig(t *gotesting.T) {
	testing.DoSkipTest(t)

	state := NewState(testDBSession, testing.MongodbReplsetName)
	assert.NoError(t, state.fetchConfig(), "state.fetchConfig() failed with error")

	assert.NotNil(t, state.Config, "state.Config is nil")
	assert.Equal(t, state.Config.Name, testing.MongodbReplsetName, "state.Config.Name is incorrect")
	assert.NotZero(t, state.Config.Members, "state.Config.Members has no members")
}

func TestFetchStatus(t *gotesting.T) {
	testing.DoSkipTest(t)

	state := NewState(testDBSession, testing.MongodbReplsetName)
	assert.NoError(t, state.fetchStatus(), "state.fetchStatus() failed with error")

	assert.NotNil(t, state.Status, "state.Status is nil")
	assert.Equal(t, state.Status.Set, testing.MongodbReplsetName, "state.Status.Set is incorrect")
	assert.NotZero(t, state.Status.Members, "state.Status.Members has no members")
}

func TestFetch(t *gotesting.T) {
	testing.DoSkipTest(t)

	state := NewState(testDBSession, testing.MongodbReplsetName)
	assert.NoError(t, state.Fetch(), "state.Fetch() failed with error")
}

func TestRemoveAddConfigMembers(t *gotesting.T) {
	testing.DoSkipTest(t)

	state := NewState(testDBSession, testing.MongodbReplsetName)
	assert.NoError(t, state.Fetch(), "state.Fetch() failed with error")

	memberCount := len(state.Config.Members)
	removeMember := state.Config.Members[len(state.Config.Members)-1]
	state.RemoveConfigMembers([]*rsConfig.Member{removeMember})
	assert.False(t, state.doUpdate, "state.doUpdate is true after state.RemoveConfigMembers()")
	assert.Len(t, state.Config.Members, memberCount-1, "state.Config.Members count did not reduce")

	hostPort := strings.SplitN(removeMember.Host, ":", 2)
	port, _ := strconv.Atoi(hostPort[1])
	addMongod := &Mongod{
		Host:          hostPort[0],
		Port:          port,
		Replset:       testing.MongodbReplsetName,
		FrameworkName: "test",
		PodName:       "mongo",
	}
	state.AddConfigMembers([]*Mongod{addMongod})
	assert.Falsef(t, state.doUpdate, "state.doUpdate is true after state.AddConfigMembers()")
	assert.Len(t, state.Config.Members, memberCount, "state.Config.Members count did not increase")

	member := state.Config.GetMember(removeMember.Host)
	assert.NotNil(t, member, "state.Config.HasMember() returned no member")
	assert.True(t, member.Tags.HasMatch(frameworkTagName, "test"), "member has missing replica set tag")
}
