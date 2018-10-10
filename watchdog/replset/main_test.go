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

package replset

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/percona/mongodb-orchestration-tools/internal/logger"
	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	wdConfig "github.com/percona/mongodb-orchestration-tools/watchdog/config"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	"gopkg.in/mgo.v2"
)

var (
	testDBSession       *mgo.Session
	testRsConfigManager rsConfig.Manager
	testRsConfigBefore  *rsConfig.Config
	testMongod          *Mongod
	testState           *State
	testLogBuffer       = new(bytes.Buffer)
	testWatchdogConfig  = &wdConfig.Config{
		Username:       "admin",
		Password:       "123456",
		ReplsetTimeout: time.Second,
	}
	testReplsetName   = "rs"
	testReplset       = &Replset{}
	testReplsetMongod = &Mongod{
		Host:        "test123",
		Port:        12345,
		Replset:     testReplsetName,
		PodName:     "mongod",
		ServiceName: "test",
	}
)

func TestMain(m *testing.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter("test"), testLogBuffer)
	if testutils.Enabled() {
		var err error
		testDBSession, err = testutils.GetSession(testutils.MongodbPrimaryPort)
		if err != nil {
			panic(err)
		}
		testRsConfigManager = rsConfig.New(testDBSession)
		_ = testRsConfigManager.Load()
		testRsConfigBefore = testRsConfigManager.Get()
	}
	exit := m.Run()
	if testDBSession != nil {
		_ = testRsConfigManager.Load()
		config := testRsConfigManager.Get()
		testRsConfigBefore.Version = config.Version + 1
		testRsConfigManager.Set(testRsConfigBefore)
		testDBSession.Close()
	}
	os.Exit(exit)
}
