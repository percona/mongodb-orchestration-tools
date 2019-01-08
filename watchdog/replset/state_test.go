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
	"bytes"
	"strconv"
	"strings"
	"testing"

	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	"github.com/percona/mongodb-orchestration-tools/pkg"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod/mocks"
	"github.com/stretchr/testify/assert"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
)

var testMemberRemoved *rsConfig.Member
var stdoutBuffer bytes.Buffer

func TestWatchdogReplsetNewState(t *testing.T) {
	testutils.DoSkipTest(t)

	testState = NewState(testutils.MongodbReplsetName)
	testState.configOut = &stdoutBuffer
	assert.Equal(t, testState.Replset, testutils.MongodbReplsetName, "replset.NewState() returned State struct with incorrect 'Replset' name")
	assert.False(t, testState.doUpdate, "replset.NewState() returned State struct with 'doUpdate' set to true")
}

func TestWatchdogReplsetStateFetchConfig(t *testing.T) {
	testutils.DoSkipTest(t)

	err := testState.fetchConfig(testRsConfigManager)
	assert.NoError(t, err, ".fetchConfig() should not return an error")

	config := testState.GetConfig()
	assert.NotNil(t, config, ".GetConfig() should not return nil")
	assert.Equal(t, config.Name, testutils.MongodbReplsetName, "testState.Config.Name is incorrect")
	assert.NotZero(t, config.Members, "testState.Config.Members has no members")
}

func TestWatchdogReplsetStateFetchStatus(t *testing.T) {
	testutils.DoSkipTest(t)

	err := testState.fetchStatus(testDBSession)
	assert.NoError(t, err, ".fetchStatus() should not return an error")

	status := testState.GetStatus()
	assert.NotNil(t, status, ".GetStatus() should not return nil")
	assert.Equal(t, status.Set, testutils.MongodbReplsetName, "testState.Status.Set is incorrect")
	assert.NotZero(t, status.Members, "testState.Status.Members has no members")
}

func TestWatchdogReplsetStateFetch(t *testing.T) {
	testutils.DoSkipTest(t)

	assert.NoError(t, testState.Fetch(testDBSession, testRsConfigManager), "testState.Fetch() failed with error")
}

func TestWatchdogReplsetStateRemoveConfigMembers(t *testing.T) {
	testutils.DoSkipTest(t)

	config := testState.GetConfig()
	assert.NotNil(t, config, ".GetConfig() should not return nil")

	memberCount := len(config.Members)
	testMemberRemoved = config.Members[len(config.Members)-1]
	testState.RemoveConfigMembers(testDBSession, testRsConfigManager, []*rsConfig.Member{testMemberRemoved})
	assert.False(t, testState.doUpdate, "testState.doUpdate is true after testState.RemoveConfigMembers()")
	assert.Len(t, testState.GetConfig().Members, memberCount-1, "testState.Config.Members count did not reduce")
}

func TestWatchdogReplsetStateAddConfigMembers(t *testing.T) {
	testutils.DoSkipTest(t)

	hostPort := strings.SplitN(testMemberRemoved.Host, ":", 2)
	port, _ := strconv.Atoi(hostPort[1])
	addMongod := &Mongod{
		Host:        hostPort[0],
		Port:        port,
		Replset:     testutils.MongodbReplsetName,
		ServiceName: pkg.DefaultServiceName,
		PodName:     t.Name(),
	}
	config := testState.GetConfig()
	memberCount := len(config.Members)

	// test add/remove of backup node
	t.Run("backup", func(t *testing.T) {
		mockTask := &mocks.Task{}
		mockTask.On("IsTaskType", pod.TaskTypeMongodBackup).Return(true)
		mockTask.On("IsTaskType", pod.TaskTypeArbiter).Return(false)
		addMongod.Task = mockTask

		// add backup node
		assert.NoError(t, testState.AddConfigMembers(testDBSession, testRsConfigManager, []*Mongod{addMongod}))
		assert.Falsef(t, testState.doUpdate, "testState.doUpdate is true after testState.AddConfigMembers()")

		// test backup node config after add
		assert.NoError(t, testState.Fetch(testDBSession, testRsConfigManager))
		config = testState.GetConfig()
		addedMember := config.GetMember(addMongod.Name())
		assert.NotNil(t, addedMember)
		assert.Truef(t, addedMember.Hidden, "backup node must have Hidden set to true")
		assert.Falsef(t, addedMember.ArbiterOnly, "backup node must have ArbiterOnly set to false")
		assert.Equalf(t, addedMember.Votes, 0, "backup node must have zero Votes")

		// remove backup node
		assert.NoError(t, testState.RemoveConfigMembers(testDBSession, testRsConfigManager, []*rsConfig.Member{{Host: addMongod.Name()}}))
	})

	// test add/remove of arbiter node
	t.Run("arbiter", func(t *testing.T) {
		mockTask := &mocks.Task{}
		mockTask.On("IsTaskType", pod.TaskTypeMongodBackup).Return(false)
		mockTask.On("IsTaskType", pod.TaskTypeArbiter).Return(true)
		addMongod.Task = mockTask

		// add arbiter
		assert.NoError(t, testState.AddConfigMembers(testDBSession, testRsConfigManager, []*Mongod{addMongod}))
		assert.Falsef(t, testState.doUpdate, "testState.doUpdate is true after testState.AddConfigMembers()")

		// test arbiter member config after add
		assert.NoError(t, testState.Fetch(testDBSession, testRsConfigManager))
		config = testState.GetConfig()
		addedMember := config.GetMember(addMongod.Name())
		assert.NotNil(t, addedMember)
		assert.Truef(t, addedMember.ArbiterOnly, "arbiter node must have ArbiterOnly set to true")
		assert.Equalf(t, addedMember.Priority, 0, "arbiter node must have zero Votes")

		// remove arbiter
		assert.NoError(t, testState.RemoveConfigMembers(testDBSession, testRsConfigManager, []*rsConfig.Member{{Host: addMongod.Name()}}))
	})

	// test add/remove of plain-mongod node
	t.Run("mongod", func(t *testing.T) {
		mockTask := &mocks.Task{}
		mockTask.On("IsTaskType", pod.TaskTypeMongodBackup).Return(false)
		mockTask.On("IsTaskType", pod.TaskTypeArbiter).Return(false)
		addMongod.Task = mockTask

		assert.NoError(t, testState.AddConfigMembers(testDBSession, testRsConfigManager, []*Mongod{addMongod}))
		assert.Falsef(t, testState.doUpdate, "testState.doUpdate is true after testState.AddConfigMembers()")
	})

	// test config after adding plain-mongod
	assert.NoError(t, testState.Fetch(testDBSession, testRsConfigManager))
	config = testState.GetConfig()
	assert.NotNil(t, config, ".GetConfig() should not return nil")
	assert.Len(t, config.Members, memberCount+1, "config.Members count did not increase")
	member := config.GetMember(testMemberRemoved.Host)
	assert.NotNil(t, member, "config.HasMember() returned no member")
	assert.True(t, member.Tags.HasMatch(serviceTagName, addMongod.ServiceName), "member has missing replica set tag")
}

