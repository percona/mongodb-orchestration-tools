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
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	currentUser, _  = user.Current()
	currentGroup, _ = user.LookupGroupId(currentUser.Gid)
	testMongod      *Mongod
	testMongodQuit  = make(chan bool)
	testConfig      = &Config{
		BinDir: "/usr/bin",
		User:   currentUser.Name,
		Group:  currentGroup.Name,
	}
)

func TestExecutorMongoDBNewMongod(t *testing.T) {
	testMongod = NewMongod(testConfig, &testMongodQuit)
	assert.NotNil(t, testMongod, ".NewMongod() should not return nil")
	assert.Contains(t, testMongod.commandBin, testConfig.BinDir)
	assert.Contains(t, testMongod.configFile, testConfig.ConfigDir)
}

// Test .getWiredTigerCacheSizeGB() mimics the cache-sizing logic described in the documentation:
// https://docs.mongodb.com/manual/reference/configuration-options/#storage.wiredTiger.engineConfig.cacheSizeGB
func TestExecutorMongoDBGetWiredTigerCacheSizeGB(t *testing.T) {
	mongod := &Mongod{
		config: &Config{
			WiredTigerCacheRatio: 0.5,
		},
	}

	mongod.config.TotalMemoryMB = 1024
	assert.Equal(t, minWiredTigerCacheSizeGB, mongod.getWiredTigerCacheSizeGB())

	mongod.config.TotalMemoryMB = 4 * 1024
	assert.Equal(t, 1.5, mongod.getWiredTigerCacheSizeGB())

	mongod.config.TotalMemoryMB = 64 * 1024
	assert.Equal(t, 31.5, mongod.getWiredTigerCacheSizeGB())

	mongod.config.TotalMemoryMB = 128 * 1024
	assert.Equal(t, 63.5, mongod.getWiredTigerCacheSizeGB())
}

func TestExecutorMongoDBMkdir(t *testing.T) {
	dir, _ := ioutil.TempDir("", "TestExecutorMongoDBMkdir")
	os.RemoveAll(dir)
	if _, err := os.Stat(dir); err == nil {
		assert.FailNow(t, "dir should not exist before .mkdir()")
	}

	uid, _ := strconv.Atoi(currentUser.Uid)
	gid, _ := strconv.Atoi(currentGroup.Gid)

	// bad path
	assert.Error(t, mkdir("C://%$#$@R", uid, gid, DefaultDirMode))

	// good .mkdir()
	assert.NoError(t, mkdir(dir, uid, gid, DefaultDirMode))
	defer os.RemoveAll(dir)

	stat, err := os.Stat(dir)
	if err != nil {
		assert.FailNow(t, ".mkdir() did not create a directory")
	}
	assert.True(t, stat.IsDir())
}
