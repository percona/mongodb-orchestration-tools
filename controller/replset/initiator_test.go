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
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/percona/mongodb-orchestration-tools/controller"
	"github.com/percona/mongodb-orchestration-tools/controller/user"
	"github.com/percona/mongodb-orchestration-tools/internal/db"
	"github.com/percona/mongodb-orchestration-tools/internal/logger"
	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

var (
	testSession   *mgo.Session
	testInitiator *Initiator
	testConfig    = &controller.Config{
		SSL: &db.SSLConfig{},
	}
)

func TestMain(m *testing.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter(), os.Stdout)

	if testutils.Enabled() {
		var err error
		testSession, err = testutils.GetSession(testutils.MongodbPrimaryPort)
		if err != nil {
			fmt.Printf("Error getting session: %v", err)
			os.Exit(1)
		}
	}
	exit := m.Run()
	if testSession != nil {
		testSession.Close()
	}
	os.Exit(exit)
}

func TestIsError(t *testing.T) {
	assert.True(t, isError(errors.New(ErrMsgNotAuthorizedPrefix+" some command here"), ErrMsgNotAuthorizedPrefix))
	assert.False(t, isError(errors.New("this is not an auth error"), ErrMsgNotAuthorizedPrefix))
	assert.False(t, isError(nil, ErrMsgNotAuthorizedPrefix))
}

func TestNewInitiator(t *testing.T) {
	testInitiator = NewInitiator(testConfig)
	assert.NotNil(t, testInitiator)
}

func TestControllerReplsetInitiatorInitAdminUser(t *testing.T) {
	testutils.DoSkipTest(t)

	user.UserAdmin = &mgo.User{
		Username: "testAdmin",
		Password: "testAdminPassword",
		Roles: []mgo.Role{
			mgo.RoleUserAdminAny,
		},
	}
	assert.NoError(t, testInitiator.initAdminUser(testSession))
	assert.NoError(t, user.RemoveUser(testSession, user.UserAdmin.Username, "admin"))
}

func TestControllerReplsetInitiatorInitUsers(t *testing.T) {
	testutils.DoSkipTest(t)

	user.SetSystemUsers([]*mgo.User{
		{
			Username: "testUser",
			Password: "testUserPassword",
			Roles: []mgo.Role{
				mgo.RoleReadWrite,
			},
		},
	})
	assert.NoError(t, testInitiator.initUsers(testSession))
	users := user.SystemUsers()
	assert.Len(t, users, 1)
	assert.NoError(t, user.RemoveUser(testSession, users[0].Username, "admin"))
}

func TestControllerReplsetInitiatorPrepareReplset(t *testing.T) {
	testutils.DoSkipTest(t)

	i := &Initiator{
		replInitTries: 0,
		config: &controller.Config{
			Replset: testutils.MongodbReplsetName,
			ReplsetInit: &controller.ConfigReplsetInit{
				PrimaryAddr:  testutils.MongodbHost + ":" + testutils.MongodbPrimaryPort,
				MaxReplTries: 1,
			},
		},
	}

	// test prepareReplset passes when using the Primary connection
	// on an already-initiated replset
	//
	// https://jira.percona.com/browse/CLOUD-46
	var out bytes.Buffer
	assert.NoError(t, i.prepareReplset(testSession, &out))

	// test using a Secondary node
	i.replInitTries = 0
	i.config.ReplsetInit.PrimaryAddr = testutils.MongodbHost + ":" + testutils.MongodbSecondary1Port
	testSecondarySession, err := i.getLocalhostNoAuthSession()
	if err != nil {
		t.Fatalf("cannot get secondary session: %v", err)
	}
	defer testSecondarySession.Close()

	// test prepareReplset fails gracefully on a Secondary that is
	// part of an already-initiated replset
	//
	// https://jira.percona.com/browse/CLOUD-46
	err = i.prepareReplset(testSecondarySession, &out)
	assert.Error(t, err)
	assert.True(t, isError(err, ErrMsgRsInitRequiresAuth))
}
