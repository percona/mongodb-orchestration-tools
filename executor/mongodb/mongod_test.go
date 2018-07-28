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
	"strconv"
	gotesting "testing"
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

func TestExecutorMongoDBIsStarted(t *gotesting.T) {
	assert.False(t, testMongod.IsStarted())
}

func TestExecutorMongoDBStart(t *gotesting.T) {
	// get random open TCP port for mongod
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	listenPort := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// get tmpdir for mongod dbPath
	tmpDBPath, err := ioutil.TempDir("", "TestExecutorMongoDBStartDBPath")
	assert.NoError(t, err)

	// get tmpdir
	tmpPath, err := ioutil.TempDir("", "TestExecutorMongoDBStartTmpPath")
	assert.NoError(t, err)
	testConfig.TmpDir = tmpPath

	// make the security.keyFile tmpfile
	tmpKeyFile, err := ioutil.TempFile("", "TestExecutorMongoDBStartKeyFile")
	assert.NoError(t, err)
	_, err = tmpKeyFile.Write([]byte("123456789101112"))
	assert.NoError(t, err)
	tmpKeyFile.Close()

	// make the config tmpdir
	tmpConfigDir, err := ioutil.TempDir("", "TestExecutorMongoDBStartConfigDir")
	assert.NoError(t, err)
	testConfig.ConfigDir = tmpConfigDir

	defer func() {
		os.Remove(tmpKeyFile.Name())
		os.RemoveAll(tmpDBPath)
		os.RemoveAll(tmpPath)
		os.RemoveAll(tmpConfigDir)
	}()

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

	err = config.Write(testConfig.ConfigDir + "/mongod.conf")
	assert.NoError(t, err)

	testMongod = NewMongod(testConfig)
	assert.NotNil(t, testMongod, ".NewMongod() should not return nil")
	assert.NoError(t, testMongod.Start(), ".Start() should not return an error")

	var tries int
	for tries < 30 {
		session, err := mgo.Dial(config.Net.BindIp + ":" + strconv.Itoa(config.Net.Port))
		if err == nil && session.Ping() == nil {
			session.Close()
			break
		}
		time.Sleep(time.Second)
		tries++
	}
	if tries > 30 {
		assert.FailNowf(t, "could not connect to tmp mongod: %v", err.Error())
	}

	assert.NoError(t, testMongod.Kill())
	testMongod.Wait()
	assert.False(t, testMongod.IsStarted())
}
