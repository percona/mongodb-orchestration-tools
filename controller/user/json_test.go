package user

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
	testUserJSON *UserJSON
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
	tmpfile.Write([]byte("notjson"))
	_, err = NewFromJSONFile(tmpfile.Name())
	assert.Error(t, err)
}

func TestControllerUserJSONNewFromJSONBase64File(t *testing.T) {
	_, err := NewFromJSONBase64File(testUserBase64File)
	assert.NoError(t, err)
}

func TestControllerUserJSONValidate(t *testing.T) {
	assert.NoError(t, testUserJSON.validate("admin"))
	assert.Error(t, testUserJSON.validate("notadmin"))

	u, _ := NewFromJSONFile(testUserFileBroken)
	assert.Error(t, u.validate("admin"))
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
