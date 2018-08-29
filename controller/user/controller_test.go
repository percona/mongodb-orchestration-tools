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

package user

import (
	gotesting "testing"

	"github.com/percona/dcos-mongo-tools/internal"
	"github.com/percona/dcos-mongo-tools/internal/api/mocks"
	"github.com/percona/dcos-mongo-tools/internal/testing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

func TestControllerUserNew(t *gotesting.T) {
	testing.DoSkipTest(t)

	mockAPI := &mocks.Client{}
	mockAPI.On("GetEndpoint", common.DefaultMongoDBMongodEndpointName).Return(testing.GetMongoPortAPIEndpoint(), nil)

	var err error
	testController, err = NewController(testControllerConfig, mockAPI)
	assert.NoError(t, err, ".NewController() should not return an error")
	assert.NotNil(t, testController, ".NewController() should return a Controller that is not nil")
	assert.NotNil(t, testController.session, ".NewController() should return a Controller with a session field that is not nil")
	assert.NoError(t, testController.session.Ping(), ".NewController() should return a Controller with a session that is pingable")
	assert.Equal(t, mgo.Primary, testController.session.Mode(), ".NewController() should return a Controller with a session that is in mgo.Primary mode")
}

func TestControllerUserControllerUpdateUsers(t *gotesting.T) {
	testing.DoSkipTest(t)

	assert.Error(
		t,
		checkUserExists(testSession, testControllerConfig.User.Username, testControllerConfig.User.Database),
		"mongo user should not exist before .UpdateUsers() call",
	)
	assert.NoError(t, testController.UpdateUsers(), ".UpdateUsers() should not return an error")
	assert.NoError(
		t,
		checkUserExists(testSession, testControllerConfig.User.Username, testControllerConfig.User.Database),
		"mongo user should exist after .UpdateUsers() call",
	)
}

func TestControllerUserControllerRemoveUser(t *gotesting.T) {
	testing.DoSkipTest(t)

	assert.NoError(t, testController.RemoveUser(), ".RemoveUser() should not return an error")
	assert.Error(
		t,
		checkUserExists(testSession, testControllerConfig.User.Username, testControllerConfig.User.Database),
		"mongo user should not exist after .RemoveUser() call",
	)
}

func TestControllerUserControllerReloadSystemUsers(t *gotesting.T) {
	testing.DoSkipTest(t)

	for _, user := range testSystemUsers {
		assert.Error(
			t,
			checkUserExists(testSession, user.Username, SystemUserDatabase),
			"mongo test system user should not exist before .ReloadSystemUsers() call",
		)
	}

	SystemUsers = testSystemUsers
	assert.NoError(t, testController.ReloadSystemUsers(), ".ReloadSystemUsers() should not return an error")
	for _, user := range testSystemUsers {
		assert.NoError(
			t,
			checkUserExists(testSession, user.Username, SystemUserDatabase),
			"mongo test system user should exist after .ReloadSystemUsers() call",
		)
		testControllerConfig.User.Username = user.Username
		assert.NoError(t, testController.RemoveUser(), ".RemoveUser() should not return an error")
		assert.Error(
			t,
			checkUserExists(testSession, user.Username, SystemUserDatabase),
			"mongo test system user should not exist after .RemoveUser() call",
		)
	}
}

func TestControllerUserControllerClose(t *gotesting.T) {
	testing.DoSkipTest(t)

	testController.Close()
	assert.Nil(t, testController.session, "Controller session should not nil after .Close()")
}
