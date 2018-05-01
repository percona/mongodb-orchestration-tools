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

package testing

import (
	"os"
	gotesting "testing"
	"time"

	"gopkg.in/mgo.v2"
)

var (
	EnvEnableDBTests   = "ENABLE_MONGODB_TESTS"
	EnableDBTests      = os.Getenv(EnvEnableDBTests)
	MongoDBPrimaryPort = os.Getenv("TEST_PRIMARY_PORT")
)

const (
	MongoDBPrimaryHost = "127.0.0.1"
	MongoDBTimeout     = time.Duration(10) * time.Second
)

func Enabled() bool {
	return EnableDBTests == "true"
}

func PrimaryDialInfo() *mgo.DialInfo {
	if Enabled() && MongoDBPrimaryPort != "" {
		return &mgo.DialInfo{
			Addrs:   []string{MongoDBPrimaryHost + ":" + MongoDBPrimaryPort},
			Direct:  true,
			Timeout: MongoDBTimeout,
		}
	}
	return nil
}

func DoSkipTest(t *gotesting.T) {
	if !Enabled() {
		t.Skipf("Skipping test, env var %s is not 'true'", EnvEnableDBTests)
	}
}
