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

	"github.com/percona/dcos-mongo-tools/common"
	testing "github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
)

var testMemberRemoved *rsConfig.Member

func TestWatchdogReplsetNewState(t *gotesting.T) {
	testing.DoSkipTest(t)

	testState = NewState(testing.MongodbReplsetName)
	assert.Equal(t, testState.Replset, testing.MongodbReplsetName, "replset.NewState() returned State struct with incorrect 'Replset' name")
	assert.False(t, testState.doUpdate, "replset.NewState() returned State struct with 'doUpdate' set to true")
}

func TestWatchdogReplsetStateFetchConfig(t *gotesting.T) {
	testing.DoSkipTest(t)

	err := testState.fetchConfig(testRsConfigManager)
	assert.NoError(t, err, ".fetchConfig() should not return an error")

	config := testState.GetConfig()
	assert.NotNil(t, config, ".GetConfig() should not return nil")
	assert.Equal(t, config.Name, testing.MongodbReplsetName, "testState.Config.Name is incorrect")
	assert.NotZero(t, config.Members, "testState.Config.Members has no members")
}

func TestWatchdogReplsetStateFetchStatus(t *gotesting.T) {
	testing.DoSkipTest(t)

	err := testState.fetchStatus(testDBSession)
	assert.NoError(t, err, ".fetchStatus() should not return an error")

	status := testState.GetStatus()
	assert.NotNil(t, status, ".GetStatus() should not return nil")
	assert.Equal(t, status.Set, testing.MongodbReplsetName, "testState.Status.Set is incorrect")
	assert.NotZero(t, status.Members, "testState.Status.Members has no members")
}

func TestWatchdogReplsetStateFetch(t *gotesting.T) {
	testing.DoSkipTest(t)

	assert.NoError(t, testState.Fetch(testDBSession, testRsConfigManager), "testState.Fetch() failed with error")
}

func TestWatchdogReplsetStateRemoveConfigMembers(t *gotesting.T) {
	testing.DoSkipTest(t)

	config := testState.GetConfig()
	assert.NotNil(t, config, ".GetConfig() should not return nil")

	memberCount := len(config.Members)
	testMemberRemoved = config.Members[len(config.Members)-1]
	testState.RemoveConfigMembers(testDBSession, testRsConfigManager, []*rsConfig.Member{testMemberRemoved})
	assert.False(t, testState.doUpdate, "testState.doUpdate is true after testState.RemoveConfigMembers()")
	assert.Len(t, testState.GetConfig().Members, memberCount-1, "testState.Config.Members count did not reduce")
}

func TestWatchdogReplsetStateAddConfigMembers(t *gotesting.T) {
	testing.DoSkipTest(t)

	hostPort := strings.SplitN(testMemberRemoved.Host, ":", 2)
	port, _ := strconv.Atoi(hostPort[1])
	addMongod := &Mongod{
		Host:          hostPort[0],
		Port:          port,
		Replset:       testing.MongodbReplsetName,
		FrameworkName: common.DefaultFrameworkName,
		PodName:       "mongo",
	}
	config := testState.GetConfig()
	memberCount := len(config.Members)
	testState.AddConfigMembers(testDBSession, testRsConfigManager, []*Mongod{addMongod})
	assert.Falsef(t, testState.doUpdate, "testState.doUpdate is true after testState.AddConfigMembers()")

	config = testState.GetConfig()
	assert.NotNil(t, config, ".GetConfig() should not return nil")
	assert.Len(t, config.Members, memberCount+1, "config.Members count did not increase")
	member := config.GetMember(testMemberRemoved.Host)
	assert.NotNil(t, member, "config.HasMember() returned no member")
	assert.True(t, member.Tags.HasMatch(frameworkTagName, addMongod.FrameworkName), "member has missing replica set tag")
}
