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
	"time"

	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	"github.com/percona/mongodb-orchestration-tools/pkg"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod/mocks"
	"github.com/stretchr/testify/assert"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	rsStatus "github.com/timvaillancourt/go-mongodb-replset/status"
)

var testMemberRemoved *rsConfig.Member
var stdoutBuffer bytes.Buffer

func getRsConfig() *rsConfig.Config {
	err := testState.Fetch(testDBSession, testRsConfigManager)
	if err != nil {
		return nil
	}
	return testState.GetConfig()
}

func waitForRsMemberState(t *testing.T, mongod *Mongod, rsState rsStatus.MemberState, timeout time.Duration) {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-time.After(timeout):
			ticker.Stop()
			assert.FailNowf(t, "timeout waiting for member to change to replset state: %s", rsState.String())
			return
		case <-ticker.C:
			err := testState.Fetch(testDBSession, testRsConfigManager)
			if err != nil {
				continue
			}
			status := testState.GetStatus()
			for _, member := range status.GetMembersByState(rsState, 0) {
				if member.Name == mongod.Name() {
					ticker.Stop()
					return
				}
			}
		}
	}
}

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

func TestWatchdogReplsetStateAddRemoveConfigMembers(t *testing.T) {
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
	config := getRsConfig()
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
		config = getRsConfig()
		addedMember := config.GetMember(addMongod.Name())
		assert.NotNil(t, addedMember)
		assert.Truef(t, addedMember.Hidden, "backup node must have Hidden set to true")
		assert.Falsef(t, addedMember.ArbiterOnly, "backup node must have ArbiterOnly set to false")
		assert.Equalf(t, addedMember.Votes, 0, "backup node must have zero Votes")

		// wait for backup node to reach SECONDARY state
		waitForRsMemberState(t, addMongod, rsStatus.MemberStateSecondary, 15*time.Second)

		// remove backup node
		assert.NoError(t, testState.RemoveConfigMembers(testDBSession, testRsConfigManager, []*rsConfig.Member{{Host: addMongod.Name()}}))
	})

	// Comment out arbiter test due to mongod crash (3.6+4.0) when transitioning from arbiter->secondary:
	//	2019-01-08T20:03:38.028+0000 F -        [rsSync] Fatal assertion 34361 OplogOutOfOrder: Attempted to apply an oplog entry ({ ts: Timestamp(1546977817, 2), t: 1 }) which is not greater than our last applied OpTime ({ ts: Timestamp(1546977817, 5), t: 1 }). at src/mongo/db/repl/sync_tail.cpp 915
	//	2019-01-08T20:03:38.028+0000 F -        [rsSync]
	//
	//	***aborting after fassert() failure
	//
	//
	//	2019-01-08T20:03:38.036+0000 F -        [rsSync] Got signal: 6 (Aborted).
	//
	//	 0x563cc68566e1 0x563cc68558f9 0x563cc6855ddd 0x7f8386bbf890 0x7f838683a067 0x7f838683b448 0x563cc4ea5d22 0x563cc54072a3 0x563cc5407419 0x563cc53873b1 0x563cc6b6e430 0x7f8386bb8064 0x7f83868ed62d
	//	----- BEGIN BACKTRACE -----
	//	{"backtrace":[{"b":"563CC4467000","o":"23EF6E1","s":"_ZN5mongo15printStackTraceERSo"},{"b":"563CC4467000","o":"23EE8F9"},{"b":"563CC4467000","o":"23EEDDD"},{"b":"7F8386BB0000","o":"F890"},{"b":"7F8386805000","o":"35067","s":"gsignal"},{"b":"7F8386805000","o":"36448","s":"abort"},{"b":"563CC4467000","o":"A3ED22","s":"_ZN5mongo42fassertFailedWithStatusNoTraceWithLocationEiRKNS_6StatusEPKcj"},{"b":"563CC4467000","o":"FA02A3","s":"_ZN5mongo4repl8SyncTail17_oplogApplicationEPNS0_22ReplicationCoordinatorEPNS1_14OpQueueBatcherE"},{"b":"563CC4467000","o":"FA0419","s":"_ZN5mongo4repl8SyncTail16oplogApplicationEPNS0_22ReplicationCoordinatorE"},{"b":"563CC4467000","o":"F203B1","s":"_ZN5mongo4repl10RSDataSync4_runEv"},{"b":"563CC4467000","o":"2707430"},{"b":"7F8386BB0000","o":"8064"},{"b":"7F8386805000","o":"E862D","s":"clone"}],"processInfo":{ "mongodbVersion" : "3.6.8-2.0", "gitVersion" : "1f363207aaa5cb6efe54c8a77152cfd8aba75442", "compiledModules" : [], "uname" : { "sysname" : "Linux", "release" : "3.10.0-957.1.3.el7.x86_64", "version" : "#1 SMP Thu Nov 29 14:49:43 UTC 2018", "machine" : "x86_64" }, "somap" : [ { "b" : "563CC4467000", "elfType" : 3, "buildId" : "AC92670D34F2EA2D353D850AA23195CADDCB0ABC" }, { "b" : "7FFD710B0000", "path" : "linux-vdso.so.1", "elfType" : 3, "buildId" : "DF8F6BF69E976BF1266E476EA2E37CEE06F10C1D" }, { "b" : "7F8388391000", "path" : "/lib/x86_64-linux-gnu/libz.so.1", "elfType" : 3, "buildId" : "ADCC4A5E27D5DE8F0BC3C6021B50BA2C35EC9A8E" }, { "b" : "7F8388181000", "path" : "/lib/x86_64-linux-gnu/libbz2.so.1.0", "elfType" : 3, "buildId" : "33F03A49E909FFDAD6BC7EBA05F82B03617590E5" }, { "b" : "7F8387F65000", "path" : "/usr/lib/x86_64-linux-gnu/libsasl2.so.2", "elfType" : 3, "buildId" : "D570D2B1AB4231175CC17AF644A037C238BC1CDF" }, { "b" : "7F8387D4E000", "path" : "/lib/x86_64-linux-gnu/libresolv.so.2", "elfType" : 3, "buildId" : "C0E9A6CE03F960E690EA8F72575FFA29570E4A0B" }, { "b" : "7F8387951000", "path" : "/usr/lib/x86_64-linux-gnu/libcrypto.so.1.0.0", "elfType" : 3, "buildId" : "CFDB319C26A6DB0ED14D33D44024ED461D8A5C23" }, { "b" : "7F83876F0000", "path" : "/usr/lib/x86_64-linux-gnu/libssl.so.1.0.0", "elfType" : 3, "buildId" : "90275AC4DD8167F60BC7C599E0DBD63741D8F191" }, { "b" : "7F83874EC000", "path" : "/lib/x86_64-linux-gnu/libdl.so.2", "elfType" : 3, "buildId" : "D70B531D672A34D71DB42EB32B68E63F2DCC5B6A" }, { "b" : "7F83872E4000", "path" : "/lib/x86_64-linux-gnu/librt.so.1", "elfType" : 3, "buildId" : "A63C95FB33CCA970E141D2E13774B997C1CF0565" }, { "b" : "7F8386FE3000", "path" : "/lib/x86_64-linux-gnu/libm.so.6", "elfType" : 3, "buildId" : "152C93BA3E8590F7ED0BCDDF868600D55EC4DD6F" }, { "b" : "7F8386DCD000", "path" : "/lib/x86_64-linux-gnu/libgcc_s.so.1", "elfType" : 3, "buildId" : "BAC839560495859598E8515CBAED73C7799AE1FF" }, { "b" : "7F8386BB0000", "path" : "/lib/x86_64-linux-gnu/libpthread.so.0", "elfType" : 3, "buildId" : "9DA9387A60FFC196AEDB9526275552AFEF499C44" }, { "b" : "7F8386805000", "path" : "/lib/x86_64-linux-gnu/libc.so.6", "elfType" : 3, "buildId" : "48C48BC6ABB794461B8A558DD76B29876A0551F0" }, { "b" : "7F83885AC000", "path" : "/lib64/ld-linux-x86-64.so.2", "elfType" : 3, "buildId" : "1D98D41FBB1EABA7EC05D0FD7624B85D6F51C03C" } ] }}
	//	 mongod(_ZN5mongo15printStackTraceERSo+0x41) [0x563cc68566e1]
	//	 mongod(+0x23EE8F9) [0x563cc68558f9]
	//	 mongod(+0x23EEDDD) [0x563cc6855ddd]
	//	 libpthread.so.0(+0xF890) [0x7f8386bbf890]
	//	 libc.so.6(gsignal+0x37) [0x7f838683a067]
	//	 libc.so.6(abort+0x148) [0x7f838683b448]
	//	 mongod(_ZN5mongo42fassertFailedWithStatusNoTraceWithLocationEiRKNS_6StatusEPKcj+0x0) [0x563cc4ea5d22]
	//	 mongod(_ZN5mongo4repl8SyncTail17_oplogApplicationEPNS0_22ReplicationCoordinatorEPNS1_14OpQueueBatcherE+0x11B3) [0x563cc54072a3]
	//	 mongod(_ZN5mongo4repl8SyncTail16oplogApplicationEPNS0_22ReplicationCoordinatorE+0x129) [0x563cc5407419]
	//	 mongod(_ZN5mongo4repl10RSDataSync4_runEv+0x111) [0x563cc53873b1]
	//	 mongod(+0x2707430) [0x563cc6b6e430]
	//	 libpthread.so.0(+0x8064) [0x7f8386bb8064]
	//	 libc.so.6(clone+0x6D) [0x7f83868ed62d]
	//	-----  END BACKTRACE  -----
	//

	// test add/remove of arbiter node
	//t.Run("arbiter", func(t *testing.T) {
	//	mockTask := &mocks.Task{}
	//	mockTask.On("IsTaskType", pod.TaskTypeMongodBackup).Return(false)
	//	mockTask.On("IsTaskType", pod.TaskTypeArbiter).Return(true)
	//	addMongod.Task = mockTask

	//	// add arbiter
	//	assert.NoError(t, testState.AddConfigMembers(testDBSession, testRsConfigManager, []*Mongod{addMongod}))
	//	assert.Falsef(t, testState.doUpdate, "testState.doUpdate is true after testState.AddConfigMembers()")

	//	// test arbiter member config after add
	//	config = getRsConfig()
	//	addedMember := config.GetMember(addMongod.Name())
	//	assert.NotNil(t, addedMember)
	//	assert.Truef(t, addedMember.ArbiterOnly, "arbiter node must have ArbiterOnly set to true")
	//	assert.Equalf(t, addedMember.Priority, 0, "arbiter node must have zero Votes")

	//	// wait for backup node to reach ARBITER state
	//	waitForRsMemberState(t, addMongod, rsStatus.MemberStateArbiter, 15*time.Second)

	//	// remove arbiter
	//	assert.NoError(t, testState.RemoveConfigMembers(testDBSession, testRsConfigManager, []*rsConfig.Member{{Host: addMongod.Name()}}))
	//})

	// test add/remove of plain-mongod node
	t.Run("mongod", func(t *testing.T) {
		mockTask := &mocks.Task{}
		mockTask.On("IsTaskType", pod.TaskTypeMongodBackup).Return(false)
		mockTask.On("IsTaskType", pod.TaskTypeArbiter).Return(false)
		addMongod.Task = mockTask

		assert.NoError(t, testState.AddConfigMembers(testDBSession, testRsConfigManager, []*Mongod{addMongod}))
		assert.Falsef(t, testState.doUpdate, "testState.doUpdate is true after testState.AddConfigMembers()")

		// wait for backup node to reach SECONDARY state
		waitForRsMemberState(t, addMongod, rsStatus.MemberStateSecondary, 15*time.Second)
	})

	// test config after adding plain-mongod
	config = getRsConfig()
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
