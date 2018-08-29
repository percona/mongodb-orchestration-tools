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
	gotesting "testing"
	"time"

	testing "github.com/percona/dcos-mongo-tools/internal/testing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

func TestCommonDBGetSession(t *gotesting.T) {
	testing.DoSkipTest(t)

	assert.False(t, testPrimaryDbConfig.SSL.Enabled, "SSL should be disabled")

	// no auth
	var err error
	testPrimarySession, err = GetSession(testPrimaryDbConfig)
	assert.NoErrorf(t, err, ".GetSession() returned error for %s:%s: %s", testing.MongodbHost, testing.MongodbPrimaryPort, err)
	assert.NotNil(t, testPrimarySession, ".GetSession() should not return a nil testPrimarySession")
	assert.Equal(t, mgo.Monotonic, testPrimarySession.Mode(), ".GetSession() must return a *mgo.Session with 'Monotonic' mode set")
	assert.Len(t, testPrimarySession.LiveServers(), 1, ".GetSession() must return a *mgo.Session that is a direct connection")
	testPrimarySession.Close()

	// with auth
	testPrimaryDbConfig.DialInfo.Username = testing.MongodbAdminUser
	testPrimaryDbConfig.DialInfo.Password = testing.MongodbAdminPassword
	testPrimarySession, err = GetSession(testPrimaryDbConfig)
	assert.NoErrorf(t, err, ".GetSession() returned error for %s:%s: %s", testing.MongodbHost, testing.MongodbPrimaryPort, err)
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

func TestCommonDBWaitForSession(t *gotesting.T) {
	testing.DoSkipTest(t)

	failConfig := &Config{
		DialInfo: &mgo.DialInfo{
			Addrs:    []string{"thisshouldfail:27017"},
			FailFast: true,
			Timeout:  time.Second,
		},
		SSL: &SSLConfig{},
	}
	session, err := WaitForSession(failConfig, 1, time.Second)
	assert.Error(t, err, ".WaitForSession() should fail due to bad dial info")
	assert.Nil(t, session, ".WaitForSession() should return a nil *mgo.Session on failure")

	session, err = WaitForSession(testPrimaryDbConfig, 3, time.Second)
	assert.NoError(t, err, ".WaitForSession() should not return an error")
	assert.NotNil(t, session, ".WaitForSession() should not return a nil session")
	defer session.Close()
	assert.NoError(t, session.Ping(), ".WaitForSession() should return a ping-able session")
}

func TestCommonDBWaitForPrimary(t *gotesting.T) {
	testing.DoSkipTest(t)

	assert.NoError(t, WaitForPrimary(testPrimarySession, 1, time.Second), ".WaitForPrimary() should return no error for primary")

	secondarySession, err := testing.GetSession(testing.MongodbSecondary1Port)
	assert.NoError(t, err, "could not get secondary-host session for testing .WaitForPrimary()")
	assert.NotNil(t, secondarySession)
	defer secondarySession.Close()
	secondarySession.SetMode(mgo.Eventual, true)

	err = WaitForPrimary(secondarySession, 1, time.Second)
	assert.Error(t, err, ".WaitForPrimary() should return an error for secondary")
	assert.Equal(t, err, ErrPrimaryTimeout, ".WaitForPrimary() should return a ErrPrimaryTimeout error on timeout")
}
