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

package watcher

import (
	"os"
	gotesting "testing"
	"time"

	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/common/logger"
	"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"github.com/percona/dcos-mongo-tools/watchdog/replset"
)

var (
	testManager *WatcherManager
	testConfig  = &config.Config{
		Username:    testing.MongodbAdminUser,
		Password:    testing.MongodbAdminPassword,
		ReplsetPoll: 350 * time.Millisecond,
		SSL:         &db.SSLConfig{},
	}
	testStopChan = make(chan bool)
	testWatchRs  = replset.New(testConfig, testing.MongodbReplsetName)
	rsName       = testing.MongodbReplsetName
)

func TestMain(m *gotesting.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter("test"), os.Stdout)
	os.Exit(m.Run())
}
