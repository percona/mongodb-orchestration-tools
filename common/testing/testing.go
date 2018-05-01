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
	envMongoDBPrimaryPort   = "TEST_PRIMARY_PORT"
	envMongoDBAdminUser     = "TEST_ADMIN_USER"
	envMongoDBAdminPassword = "TEST_ADMIN_PASSWORD"
	mongodbPrimaryHost      = "127.0.0.1"
)

var (
	enableDBTests        = os.Getenv(envEnableDBTests)
	mongodbPrimaryPort   = os.Getenv(envMongoDBPrimaryPort)
	mongodbAdminUser     = os.Getenv(envMongoDBAdminUser)
	mongodbAdminPassword = os.Getenv(envMongoDBAdminPassword)
	mongodbTimeout       = time.Duration(10) * time.Second
)

// Enabled returns a boolean reflecting whether testing against mongodb should occur
func Enabled() bool {
	return enableDBTests == "true"
}

// PrimaryDialInfo returns a *mgo.DialInfo configured for testing against a mongodb Primary
func PrimaryDialInfo(t *gotesting.T) *mgo.DialInfo {
	if Enabled() {
		if mongodbPrimaryPort == "" {
			t.Fatalf("Primary port env var %s is not set", envMongoDBPrimaryPort)
		} else if mongodbAdminUser == "" {
			t.Fatalf("Admin user env var %s is not set", envMongoDBAdminUser)
		} else if mongodbAdminPassword == "" {
			t.Fatalf("Admin password env var %s is not set", envMongoDBAdminPassword)
		}
		return &mgo.DialInfo{
			Addrs:    []string{mongodbPrimaryHost + ":" + mongodbPrimaryPort},
			Direct:   true,
			Timeout:  mongodbTimeout,
			Username: mongodbAdminUser,
			Password: mongodbAdminPassword,
		}
	}
	return nil
}

// DoSkipTest handles the conditional skipping of tests, based on the output of .Enabled()
func DoSkipTest(t *gotesting.T) {
	if !Enabled() {
		t.Skipf("Skipping test, env var %s is not 'true'", envEnableDBTests)
	}
}
