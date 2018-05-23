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
	gotesting "testing"
	"time"

	testing "github.com/percona/dcos-mongo-tools/common/testing"
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

func TestExecutorMetricsNew(t *gotesting.T) {
	testing.DoSkipTest(t)

	testMetrics = New(testConfig, testSession, NewMockPusher(testMetricsChan))
	assert.NotNil(t, testMetrics, ".New() should not return nil")
	assert.False(t, testMetrics.IsRunning(), ".IsRunning() should return false")
}

func TestExecutorMetricsDoRun(t *gotesting.T) {
	assert.True(t, testMetrics.DoRun())

	dontRun := &Metrics{config: &Config{}}
	assert.False(t, dontRun.DoRun())
}

func TestExecutorMetricsRun(t *gotesting.T) {
	testing.DoSkipTest(t)
	testLogBuffer.Reset()

	// wait for a ServerStatus and then send a quit
	stop := make(chan bool)
	done := make(chan bool)
	go func() {
		status := <-testMetricsChan
		assert.NotZero(t, status.Uptime, "Uptime field in ServerStatus should be greater than zero")
		stop <- true
		for testMetrics.IsRunning() {
			time.Sleep(testInterval)
		}
		done <- true
	}()

	// start the metrics.Run() in a go routine and wait
	go testMetrics.Run(&stop)
	<-done

	assert.Contains(t, testLogBuffer.String(), "Pushing DC/OS Metrics")
	assert.Contains(t, testLogBuffer.String(), "Stopping DC/OS Metrics pusher")
}
