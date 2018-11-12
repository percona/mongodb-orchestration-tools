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
	"strconv"
	"testing"
	"time"

	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod/mocks"
	"github.com/percona/mongodb-orchestration-tools/watchdog/replset"
	"github.com/stretchr/testify/assert"
)

func TestWatchdogWatcherManagerWatch(t *testing.T) {
	testutils.DoSkipTest(t)

	testManager = NewManager(testConfig, &testStopChan, pod.NewPods())
	assert.NotNil(t, testManager)

	apiTask := &mocks.Task{}
	apiTask.On("Name").Return("test")

	apiTaskState := &mocks.TaskState{}
	apiTaskState.On("String").Return("OK")
	apiTask.On("State").Return(apiTaskState)

	// primary
	port, _ := strconv.Atoi(testutils.MongodbPrimaryPort)
	mongod := &replset.Mongod{
		Host: testutils.MongodbHost,
		Port: port,
		Task: apiTask,
	}
	assert.NoError(t, testWatchRs.UpdateMember(mongod))

	// secondary1
	mongod.Port, _ = strconv.Atoi(testutils.MongodbSecondary1Port)
	assert.NoError(t, testWatchRs.UpdateMember(mongod))

	// secondary2
	mongod.Port, _ = strconv.Atoi(testutils.MongodbSecondary2Port)
	assert.NoError(t, testWatchRs.UpdateMember(mongod))

	assert.Nil(t, testManager.Get("does-not-exist"), ".Get() returned data for non-existing watcher")
	assert.False(t, testManager.HasWatcher(rsName))

	go testManager.Watch(testWatchRs)

	tries := 0
	for tries < 20 {
		if testManager.HasWatcher(rsName) && testManager.Get(rsName).IsRunning() {
			return
		}
		time.Sleep(time.Second)
		tries++
	}
	assert.FailNow(t, "failed to start watcher after 20 tries")
}

func TestWatchdogWatcherManagerClose(t *testing.T) {
	testutils.DoSkipTest(t)

	assert.Contains(t, testManager.watchers, rsName)
	testManager.Close()

	tries := 0
	for tries < 20 {
		if !testManager.Get(rsName).IsRunning() {
			return
		}
		time.Sleep(time.Second)
		tries++
	}
	assert.FailNow(t, "Failed to close watcher manager after 20 tries")

	assert.NotContains(t, testManager.watchers, rsName)
}
