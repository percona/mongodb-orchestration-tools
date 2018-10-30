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
	"testing"

	"github.com/percona/mongodb-orchestration-tools/internal/db"
	"github.com/stretchr/testify/assert"
)

func TestWatchdogNewReplset(t *testing.T) {
	testReplset = New(testWatchdogConfig, testReplsetName)
	assert.Equal(t, testReplsetName, testReplset.Name, "replset.Name is incorrect")
	assert.Len(t, testReplset.members, 0, "replset.members is not empty")
}

func TestWatchdogReplsetGetMemberFalse(t *testing.T) {
	assert.Nil(t, testReplset.GetMember(testReplsetMongod.Name()), "replset.GetMember() returned unexpected result")
}

func TestWatchdogReplsetHasMemberFalse(t *testing.T) {
	assert.False(t, testReplset.HasMember(testReplsetMongod.Name()), "replset.HasMember() returned unexpected result")
}

func TestWatchdogReplsetUpdateMember(t *testing.T) {
	assert.Len(t, testReplset.members, 0, "replset.members length is not 0")
	assert.NoError(t, testReplset.UpdateMember(testReplsetMongod))
	assert.Len(t, testReplset.members, 1, "replset.members length is not 1")

	// test that an error is returned if the MaxMembers is reached
	replset := New(testWatchdogConfig, testReplsetName)
	mongod := &Mongod{
		Host:        t.Name(),
		Port:        27017,
		Replset:     testReplsetName,
		PodName:     "mongod",
		ServiceName: "test",
	}
	for len(replset.members) < MaxMembers {
		assert.NoError(t, replset.UpdateMember(mongod))
		mongod.Port++
	}
	mongod.Port++
	assert.Error(t, replset.UpdateMember(mongod))
}

func TestWatchdogReplsetGetMember(t *testing.T) {
	member := testReplset.GetMember(testReplsetMongod.Name())
	assert.Equal(t, testReplsetMongod, member, "replset.GetMember() returned unexpected result")
}

func TestWatchdogReplsetGetMembers(t *testing.T) {
	assert.Len(t, testReplset.GetMembers(), 1, "replset.GetMembers() returned unexpected result")
}

func TestWatchdogReplsetHasMember(t *testing.T) {
	assert.True(t, testReplset.HasMember(testReplsetMongod.Name()), "replset.HasMember() returned unexpected result")
}

func TestWatchdogReplsetGetReplsetDBConfig(t *testing.T) {
	dbCnf := testReplset.GetReplsetDBConfig(&db.SSLConfig{Enabled: true})
	assert.NotNil(t, dbCnf, "replset.GetReplsetDBConfig() returned nil")
	assert.NotNil(t, dbCnf.SSL, "replset.GetReplsetDBConfig() returned nil 'SSL' config")
	assert.True(t, dbCnf.SSL.Enabled, "replset.GetReplsetDBConfig() returned 'SSL' config with false Enabled field")
	assert.NotNil(t, dbCnf.DialInfo, "replset.GetReplsetDBConfig() returned nil 'DialInfo'")
	assert.Lenf(t, dbCnf.DialInfo.Addrs, len(testReplset.GetMembers()), "*mgo.DialInfo 'Addrs' must have the length %d", len(testReplset.GetMembers()))
	assert.Equal(t, testWatchdogConfig.Username, dbCnf.DialInfo.Username, "*mgo.DialInfo 'Username' is incorrect")
	assert.Equal(t, testWatchdogConfig.Password, dbCnf.DialInfo.Password, "*mgo.DialInfo 'Password' is incorrect")
	assert.Equal(t, testReplset.Name, dbCnf.DialInfo.ReplicaSetName, "*mgo.DialInfo 'ReplicaSetName' is incorrect")
	assert.Equal(t, testWatchdogConfig.ReplsetTimeout, dbCnf.DialInfo.Timeout, "*mgo.DialInfo 'Timeout' is incorrect")
	assert.False(t, dbCnf.DialInfo.Direct, "*mgo.DialInfo 'Direct' must be false")
	assert.True(t, dbCnf.DialInfo.FailFast, "*mgo.DialInfo 'FailFast' must be true")
}

func TestWatchdogReplsetRemoveMember(t *testing.T) {
	assert.Error(t, testReplset.RemoveMember("does not exist"))
	assert.NoError(t, testReplset.RemoveMember(testReplsetMongod.Name()))
	assert.False(t, testReplset.HasMember(testReplsetMongod.Name()), "replset.HasMember() returned unexpected result after replset.RemoveMember()")
	assert.Len(t, testReplset.GetMembers(), 0, "replset.GetMembers() returned unexpected result after replset.RemoveMember()")
}
