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
	"strconv"
	"testing"
	"time"

	"github.com/percona/mongodb-orchestration-tools/internal/db"
	"github.com/percona/mongodb-orchestration-tools/internal/logger"
	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	pkgDb "github.com/percona/mongodb-orchestration-tools/pkg/db"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod/mocks"
	"github.com/percona/mongodb-orchestration-tools/watchdog/config"
	"github.com/percona/mongodb-orchestration-tools/watchdog/metrics"
	"github.com/stretchr/testify/assert"
)

var (
	testWatchdog *Watchdog
	testQuitChan = make(chan bool)
	testConfig   = &config.Config{
		ServiceName: "test",
		Username:    testutils.MongodbAdminUser,
		Password:    testutils.MongodbAdminPassword,
		APIPoll:     time.Millisecond * 75,
		ReplsetPoll: time.Millisecond * 100,
		SSL:         &db.SSLConfig{},
	}
)

func TestMain(m *testing.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter("test"), os.Stdout)
	os.Exit(m.Run())
}

func TestWatchdogDoIgnorePod(t *testing.T) {
	testConfig.IgnorePods = []string{"ignore-me"}
	watchdog := &Watchdog{config: testConfig}
	assert.True(t, watchdog.doIgnorePod("ignore-me"))
	assert.False(t, watchdog.doIgnorePod("dont-ignore-me"))
}

func TestWatchdogRun(t *testing.T) {
	testPodSource := &mocks.Source{}
	wMetrics := metrics.NewCollector()
	testWatchdog := New(testConfig, testPodSource, wMetrics, &testQuitChan)
	assert.NotNil(t, testWatchdog, ".New() returned nil")

	testPodSource.On("Name").Return("test")
	testPodSource.On("URL").Return("http://test")
	testPodSource.On("Pods").Return([]string{"testPod"}, nil)

	tasks := make([]pod.Task, 0)
	for i, portStr := range []string{testutils.MongodbPrimaryPort, testutils.MongodbSecondary1Port, testutils.MongodbSecondary2Port} {
		port, _ := strconv.Atoi(portStr)

		mockTask := &mocks.Task{}
		mockTask.On("GetMongoAddr").Return(&pkgDb.Addr{
			Host: testutils.MongodbHost,
			Port: port,
		}, nil)
		mockTask.On("GetMongoReplsetName").Return(testutils.MongodbReplsetName, nil)
		mockTask.On("IsRunning").Return(true)
		mockTask.On("IsTaskType", pod.TaskTypeMongod).Return(true)
		mockTask.On("Name").Return(t.Name() + "-" + strconv.Itoa(i))

		mockTaskState := &mocks.TaskState{}
		mockTaskState.On("String").Return("RUNNING")
		mockTask.On("State").Return(mockTaskState)

		tasks = append(tasks, mockTask)
	}
	testPodSource.On("GetTasks", "testPod").Return(tasks, nil)

	// start watchdog
	go testWatchdog.Run()
	tries := 0
	for tries < 100 {
		if testWatchdog.getRunning() {
			break
		}
		time.Sleep(50 * time.Millisecond)
		tries++
	}
	if tries >= 100 {
		assert.FailNow(t, "could not start watchdog after 100 tries")
	}

	// test watchdog started a watcher and fetched data
	watcher := testWatchdog.watcherManager.Get(testutils.MongodbReplsetName)
	assert.NotNil(t, watcher)
	tries = 0
	for tries < 100 {
		state := watcher.State()
		if watcher.IsRunning() && state != nil && state.GetStatus() != nil {
			break
		}
		time.Sleep(50 * time.Millisecond)
		tries++
	}
	if tries >= 100 {
		assert.FailNow(t, "could not start watchdog after 100 tries")
	}
	state := watcher.State()
	assert.NotNil(t, state)
	assert.NotNil(t, state.GetStatus())

	// stop watchdog
	close(testQuitChan)
	tries = 0
	for tries < 100 {
		if !testWatchdog.getRunning() {
			break
		}
		time.Sleep(50 * time.Millisecond)
		tries++
	}
	if tries >= 100 {
		assert.FailNow(t, "could not stop watchdog after 100 tries")
	}
}
