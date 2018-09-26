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

package metrics

import (
	"testing"
	"time"

	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	mgostatsd "github.com/scullxbones/mgo-statsd"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

type MockPusher struct {
	statusChan chan *mgostatsd.ServerStatus
}

func NewMockPusher(statusChan chan *mgostatsd.ServerStatus) *MockPusher {
	return &MockPusher{
		statusChan: statusChan,
	}
}

func (p *MockPusher) GetServerStatus(session *mgo.Session) (*mgostatsd.ServerStatus, error) {
	return mgostatsd.GetServerStatus(session)
}

func (p *MockPusher) Push(status *mgostatsd.ServerStatus) error {
	p.statusChan <- status
	return nil
}

func TestExecutorMetricsNew(t *testing.T) {
	testutils.DoSkipTest(t)

	testMetrics = New(testConfig, testSession, NewMockPusher(testMetricsChan))
	assert.NotNil(t, testMetrics, ".New() should not return nil")
	assert.False(t, testMetrics.IsRunning(), ".IsRunning() should return false")
}

func TestExecutorMetricsName(t *testing.T) {
	testutils.DoSkipTest(t)

	assert.Equal(t, jobName, testMetrics.Name(), ".Name() has unexpected string")
}

func TestExecutorMetricsDoRun(t *testing.T) {
	testutils.DoSkipTest(t)

	assert.True(t, testMetrics.DoRun())

	dontRun := &Metrics{config: &Config{}}
	assert.False(t, dontRun.DoRun())
}

func TestExecutorMetricsRun(t *testing.T) {
	testutils.DoSkipTest(t)

	testLogBuffer.Reset()

	// start the metrics.Run() in a go routine and wait for ServerStatus struct
	stop := make(chan bool)
	go testMetrics.Run(&stop)
	serverStatus := <-testMetricsChan
	assert.NotZero(t, serverStatus.Uptime, "Uptime field in ServerStatus should be greater than zero")
	stop <- true

	// wait for the .Run() goroutine to stop
	var tries int
	for testMetrics.IsRunning() || tries < 60 {
		time.Sleep(testInterval)
		tries += 1
	}

	assert.Contains(t, testLogBuffer.String(), "Pushing DC/OS Metrics")
	assert.Contains(t, testLogBuffer.String(), "Stopping DC/OS Metrics pusher")
}
