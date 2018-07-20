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

package watchdog

import (
	"os"
	gotesting "testing"
	"time"

	//"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/common/api/mocks"
	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/common/logger"
	"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"github.com/stretchr/testify/assert"
)

var (
	testWatchdog *Watchdog
	testQuitChan = make(chan bool)
	testConfig   = &config.Config{
		FrameworkName: "test",
		Username:      testing.MongodbAdminUser,
		Password:      testing.MongodbAdminPassword,
		APIPoll:       time.Millisecond * 100,
		ReplsetPoll:   time.Millisecond * 100,
		MetricsPort:   "65432",
		SSL:           &db.SSLConfig{},
	}
	testAPIClient = &mocks.Client{}
)

func TestMain(m *gotesting.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter("test"), os.Stdout)
	os.Exit(m.Run())
}

func TestWatchdogNew(t *gotesting.T) {
	testWatchdog = New(testConfig, &testQuitChan, testAPIClient)
	assert.NotNil(t, testWatchdog, ".New() returned nil")
}

func TestWatchdogDoIgnorePod(t *gotesting.T) {
	testConfig.IgnorePods = []string{"ignore-me"}
	assert.True(t, testWatchdog.doIgnorePod("ignore-me"))
	assert.False(t, testWatchdog.doIgnorePod("dont-ignore-me"))
}

func TestWatchdogRun(t *gotesting.T) {
	testAPIClient.On("GetPodURL").Return("http://test")
	testAPIClient.On("GetPods").Return(&api.Pods{"test"}, nil)
	testAPIClient.On("GetPodTasks", "test").Return([]api.PodTask{
		&api.PodTaskHTTP{
			Info: &api.PodTaskInfo{
				//		Name: "test-mongod",
				//		Command: &api.PodTaskCommand{
				//		Environment: &api.PodTaskCommandEnvironment{
				//		Variables: []*api.PodTaskCommandEnvironmentVariable{
				//			{Name: common.EnvMongoDBPort, Value: testing.MongodbPrimaryPort},
				//			{Name: common.EnvMongoDBReplset, Value: testing.MongodbReplsetName},
				//		},
				//	},
				//	Value: "mongodb-executor-",
				//},
			},
		},
	}, nil)

	go testWatchdog.Run()
	time.Sleep(time.Millisecond * 150)
	close(testQuitChan)
}
