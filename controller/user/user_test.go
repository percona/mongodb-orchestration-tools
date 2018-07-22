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
	"path/filepath"
	gotesting "testing"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

func TestControllerUserLoadFromBase64BSONFile(t *gotesting.T) {
	_, err := loadFromBase64BSONFile("/this/should/not/exist/...")
	assert.Error(t, err, ".loadFromBase64BSONFile() should return an error for missing file")

	_, err = loadFromBase64BSONFile(common.RelPathToAbs(filepath.Join(testDirRelPath, testBase64BSONFileMalformedBase64)))
	assert.Error(t, err, ".loadFromBase64BSONFile() should return an error due to malformed base64")

	_, err = loadFromBase64BSONFile(common.RelPathToAbs(filepath.Join(testDirRelPath, testBase64BSONFileMalformedBSON)))
	assert.Error(t, err, ".loadFromBase64BSONFile() should return an error due to malformed bson")

	change, err := loadFromBase64BSONFile(common.RelPathToAbs(filepath.Join(testDirRelPath, testBase64BSONFile)))
	assert.NoError(t, err, ".loadFromBase64BSONFile() should not return an error")
	assert.NotNil(t, change, ".loadFromBase64BSONFile() should not return a nil UserChangeData struct")
	assert.Len(t, change.Users, 1, ".loadFromBase64BSONFile() should not return exactly one user")
	assert.Equal(t, testBase64BSONUser, change.Users[0], ".loadFromBase64BSONFile() returned an unexpected mgo.User")
}

func TestControllerUserIsSystemUser(t *gotesting.T) {
	SystemUsernames = []string{"admin"}
	assert.True(t, isSystemUser("admin", SystemUserDatabase))
	assert.False(t, isSystemUser("notadmin", SystemUserDatabase))
	assert.False(t, isSystemUser(SystemUsernames[0], "test"))
}

func TestControllerUserUpdateUser(t *gotesting.T) {
	testing.DoSkipTest(t)

	// TODO: use mock of pmgo.SessionManager
	// ensure fixes for https://github.com/mesosphere/dcos-mongo/issues/218
	assert.NoError(t, UpdateUser(testSession, &mgo.User{
		Username: "testUserUpdate",
		Password: "123456",
		Roles:    []mgo.Role{},
		OtherDBRoles: map[string][]mgo.Role{
			"products": []mgo.Role{
				mgo.RoleReadWrite,
			},
		},
	}, "admin"))

	// no roles or otherDBRoles
	assert.Error(t, UpdateUser(testSession, &mgo.User{
		Username: "testUserUpdate",
		Password: "123456",
		Roles:    []mgo.Role{},
	}, "admin"))

	// valid user
	assert.NoError(t, UpdateUser(testSession, &mgo.User{
		Username: "testUserUpdate",
		Password: "123456",
		Roles:    []mgo.Role{mgo.RoleRead},
	}, "admin"))
	assert.NoError(t, checkUserExists(testSession, "testUserUpdate", "admin"))
	RemoveUser(testSession, "testUserUpdate", "admin")
}

func TestControllerUserRemoveUser(t *gotesting.T) {
	testing.DoSkipTest(t)

	// valid user
	assert.NoError(t, UpdateUser(testSession, &mgo.User{
		Username: "testUserRemove",
		Password: "123456",
		Roles:    []mgo.Role{mgo.RoleRead},
	}, "admin"))
	assert.NoError(t, checkUserExists(testSession, "testUserRemove", "admin"))

	// test .RemoveUser()
	assert.NoError(t, RemoveUser(testSession, "testUserRemove", "admin"))
	assert.Error(t, checkUserExists(testSession, "testUserRemove", "admin"))
}
