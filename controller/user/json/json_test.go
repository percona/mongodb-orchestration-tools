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

package json

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

const (
	testUserFile       = "test/test-user.json"
	testUserBase64File = "test/test-user.json.base64"
	testUserFileBroken = "test/test-user-broken.json"
)

var (
	testUserJSON *JSON
)

func TestControllerUserJSONNewFromJSONFile(t *testing.T) {
	var err error

	// load valid file
	testUserJSON, err = NewFromJSONFile(testUserFile)
	assert.NoError(t, err)
	assert.NotNil(t, testUserJSON)
	assert.Equal(t, "testUser", testUserJSON.Username)

	// undefined file name
	_, err = NewFromJSONFile("")
	assert.Error(t, err)

	// non-existing file name
	_, err = NewFromJSONFile("/does/not/exist")
	assert.Error(t, err)

	// malformed json
	tmpfile, _ := ioutil.TempFile("", "TestControllerUserJSONNewFromJSONFile")
	defer os.Remove(tmpfile.Name())
	_, _ = tmpfile.Write([]byte("notjson"))
	_, err = NewFromJSONFile(tmpfile.Name())
	assert.Error(t, err)
}

func TestControllerUserJSONNewFromCLIPayloadFile(t *testing.T) {
	// not json+base64
	_, err := NewFromCLIPayloadFile(testUserFile)
	assert.Error(t, err)

	// good json+base64
	u, err := NewFromCLIPayloadFile(testUserBase64File)
	assert.NoError(t, err)
	assert.Len(t, u, 1)
	assert.Equal(t, testCLIPayload.Users, u)
}

func TestControllerUserJSONValidate(t *testing.T) {
	assert.NoError(t, testUserJSON.Validate("admin"))
	assert.Error(t, testUserJSON.Validate("notadmin"))

	u, _ := NewFromJSONFile(testUserFileBroken)
	assert.Error(t, u.Validate("admin"))
}

func TestControllerUserJSONToMgoUser(t *testing.T) {
	mgoUser, err := testUserJSON.ToMgoUser("admin")
	assert.NoError(t, err)
	assert.NotNil(t, mgoUser)

	// test for https://github.com/mesosphere/dcos-mongo/issues/218
	// ensure the test-user role for the "admin" db ends-up in the 'Roles' slice
	// and the role for the 2 other dbs ends up in 'OtherDBRoles' slice
	assert.Len(t, mgoUser.Roles, 1)
	assert.Equal(t, mgo.RoleClusterAdmin, mgoUser.Roles[0])
	assert.Len(t, mgoUser.OtherDBRoles, 2)
	assert.Len(t, mgoUser.OtherDBRoles["testApp1"], 1)
	assert.Equal(t, mgo.RoleReadWrite, mgoUser.OtherDBRoles["testApp1"][0])

	// ensure 'OtherDBRoles' logic fails if database is not "admin"
	mgoUser, err = testUserJSON.ToMgoUser("testApp1")
	assert.Error(t, err)
}
