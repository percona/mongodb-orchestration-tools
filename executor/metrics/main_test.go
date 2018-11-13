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
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/percona/mongodb-orchestration-tools/internal/logger"
	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	mgostatsd "github.com/scullxbones/mgo-statsd"
	"gopkg.in/mgo.v2"
)

var (
	testMetrics     *Metrics
	testLogBuffer   = new(bytes.Buffer)
	testMetricsChan = make(chan *mgostatsd.ServerStatus)
	testSession     *mgo.Session
	testInterval    = time.Duration(100) * time.Millisecond
	testConfig      = &Config{
		Enabled:    true,
		StatsdHost: "localhost",
		StatsdPort: 9999,
		Interval:   testInterval,
	}
)

func TestMain(m *testing.M) {
	logger.SetupLogger(nil, logger.GetLogFormatter(), testLogBuffer)

	if testutils.Enabled() {
		var err error
		testSession, err = testutils.GetSession(testutils.MongodbPrimaryPort)
		if err != nil {
			panic(err)
		}
	}

	exit := m.Run()
	if testSession != nil {
		testSession.Close()
	}
	os.Exit(exit)
}
