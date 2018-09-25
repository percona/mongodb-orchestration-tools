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

	"github.com/percona/dcos-mongo-tools/internal/pod"
	"github.com/percona/dcos-mongo-tools/internal/pod/mocks"
	"github.com/percona/dcos-mongo-tools/internal/testutils"
	"github.com/percona/dcos-mongo-tools/watchdog/replset"
	"github.com/stretchr/testify/assert"
)

func TestWatchdogWatcherNewManager(t *testing.T) {
	testManager = NewManager(testConfig, &testStopChan, pod.NewActivePods())
	assert.NotNil(t, testManager)
}

func TestWatchdogWatcherManagerWatch(t *testing.T) {
	testutils.DoSkipTest(t)

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
	testWatchRs.UpdateMember(mongod)

	// secondary1
	mongod.Port, _ = strconv.Atoi(testutils.MongodbSecondary1Port)
	testWatchRs.UpdateMember(mongod)

	// secondary2
	mongod.Port, _ = strconv.Atoi(testutils.MongodbSecondary2Port)
	testWatchRs.UpdateMember(mongod)

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

func TestWatchdogWatcherManagerHasWatcher(t *testing.T) {
	testutils.DoSkipTest(t)

	assert.True(t, testManager.HasWatcher(rsName))
}

func TestWatchdogWatcherManagerGet(t *testing.T) {
	testutils.DoSkipTest(t)

	assert.NotNil(t, testManager.Get(rsName), ".Get() returned nil for existing watcher")
	assert.Nil(t, testManager.Get("does-not-exist"), ".Get() returned data for non-existing watcher")
}

func TestWatchdogWatcherManagerStop(t *testing.T) {
	testutils.DoSkipTest(t)

	testManager.Stop(rsName)
	tries := 0
	for tries < 20 {
		if !testManager.Get(rsName).IsRunning() {
			return
		}
		time.Sleep(time.Second)
		tries++
	}
	assert.FailNow(t, "Failed to stop watcher after 20 tries")
}
