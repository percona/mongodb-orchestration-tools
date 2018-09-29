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
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	mdbconfig "github.com/timvaillancourt/go-mongodb-config/config"
	"gopkg.in/mgo.v2"
)

var (
	currentUser, _  = user.Current()
	currentGroup, _ = user.LookupGroupId(currentUser.Gid)
	testMongod      *Mongod
	testConfig      = &Config{
		BinDir:    "/usr/bin",
		ConfigDir: "test",
		User:      currentUser.Name,
		Group:     currentGroup.Name,
	}
)

func TestExecutorMongoDBNewMongod(t *testing.T) {
	testStateChan := make(chan *os.ProcessState)
	testMongod = NewMongod(testConfig, testStateChan)
	assert.NotNil(t, testMongod, ".NewMongod() should not return nil")
	assert.Contains(t, testMongod.commandBin, testConfig.BinDir)
	assert.Contains(t, testMongod.configFile, testConfig.ConfigDir)
	assert.Equal(t, "mongod", testMongod.Name())
}

func TestExecutorMongoDBLoadConfig(t *testing.T) {
	testStateChan := make(chan *os.ProcessState)

	mongod := NewMongod(&Config{ConfigDir: "testdata"}, testStateChan)
	config, err := mongod.loadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 27017, config.Net.Port)

	// test missing config
	mongod = NewMongod(&Config{ConfigDir: "/does/not/exist"}, testStateChan)
	_, err = mongod.loadConfig()
	assert.Error(t, err)
}

func TestExecutorMongoDBProcessWiredTigerConfig(t *testing.T) {
	// copy testdata/mongod.conf to a temp dir
	tempDir, _ := ioutil.TempDir("", t.Name())
	defer os.RemoveAll(tempDir)
	src, err := ioutil.ReadFile("testdata/mongod.conf")
	assert.NoError(t, err)
	assert.NoError(t, ioutil.WriteFile(filepath.Join(tempDir, "mongod.conf"), src, 0644))

	testStateChan := make(chan *os.ProcessState)
	config := &Config{
		ConfigDir:            tempDir,
		TotalMemoryMB:        8 * 1024,
		WiredTigerCacheRatio: 0.5,
	}
	mongod := NewMongod(config, testStateChan)
	mongodConfig, err := mongod.loadConfig()
	assert.NoError(t, err)
	assert.Nil(t, mongodConfig.Storage.WiredTiger)

	assert.NoError(t, mongod.processWiredTigerConfig(mongodConfig))
	assert.NotNil(t, mongodConfig.Storage.WiredTiger)
	assert.NotNil(t, mongodConfig.Storage.WiredTiger.EngineConfig)
	assert.Equal(t, 3.5, mongodConfig.Storage.WiredTiger.EngineConfig.CacheSizeGB)
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

func TestExecutorMongoDBIsStarted(t *testing.T) {
	assert.False(t, testMongod.IsStarted())
}

func TestExecutorMongoDBStart(t *testing.T) {
	if os.Getenv("TEST_EXECUTOR_MONGODB") != "true" {
		t.Logf("Skipping test because TEST_EXECUTOR_MONGODB is not 'true'")
		return
	}

	// get random open TCP port for mongod
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	listenPort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// get tmpdir for mongod dbPath
	tmpDBPath, _ := ioutil.TempDir("", "TestExecutorMongoDBStartDBPath")
	defer os.RemoveAll(tmpDBPath)

	// get tmpdir
	testConfig.TmpDir, _ = ioutil.TempDir("", "TestExecutorMongoDBStartTmpPath")
	defer os.RemoveAll(testConfig.TmpDir)

	// make the security.keyFile tmpfile
	tmpKeyFile, _ := ioutil.TempFile("", "TestExecutorMongoDBStartKeyFile")
	_, err = tmpKeyFile.Write([]byte("123456789101112"))
	assert.NoError(t, err)
	tmpKeyFile.Close()
	defer os.Remove(tmpKeyFile.Name())

	// make the config tmpdir
	testConfig.ConfigDir, err = ioutil.TempDir("", "TestExecutorMongoDBStartConfigDir")
	defer os.RemoveAll(testConfig.ConfigDir)

	// write mongod config
	config := &mdbconfig.Config{
		Net: &mdbconfig.Net{
			BindIp: "127.0.0.1",
			Port:   listenPort,
		},
		Security: &mdbconfig.Security{
			KeyFile: tmpKeyFile.Name(),
		},
		Storage: &mdbconfig.Storage{
			DbPath: tmpDBPath,
		},
	}
	assert.NoError(t, config.Write(testConfig.ConfigDir+"/mongod.conf"))

	testStateChan := make(chan *os.ProcessState)
	testMongod = NewMongod(testConfig, testStateChan)
	assert.NotNil(t, testMongod, ".NewMongod() should not return nil")
	assert.NoError(t, testMongod.Start(), ".Start() should not return an error")

	var tries = 0
	var maxTries = 60
	var session *mgo.Session
	for tries < maxTries {
		if session == nil {
			s, err := mgo.Dial(config.Net.BindIp + ":" + strconv.Itoa(config.Net.Port))
			if err == nil {
				session = s
			}
		} else if session.Ping() == nil {
			session.Close()
			break
		}
		time.Sleep(time.Second)
		tries++
	}
	if tries > maxTries {
		assert.FailNowf(t, "could not connect to tmp mongod: %v", err.Error())
	}

	assert.True(t, testMongod.IsStarted())
	assert.NoError(t, testMongod.Kill())
	testMongod.Wait()
	assert.False(t, testMongod.IsStarted())
}
