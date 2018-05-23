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

const testHostname = "test.example.com"

type MockPusher struct {
	statusChan chan *mgostatsd.ServerStatus
}

func NewMockPusher(statusChan chan *mgostatsd.ServerStatus) *MockPusher {
	return &MockPusher{
		statusChan: statusChan,
	}
}

func (p *MockPusher) GetServerStatus(session *mgo.Session) (*mgostatsd.ServerStatus, error) {
	return &mgostatsd.ServerStatus{Host: testHostname}, nil
}

func (p *MockPusher) Push(status *mgostatsd.ServerStatus) error {
	p.statusChan <- status
	return nil
}

func TestExecutorMetricsNew(t *gotesting.T) {
	testing.DoSkipTest(t)

	testMetricsPusher = NewMockPusher(testMetricsChan)
	testMetrics = New(testConfig, testSession, testMetricsPusher)
	assert.NotNil(t, testMetrics, ".New() should not return nil")
	assert.False(t, testMetrics.IsRunning(), ".IsRunning() should return false")
}

func TestExecutorMetricsRun(t *gotesting.T) {
	testing.DoSkipTest(t)

	// start the metrics.Run() in a go routine
	go testMetrics.Run(&testMetricsRunQuit)
	tries := 0
	for !testMetrics.IsRunning() || tries < 10 {
		tries += 1
		time.Sleep(testInterval)

	}
}

func TestExecutorMetricsIsRunning(t *gotesting.T) {
	assert.True(t, testMetrics.IsRunning(), ".IsRunning() should return true at this stage")
}

func TestExecutorMetricsRunStop(t *gotesting.T) {
	// wait for a ServerStatus and then send a quit
	status := <-testMetricsChan
	testMetricsRunQuit <- true
	assert.Equal(t, testHostname, status.Host, "Host field in ServerStatus is unexpected")

	// wait for metrics.Run() goroutine to stop
	tries := 0
	for testMetrics.IsRunning() || tries < 10 {
		tries += 1
		time.Sleep(testInterval)
	}
	assert.False(t, testMetrics.IsRunning(), ".IsRunning() should return false at this stage")
}
