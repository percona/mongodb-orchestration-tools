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

package db

import (
	"os"
	gotesting "testing"

	testing "github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	testPrimarySession  *mgo.Session = nil
	testPrimaryDbConfig              = &Config{
		DialInfo: &mgo.DialInfo{
			Addrs:   []string{testing.MongodbPrimaryHost + ":" + testing.MongodbPrimaryPort},
			Direct:  true,
			Timeout: testing.MongodbTimeout,
		},
		SSL: &SSLConfig{},
	}
)

func TestMain(m *gotesting.M) {
	exit := m.Run()
	if testPrimarySession != nil {
		testPrimarySession.Close()
	}
	os.Exit(exit)
}

func TestGetSession(t *gotesting.T) {
	testing.DoSkipTest(t)

	// no auth
	testPrimarySession, err := GetSession(testPrimaryDbConfig)
	assert.NoErrorf(t, err, ".GetSession() returned error for %s:%s: %s", testing.MongodbPrimaryHost, testing.MongodbPrimaryPort, err)
	assert.NotNil(t, testPrimarySession, ".GetSession() should not return a nil testPrimarySession")
	assert.Equal(t, mgo.Monotonic, testPrimarySession.Mode(), ".GetSession() must return a *mgo.Session with 'Monotonic' mode set")
	assert.Len(t, testPrimarySession.LiveServers(), 1, ".GetSession() must return a *mgo.Session that is a direct connection")
	testPrimarySession.Close()

	// with auth
	testPrimaryDbConfig.DialInfo.Username = testing.MongodbAdminUser
	testPrimaryDbConfig.DialInfo.Password = testing.MongodbAdminPassword
	testPrimarySession, err = GetSession(testPrimaryDbConfig)
	assert.NoErrorf(t, err, ".GetSession() returned error for %s:%s: %s", testing.MongodbPrimaryHost, testing.MongodbPrimaryPort, err)
	assert.NotNil(t, testPrimarySession, ".GetSession() should not return a nil testPrimarySession")

	// test auth was setup correctly by running a 'usersInfo' server command for self
	resp := struct {
		Users []struct {
			User string `bson:"user"`
		} `bson:"users"`
		Ok int `bson:"ok"`
	}{}
	err = testPrimarySession.Run(bson.D{{"usersInfo", testing.MongodbAdminUser}}, &resp)
	assert.NoError(t, err, "session returned by .GetSession() should succeed at running 'usersInfo'")
	assert.NotNil(t, resp, "got empty response from 'usersInfo' server command")
	assert.Equal(t, resp.Ok, 1, "got 'ok' code that is not 1")
	assert.Len(t, resp.Users, 1, "got 'users' slice that is not length 1")
	assert.Equal(t, "admin", resp.Users[0].User, "'user' field of 'usersInfo' response is not correct")
}