func TestWatchdogReplsetStateGetMaxIDVotingMember(t *testing.T) {
	maxIDMember := &rsConfig.Member{Id: 5, Votes: 1}
	state := NewState("test")
	state.Config = &rsConfig.Config{
		Members: []*rsConfig.Member{
			{Id: 0, Votes: 1},
			{Id: 1, Votes: 1},
			maxIDMember,
			{Id: 2, Votes: 1},
		},
	}
	assert.Equal(t, maxIDMember, state.getMaxIDVotingMember(), ".getMaxIDMember() returned incorrect result or member")
}

func TestWatchdogReplsetStateGetMinIDNonVotingMember(t *testing.T) {
	minIDMember := &rsConfig.Member{Id: 1}
	s := NewState("test")
	s.Config = &rsConfig.Config{
		Members: []*rsConfig.Member{
			{Id: 0, Votes: 1},
			minIDMember,
			{Id: 2},
			{Id: 3, Votes: 1},
			{Id: 4, Votes: 1},
			{Id: 5},
		},
	}
	assert.Equal(t, minIDMember, s.getMinIDNonVotingMember(), ".getMinIDNonVotingMember() returned incorrect result or member")
}

func TestWatchdogReplsetStateResetConfigVotes(t *testing.T) {
	state := NewState("test")

	// test .restConfigVotes() will reduce voting members (9/too-many) to the max (7)
	maxMember := &rsConfig.Member{Id: 8, Votes: 1, Host: "test8"}
	state.Config = &rsConfig.Config{
		Members: []*rsConfig.Member{
			{Id: 0, Votes: 1, Host: "test0"},
			{Id: 1, Votes: 1, Host: "test1"},
			{Id: 2, Votes: 1, Host: "test2"},
			{Id: 3, Votes: 1, Host: "test3"},
			{Id: 4, Votes: 1, Host: "test4"},
			{Id: 5, Votes: 1, Host: "test5"},
			{Id: 6, Votes: 1, Host: "test6"},
			{Id: 7, Votes: 1, Host: "test7"},
			maxMember,
		},
	}
	memberCnt := len(state.Config.Members)
	state.resetConfigVotes()
	assert.Equal(t, MaxVotingMembers, state.VotingMembers())
	assert.Equal(t, 0, maxMember.Votes)
	assert.Len(t, state.Config.Members, memberCnt)

	// test .restConfigVotes() will reduce voting members when the number is even and adding votes is nott possible
	// there should be 4 voting members before and 3 after
	maxMember = &rsConfig.Member{Id: 3, Votes: 1, Host: "test3"}
	state.Config = &rsConfig.Config{
		Members: []*rsConfig.Member{
			{Id: 0, Votes: 1, Host: "test0"},
			{Id: 1, Votes: 1, Host: "test1"},
			{Id: 2, Votes: 1, Host: "test2"},
			maxMember,
		},
	}
	state.resetConfigVotes()
	assert.Equal(t, 3, state.VotingMembers(), ".resetConfigVotes() did not reduce to correct votes")
	assert.Equal(t, 0, maxMember.Votes, ".resetConfigVotes() did not remove vote from max member")

	// test .restConfigVotes() will add voting members when the number is even and adding votes to non-voting members IS possible
	// there should be 4 voting members before and 5 after
	maxMember = &rsConfig.Member{Id: 4, Votes: 0, Host: "test4"}
	state.Config = &rsConfig.Config{
		Members: []*rsConfig.Member{
			{Id: 0, Votes: 1, Host: "test0"},
			{Id: 1, Votes: 1, Host: "test1"},
			{Id: 2, Votes: 1, Host: "test2"},
			{Id: 3, Votes: 1, Host: "test3"},
			maxMember,
		},
	}
	state.resetConfigVotes()
	assert.Equal(t, 5, state.VotingMembers())
	assert.Equal(t, 1, maxMember.Votes, ".resetConfigVotes() did not increase vote of max member")
}
