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
	"strconv"
	"testing"

	"github.com/percona/dcos-mongo-tools/internal"
	"github.com/percona/dcos-mongo-tools/internal/db"
	"github.com/percona/dcos-mongo-tools/internal/pod/dcos"
	"github.com/percona/dcos-mongo-tools/internal/pod/mocks"
	"github.com/percona/dcos-mongo-tools/internal/testutils"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

// This test needs a mock of common/api PodTask
func TestWatchdogReplsetNewMongod(t *testing.T) {
	testutils.DoSkipTest(t)

	podTask := &mocks.Task{}
	port, _ := strconv.Atoi(testutils.MongodbPrimaryPort)
	podTask.On("GetMongoAddr").Return(&db.Addr{
		Host: "test." + internal.DefaultFrameworkName + "." + dcos.DCOSAutoIPDnsSuffix,
		Port: port,
	}, nil)
	podTask.On("GetMongoReplsetName").Return(testutils.MongodbReplsetName, nil)

	var err error
	testMongod, err = NewMongod(podTask, internal.DefaultFrameworkName, "mongo-"+testutils.MongodbReplsetName)
	assert.NoError(t, err, "replset.NewMongod() returned unexpected error")
	assert.NotNil(t, testMongod, "replset.NewMongod() should not return a nil Mongod")
}

func TestWatchdogReplsetMongodName(t *testing.T) {
	testutils.DoSkipTest(t)

	expected := "test." + internal.DefaultFrameworkName + "." + dcos.DCOSAutoIPDnsSuffix + ":" + testutils.MongodbPrimaryPort
	assert.Equal(t, expected, testMongod.Name(), ".Name() has unexpected output")
}

func TestWatchdogReplsetMongodIsBackupNode(t *testing.T) {
	testutils.DoSkipTest(t)

	assert.False(t, testMongod.IsBackupNode(), "mongod.IsBackupNode() should be false")
	mongod := &Mongod{
		Host:    "test1234",
		Port:    123456,
		PodName: backupPodNamePrefix + "-something",
	}
	assert.True(t, mongod.IsBackupNode(), "mongod.IsBackupNode() should be true")
}

func TestWatchdogReplsetMongodDBConfig(t *testing.T) {
	testutils.DoSkipTest(t)

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
