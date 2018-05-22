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

package common

import (
	"os"
	"os/user"
	gotesting "testing"
	"time"

	dcosmongotools "github.com/percona/dcos-mongo-tools"
	"github.com/stretchr/testify/assert"
)

func TestCommonDoStop(t *gotesting.T) {
	stop := make(chan bool)
	stopped := make(chan bool)
	go func(stop *chan bool, stopped chan bool) {
		for !DoStop(stop) {
			time.Sleep(time.Second)
		}
		stopped <- true
	}(&stop, stopped)
	stop <- true

	var tries int
	for tries < 3 {
		select {
		case _ = <-stopped:
			return
		default:
			tries += 1
		}
	}
	t.Error("Stop did not work")
}

func TestCommonDoStopFalse(t *gotesting.T) {
	stop := make(chan bool)
	stopped := make(chan bool)
	go func(stop *chan bool, stopped chan bool) {
		for !DoStop(stop) {
			time.Sleep(time.Second)
		}
		stopped <- true
	}(&stop, stopped)

	var tries int
	for tries < 3 {
		select {
		case _ = <-stopped:
			tries += 1
		default:
			stop <- true
			return
		}
	}
	t.Error("Stop did not work")
}

func TestCommonNewApp(t *gotesting.T) {
	testApp, _ := NewApp("test help", "git-commit-here", "branch-name-here")
	appModel := testApp.Model()
	assert.Contains(t, appModel.Version, "common.test version "+dcosmongotools.Version+"\ngit commit git-commit-here, branch branch-name-here\ngo version", "kingpin.Application version is unexpected")
	assert.Equal(t, appAuthor, appModel.Author, "kingpin.Application author is unexpected")
	assert.Equal(t, "test help", appModel.Help, "kingpin.Application help is unexpected")
}

func TestCommonGetUserID(t *gotesting.T) {
	_, err := GetUserID("this-user-should-not-exist")
	assert.Error(t, err, ".GetUserID() should return error due to missing user")

	user := os.Getenv("USER")
	if user == "" {
		user = "nobody"
	}
	uid, err := GetUserID(user)
	assert.NoError(t, err, ".GetUserID() for current user should not return an error")
	assert.NotZero(t, uid, ".GetUserID() should return a uid that is not zero")
}

func TestCommonGetGroupID(t *gotesting.T) {
	_, err := GetGroupID("this-group-should-not-exist")
	assert.Error(t, err, ".GetGroupID() should return error due to missing group")

	currentUser, err := user.Current()
	assert.NoError(t, err, "could not get current user")
	group, err := user.LookupGroupId(currentUser.Gid)
	assert.NoError(t, err, "could not get current user group")

	gid, err := GetGroupID(group.Name)
	assert.NoError(t, err, ".GetGroupID() for current user group should not return an error")
	assert.NotEqual(t, -1, gid, ".GetGroupID() should return a gid that is not zero")
}
