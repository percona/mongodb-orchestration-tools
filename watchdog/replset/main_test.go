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
	gotesting "testing"
	"time"

	"github.com/percona/dcos-mongo-tools/common/logger"
	testing "github.com/percona/dcos-mongo-tools/common/testing"
	wdConfig "github.com/percona/dcos-mongo-tools/watchdog/config"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	"gopkg.in/mgo.v2"
)

var (
	testDBSession       *mgo.Session
	testRsConfigManager rsConfig.Manager
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
		Host:          "test123",
		Port:          12345,
		Replset:       testReplsetName,
		PodName:       "mongod",
		FrameworkName: "test",
	}
)

func TestMain(m *gotesting.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter("test"), testLogBuffer)
	if testing.Enabled() {
		var err error
		testDBSession, err = testing.GetSession(testing.MongodbPrimaryPort)
		if err != nil {
			panic(err)
		}
		testRsConfigManager = rsConfig.New(testDBSession)
	}
	exit := m.Run()
	if testDBSession != nil {
		testDBSession.Close()
	}
	os.Exit(exit)
}
