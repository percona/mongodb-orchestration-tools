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

	"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

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
