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
	"time"

	wdConfig "github.com/percona/dcos-mongo-tools/watchdog/config"
	"github.com/stretchr/testify/assert"
)

var (
	testWatchdogConfig = &wdConfig.Config{
		Username:       "admin",
		Password:       "123456",
		ReplsetTimeout: time.Second,
	}
	testReplsetName   = "rs"
	testReplset       = &Replset{}
	testReplsetMongod = &Mongod{
		Host:          "test123",
		Port:          12345,
		Replset:       testReplsetName,
		PodName:       "mongod",
		FrameworkName: "test",
	}
)

func TestNewReplset(t *gotesting.T) {
	testReplset = New(testWatchdogConfig, testReplsetName)
	assert.Equal(t, testReplsetName, testReplset.Name, "replset.Name is incorrect")
	assert.Len(t, testReplset.Members, 0, "replset.Members is not empty")
	assert.Zero(t, testReplset.LastUpdated, "replset.LastUpdated is not empty/zero")
}

func TestReplsetGetMemberFalse(t *gotesting.T) {
	assert.Nil(t, testReplset.GetMember(testReplsetMongod.Name()), "replset.GetMember() returned unexpected result")
}

func TestReplsetHasMemberFalse(t *gotesting.T) {
	assert.False(t, testReplset.HasMember(testReplsetMongod.Name()), "replset.HasMember() returned unexpected result")
}

func TestReplsetUpdateMember(t *gotesting.T) {
	testReplset.UpdateMember(testReplsetMongod)
	assert.Len(t, testReplset.Members, 1, "replset.Members length is not 1")
}

func TestReplsetGetMember(t *gotesting.T) {
	member := testReplset.GetMember(testReplsetMongod.Name())
	assert.Equal(t, testReplsetMongod, member, "replset.GetMember() returned unexpected result")
}

func TestReplsetGetMembers(t *gotesting.T) {
	assert.Len(t, testReplset.GetMembers(), 1, "replset.GetMembers() returned unexpected result")
}

func TestReplsetHasMember(t *gotesting.T) {
	assert.True(t, testReplset.HasMember(testReplsetMongod.Name()), "replset.HasMember() returned unexpected result")
}

func TestGetReplsetDialInfo(t *gotesting.T) {
	dialInfo := testReplset.GetReplsetDialInfo()
	assert.NotNil(t, dialInfo, "replset.GetReplsetDialInfo() returned nil *mgo.DialInfo")
	assert.Lenf(t, dialInfo.Addrs, len(testReplset.GetMembers()), "*mgo.DialInfo 'Addrs' must have the length %d", len(testReplset.GetMembers()))
	assert.Equal(t, testWatchdogConfig.Username, dialInfo.Username, "*mgo.DialInfo 'Username' is incorrect")
	assert.Equal(t, testWatchdogConfig.Password, dialInfo.Password, "*mgo.DialInfo 'Password' is incorrect")
	assert.Equal(t, testReplset.Name, dialInfo.ReplicaSetName, "*mgo.DialInfo 'ReplicaSetName' is incorrect")
	assert.Equal(t, testWatchdogConfig.ReplsetTimeout, dialInfo.Timeout, "*mgo.DialInfo 'Timeout' is incorrect")
	assert.False(t, dialInfo.Direct, "*mgo.DialInfo 'Direct' must be false")
	assert.True(t, dialInfo.FailFast, "*mgo.DialInfo 'FailFast' must be true")
}

func TestReplsetRemoveMember(t *gotesting.T) {
	testReplset.RemoveMember(testReplsetMongod)
	assert.False(t, testReplset.HasMember(testReplsetMongod.Name()), "replset.HasMember() returned unexpected result after replset.RemoveMember()")
	assert.Len(t, testReplset.GetMembers(), 0, "replset.GetMembers() returned unexpected result after replset.RemoveMember()")
}
