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

package db

import (
	"bytes"
	"os"
	gotesting "testing"

	"github.com/percona/dcos-mongo-tools/common"
	testing "github.com/percona/dcos-mongo-tools/common/testing"
	"gopkg.in/mgo.v2"
)

var (
	testPrimarySession  *mgo.Session
	testLogBuffer       = new(bytes.Buffer)
	testPrimaryDbConfig = &Config{
		DialInfo: &mgo.DialInfo{
			Addrs:   []string{testing.MongodbHost + ":" + testing.MongodbPrimaryPort},
			Direct:  true,
			Timeout: testing.MongodbTimeout,
		},
		SSL: &SSLConfig{},
	}
)

func TestMain(m *gotesting.M) {
	common.SetupLogger(nil, common.GetLogFormatter("test"), testLogBuffer)
	exit := m.Run()
	if testPrimarySession != nil {
		testPrimarySession.Close()
	}
	os.Exit(exit)
}
