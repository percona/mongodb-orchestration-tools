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
	"net"
	"strconv"
	gotesting "testing"

	testing "github.com/percona/dcos-mongo-tools/common/testing"
	mgostatsd "github.com/scullxbones/mgo-statsd"
	"github.com/stretchr/testify/assert"
)

var (
	testStatsdPusher *StatsdPusher
	testServerStatus *mgostatsd.ServerStatus
	testUDPHost      = "localhost"
	testUDPPort      = 9999
)

func TestExecutorMetricsNewStatsdPusher(t *gotesting.T) {
	testStatsdPusher = NewStatsdPusher(mgostatsd.Statsd{
		Host: testUDPHost,
		Port: testUDPPort,
	}, false)
	assert.NotNil(t, testStatsdPusher, ".NewStatsdPusher() should not return nil")
	assert.False(t, testStatsdPusher.verbose, ".NewStatsdPusher() should return with a verbose set to false")
}

func TestExecutorMetricsPusherGetServerStatus(t *gotesting.T) {
	testing.DoSkipTest(t)

	var err error
	testServerStatus, err = testStatsdPusher.GetServerStatus(testSession)
	assert.NoError(t, err, ".GetServerStatus() should not return an error")
	assert.NotNil(t, testServerStatus, ".GetServerStatus() should not return an nil ServerStatus struct")
	assert.NotZero(t, testServerStatus.Uptime, ".GetServerStatus() should not return a ServerStatus struct with an Uptime greater than 1")
}

func TestExecutorMetricsPusherPush(t *gotesting.T) {
	testing.DoSkipTest(t)

	addr, err := net.ResolveUDPAddr("udp", ":"+strconv.Itoa(testUDPPort))
	if err != nil {
		assert.FailNowf(t, "Could not create UDP address for testing StatsdPusher: %v", err.Error())
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		assert.FailNowf(t, "Could not listen on UDP address for testing StatsdPusher: %v", err.Error())
	}
	defer conn.Close()

	data := make(chan []byte)
	go func() {
		buf := make([]byte, 1024)
		for {
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				assert.FailNowf(t, "Error reading from UDP port: %v", err.Error())
				break
			}
			data <- buf[0:n]
			break
		}
	}()

	assert.NoError(t, testStatsdPusher.Push(testServerStatus), ".Push() should not return an error")
	assert.Regexp(t, "^\\.\\S+-\\d+\\.connections.current:", string(<-data), ".Push() sent unexpected payload")
}
