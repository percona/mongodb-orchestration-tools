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

package mongodb

import (
	"io/ioutil"
	"os"
	"os/user"
	"strconv"
	gotesting "testing"

	"github.com/stretchr/testify/assert"
)

var (
	currentUser, _  = user.Current()
	currentGroup, _ = user.LookupGroupId(currentUser.Gid)
	testMongod      *Mongod
	testConfig      = &Config{
		BinDir: "/usr/bin",
		User:   currentUser.Name,
		Group:  currentGroup.Name,
	}
)

func TestExecutorMongoDBNewMongod(t *gotesting.T) {
	testMongod = NewMongod(testConfig)
	assert.NotNil(t, testMongod, ".NewMongod() should not return nil")
	assert.Contains(t, testMongod.commandBin, testConfig.BinDir)
	assert.Contains(t, testMongod.configFile, testConfig.ConfigDir)
}

func TestExecutorMongoDBMkdir(t *gotesting.T) {
	dir, _ := ioutil.TempDir("", "TestExecutorMongoDBMkdir")
	os.RemoveAll(dir)
	if _, err := os.Stat(dir); err == nil {
		assert.FailNow(t, "dir should not exist before .mkdir()")
	}

	// bad uid + gid
	assert.Error(t, mkdir(dir, 999999, 99999, 0777))

	// good .mkdir()
	uid, _ := strconv.Atoi(currentUser.Uid)
	gid, _ := strconv.Atoi(currentGroup.Gid)
	assert.NoError(t, mkdir(dir, uid, gid, DefaultDirMode))
	defer os.RemoveAll(dir)
	if _, err := os.Stat(dir); err != nil {
		assert.FailNow(t, ".mkdir() did not create a directory")
	}
	stat, _ := os.Stat(dir)
	assert.True(t, stat.IsDir())
}
