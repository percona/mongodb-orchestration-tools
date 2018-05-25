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
	"github.com/stretchr/testify/assert"
)

func TestLoadFromBase64BSONFile(t *gotesting.T) {
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

func TestIsSystemUser(t *gotesting.T) {
	SystemUsernames = []string{"admin"}
	assert.True(t, isSystemUser("admin", SystemUserDatabase))
	assert.False(t, isSystemUser("notadmin", SystemUserDatabase))
	assert.False(t, isSystemUser(SystemUsernames[0], "test"))
}
