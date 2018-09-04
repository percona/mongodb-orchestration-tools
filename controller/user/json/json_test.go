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
	testUserFile                   = "testdata/test-user.json"
	testUserCLIPayloadFile         = "testdata/test-user.json.base64"
	testUserCLIPayloadFileIssue218 = "testdata/test-user-issue218.json.base64"
	testUserFileBroken             = "testdata/test-user-broken.json"
	testUserFileNoQuotes           = "testdata/test-user-noquotes.json"
)

var (
	testUser       *User
	testCLIPayload = &CLIPayload{
		Users: []*User{
			{
				Username: "prodapp",
				Password: "123456",
				Roles: []*Role{
					{
						Database: "app",
						Role:     "readWrite",
					},
				},
			},
		},
	}
)

func TestControllerUserJSONNewFromFile(t *testing.T) {
	var err error

	// load valid file
	testUser, err = NewFromFile(testUserFile)
	assert.NoError(t, err)
	assert.NotNil(t, testUser)
	assert.Equal(t, "testUser", testUser.Username)

	// undefined file name
	_, err = NewFromFile("")
	assert.Error(t, err)

	// no quotes file (https://github.com/mesosphere/dcos-mongo/issues/257)
	_, err = NewFromFile(testUserFileNoQuotes)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user json file syntax error (see "+htmlDocsURL+"):")

	// non-existing file name
	_, err = NewFromFile("/does/not/exist")
	assert.Error(t, err)

	// malformed json
	tmpfile, _ := ioutil.TempFile("", "TestControllerUserJSONNewFromFile")
	defer os.Remove(tmpfile.Name())
	_, _ = tmpfile.Write([]byte("notjson"))
	_, err = NewFromFile(tmpfile.Name())
	assert.Error(t, err)
}

func TestControllerUserJSONNewFromCLIPayloadFile(t *testing.T) {
	// not json+base64
	_, err := NewFromCLIPayloadFile(testUserFile)
	assert.Error(t, err)

	// good json+base64
	u, err := NewFromCLIPayloadFile(testUserCLIPayloadFile)
	assert.NoError(t, err)
	assert.Len(t, u, 1)
	assert.Equal(t, testCLIPayload.Users, u)

	// test for https://github.com/mesosphere/dcos-mongo/issues/218:
	u, err = NewFromCLIPayloadFile(testUserCLIPayloadFileIssue218)
	assert.NoError(t, err)
	assert.Len(t, u, 1)
	assert.Equal(t, "tim", u[0].Username)
}

func TestControllerUserJSONValidate(t *testing.T) {
	assert.NoError(t, testUser.Validate("admin"))
	assert.Error(t, testUser.Validate("notadmin"))

	u, _ := NewFromFile(testUserFileBroken)
	assert.Error(t, u.Validate("admin"))
}

func TestControllerUserJSONToMgoUser(t *testing.T) {
	mgoUser, err := testUser.ToMgoUser("admin")
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
	mgoUser, err = testUser.ToMgoUser("testApp1")
	assert.Error(t, err)
}
