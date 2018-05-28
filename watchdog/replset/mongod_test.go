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

package replset

import (
	gotesting "testing"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

// This test needs a mock of common/api PodTask
func TestWatchdogReplsetNewMongod(t *gotesting.T) {
	testing.DoSkipTest(t)

	var err error
	apiTask := &api.PodTask{
		Info: &api.PodTaskInfo{
			Name: "test",
			Command: &api.PodTaskCommand{
				Environment: &api.PodTaskCommandEnvironment{
					Variables: []*api.PodTaskCommandEnvironmentVariable{
						{Name: common.EnvMongoDBPort, Value: testing.MongodbPrimaryPort},
						{Name: common.EnvMongoDBReplset, Value: testing.MongodbReplsetName},
					},
				},
			},
		},
	}
	testMongod, err = NewMongod(apiTask, common.DefaultFrameworkName, "mongo-"+testing.MongodbReplsetName)
	assert.NoError(t, err, "replset.NewMongod() returned unexpected error")
	assert.NotNil(t, testMongod, "replset.NewMongod() should not return a nil Mongod")
}

func TestWatchdogReplsetMongodName(t *gotesting.T) {
	expected := "test." + common.DefaultFrameworkName + "." + api.AutoIpDnsSuffix + ":" + testing.MongodbPrimaryPort
	assert.Equal(t, expected, testMongod.Name(), ".Name() has unexpected output")
}

func TestWatchdogReplsetMongodIsBackupNode(t *gotesting.T) {
	assert.False(t, testMongod.IsBackupNode(), "mongod.IsBackupNode() should be false")
	mongod := &Mongod{
		Host:    "test1234",
		Port:    123456,
		PodName: backupPodNamePrefix + "-something",
	}
	assert.True(t, mongod.IsBackupNode(), "mongod.IsBackupNode() should be true")
}

func TestWatchdogReplsetMongodDBConfig(t *gotesting.T) {
	sslConfig := &db.SSLConfig{}
	assert.Equal(t, testMongod.DBConfig(sslConfig), &db.Config{
		DialInfo: &mgo.DialInfo{
			Addrs:    []string{testMongod.Name()},
			Direct:   true,
			FailFast: true,
			Timeout:  db.DefaultMongoDBTimeoutDuration,
		},
		SSL: sslConfig,
	}, "mongod.DBConfig() response is not valid")
}
