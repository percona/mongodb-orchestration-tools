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

const (
	envEnableDBTests        = "ENABLE_MONGODB_TESTS"
	envMongoDBReplsetName   = "TEST_RS_NAME"
	envMongoDBPrimaryPort   = "TEST_PRIMARY_PORT"
	envMongoDBAdminUser     = "TEST_ADMIN_USER"
	envMongoDBAdminPassword = "TEST_ADMIN_PASSWORD"
)

var (
	enableDBTests        = os.Getenv(envEnableDBTests)
	MongodbReplsetName   = os.Getenv(envMongoDBReplsetName)
	MongodbPrimaryHost   = "127.0.0.1"
	MongodbPrimaryPort   = os.Getenv(envMongoDBPrimaryPort)
	MongodbAdminUser     = os.Getenv(envMongoDBAdminUser)
	MongodbAdminPassword = os.Getenv(envMongoDBAdminPassword)
	MongodbTimeout       = time.Duration(10) * time.Second
)

// Enabled returns a boolean reflecting whether testing against Mongodb should occur
func Enabled() bool {
	return enableDBTests == "true"
}

// getPrimaryDialInfo returns a *mgo.DialInfo configured for testing against a Mongodb Primary
func getPrimaryDialInfo(t *gotesting.T) *mgo.DialInfo {
	if Enabled() {
		if MongodbReplsetName == "" {
			t.Fatalf("Replica set name env var %s is not set", MongodbReplsetName)
		} else if MongodbPrimaryPort == "" {
			t.Fatalf("Primary port env var %s is not set", envMongoDBPrimaryPort)
		} else if MongodbAdminUser == "" {
			t.Fatalf("Admin user env var %s is not set", envMongoDBAdminUser)
		} else if MongodbAdminPassword == "" {
			t.Fatalf("Admin password env var %s is not set", envMongoDBAdminPassword)
		}
		return &mgo.DialInfo{
			Addrs:          []string{MongodbPrimaryHost + ":" + MongodbPrimaryPort},
			Direct:         true,
			Timeout:        MongodbTimeout,
			Username:       MongodbAdminUser,
			Password:       MongodbAdminPassword,
			ReplicaSetName: MongodbReplsetName,
		}
	}
	return nil
}

func GetPrimarySession(t *gotesting.T) (*mgo.Session, error) {
	dialInfo := getPrimaryDialInfo(t)
	if dialInfo == nil {
		t.Fatal("Could not build dial info for Primary")
	}
	return mgo.DialWithInfo(dialInfo)
}

// DoSkipTest handles the conditional skipping of tests, based on the output of .Enabled()
func DoSkipTest(t *gotesting.T) {
	if !Enabled() {
		t.Skipf("Skipping test, env var %s is not 'true'", envEnableDBTests)
	}
}
