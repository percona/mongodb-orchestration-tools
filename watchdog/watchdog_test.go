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
	"testing"
	"time"

	"github.com/percona/dcos-mongo-tools/internal/api/mocks"
	"github.com/percona/dcos-mongo-tools/internal/db"
	"github.com/percona/dcos-mongo-tools/internal/logger"
	"github.com/percona/dcos-mongo-tools/internal/pod"
	"github.com/percona/dcos-mongo-tools/internal/pod/dcos"
	"github.com/percona/dcos-mongo-tools/internal/testutils"
	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"github.com/stretchr/testify/assert"
)

var (
	testWatchdog *Watchdog
	testQuitChan = make(chan bool)
	testConfig   = &config.Config{
		FrameworkName: "test",
		Username:      testutils.MongodbAdminUser,
		Password:      testutils.MongodbAdminPassword,
		APIPoll:       time.Millisecond * 100,
		ReplsetPoll:   time.Millisecond * 100,
		MetricsPort:   "65432",
		SSL:           &db.SSLConfig{},
	}
	testAPIClient = &mocks.Client{}
)

func TestMain(m *testing.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter("test"), os.Stdout)
	os.Exit(m.Run())
}

func TestWatchdogNew(t *testing.T) {
	testWatchdog = New(testConfig, &testQuitChan, testAPIClient)
	assert.NotNil(t, testWatchdog, ".New() returned nil")
}

func TestWatchdogDoIgnorePod(t *testing.T) {
	testConfig.IgnorePods = []string{"ignore-me"}
	assert.True(t, testWatchdog.doIgnorePod("ignore-me"))
	assert.False(t, testWatchdog.doIgnorePod("dont-ignore-me"))
}

func TestWatchdogRun(t *testing.T) {
	testAPIClient.On("GetPodURL").Return("http://test")
	testAPIClient.On("GetPods").Return(&pod.Pods{"test"}, nil)

	tasks := make([]pod.Task, 0)
	tasks = append(tasks, &dcos.DCOSTask{
		Data: &dcos.DCOSTaskData{
			Info: &dcos.DCOSTaskInfo{},
		},
	})
	testAPIClient.On("GetPodTasks", "test").Return(tasks, nil)

	go testWatchdog.Run()
	time.Sleep(time.Millisecond * 150)
	close(testQuitChan)
}
