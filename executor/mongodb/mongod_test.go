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

func TestExecutorMongoDBGetWiredTigerCacheSizeGB(t *gotesting.T) {
	mongod := &Mongod{
		config: &Config{WiredTigerCacheRatio: 0.5},
	}
	assert.Equal(t, minWiredTigerCacheSizeGB, mongod.getWiredTigerCacheSizeGB(int64(1*gigaByte)))
	assert.Equal(t, 1.5, mongod.getWiredTigerCacheSizeGB(int64(4*gigaByte)))
	assert.Equal(t, 31.5, mongod.getWiredTigerCacheSizeGB(int64(64*gigaByte)))
	assert.Equal(t, 63.5, mongod.getWiredTigerCacheSizeGB(int64(128*gigaByte)))
}

func TestExecutorMongoDBGetMemoryLimitBytes(t *gotesting.T) {
	// does not exist
	_, err := getMemoryLimitBytes("/does/not/exist")
	assert.Error(t, err)

	// test that .getMemoryLimitBytes() returns 0 (and no error)
	// if the memory limit is equal to the noMemoryLimit const
	limitFile, _ := ioutil.TempFile("", t.Name())
	defer os.Remove(limitFile.Name())
	data := []byte(strconv.Itoa(int(noMemoryLimit)) + "\n")
	_, err = limitFile.Write(data)
	assert.NoError(t, err)
	_, err = getMemoryLimitBytes(limitFile.Name())
	assert.NoError(t, err)

	// check we can write and read successfully without error
	limitFile2, _ := ioutil.TempFile("", t.Name())
	defer os.Remove(limitFile2.Name())
	data2 := []byte(strconv.Itoa(int(gigaByte)) + "\n")
	_, err = limitFile2.Write(data2)
	assert.NoError(t, err)
	limit, err := getMemoryLimitBytes(limitFile2.Name())
	assert.NoError(t, err)
	assert.Equal(t, int64(gigaByte), limit)
}

func TestExecutorMongoDBMkdir(t *gotesting.T) {
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
