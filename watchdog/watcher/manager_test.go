package watcher

import (
	"strconv"
	gotesting "testing"
	"time"

	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/common/api/mocks"
	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"github.com/percona/dcos-mongo-tools/watchdog/replset"
	"github.com/stretchr/testify/assert"
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

func TestWatchdogWatcherNewManager(t *gotesting.T) {
	testManager = NewManager(testConfig, &testStopChan)
	assert.NotNil(t, testManager)
}

func TestWatchdogWatcherManagerWatch(t *gotesting.T) {
	testing.DoSkipTest(t)

	apiTask := &mocks.PodTask{}
	apiTask.On("Name").Return("test")
	apiTask.On("State").Return(api.PodTaskStateRunning)

	// primary
	var port int
	port, _ = strconv.Atoi(testing.MongodbPrimaryPort)
	testWatchRs.UpdateMember(&replset.Mongod{
		Host: testing.MongodbHost,
		Port: port,
		Task: apiTask,
	})

	// secondary1
	port, _ = strconv.Atoi(testing.MongodbSecondary1Port)
	testWatchRs.UpdateMember(&replset.Mongod{
		Host: testing.MongodbHost,
		Port: port,
		Task: apiTask,
	})

	// secondary2
	port, _ = strconv.Atoi(testing.MongodbSecondary2Port)
	testWatchRs.UpdateMember(&replset.Mongod{
		Host: testing.MongodbHost,
		Port: port,
		Task: apiTask,
	})

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

func TestWatchdogWatcherManagerHasWatcher(t *gotesting.T) {
	testing.DoSkipTest(t)
	assert.True(t, testManager.HasWatcher(rsName))
}

func TestWatchdogWatcherManagerGet(t *gotesting.T) {
	assert.NotNil(t, testManager.Get(rsName), ".Get() returned nil for existing watcher")
	assert.Nil(t, testManager.Get("does-not-exist"), ".Get() returned data for non-existing watcher")
}

func TestWatchdogWatcherManagerStop(t *gotesting.T) {
	testing.DoSkipTest(t)

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
