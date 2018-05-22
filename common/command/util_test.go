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

package command

import (
	"os"
	"os/user"
	gotesting "testing"

	"github.com/stretchr/testify/assert"
)

func TestGetUserId(t *gotesting.T) {
	_, err := GetUserId("this-user-should-not-exist")
	assert.Error(t, err, ".GetUserId() should return error due to missing user")

	user := os.Getenv("USER")
	if user == "" {
		user = "nobody"
	}
	uid, err := GetUserId(user)
	assert.NoError(t, err, ".GetUserId() for current user should not return an error")
	assert.NotZero(t, uid, ".GetUserId() should return a uid that is not zero")
}

func TestGetGroupId(t *gotesting.T) {
	_, err := GetGroupId("this-group-should-not-exist")
	assert.Error(t, err, ".GetGroupId() should return error due to missing group")

	currentUser, err := user.Current()
	assert.NoError(t, err, "could not get current user")
	group, err := user.LookupGroupId(currentUser.Gid)
	assert.NoError(t, err, "could not get current user group")

	gid, err := GetGroupId(group.Name)
	assert.NoError(t, err, ".GetGroupId() for current user group should not return an error")
	assert.NotEqual(t, -1, gid, ".GetGroupId() should return a gid that is not -1")
}
