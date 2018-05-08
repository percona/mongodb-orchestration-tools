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

	"github.com/percona/dcos-mongo-tools/common/db"
	//"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

//func TestNewMongod(t *gotesting.T) {
//	testing.DoSkipTest(t)
//	apiTask := new(MockApiTask)
//	mongod, err := NewMongod(apiTask, "frameworkNameHere", "mongo-"+testing.MongodbReplsetName)
//	assert.NoError(t, err, "replset.NewMongod() returned unexpected error")
//}

func TestMongodName(t *gotesting.T) {
	mongod := &Mongod{
		Host: "test1234",
		Port: 123456,
	}
	assert.Equal(t, mongod.Name(), "test1234:123456")
}

func TestIsBackupNode(t *gotesting.T) {
	mongod := &Mongod{
		Host:    "test1234",
		Port:    123456,
		PodName: "notabackupnode",
	}
	assert.False(t, mongod.IsBackupNode(), "mongod.IsBackupNode() should be false")
	mongod.PodName = backupPodNamePrefix + "-something"
	assert.True(t, mongod.IsBackupNode(), "mongod.IsBackupNode() should be true")
}

func TestDBConfig(t *gotesting.T) {
	mongod := &Mongod{
		Host: "test1234",
		Port: 123456,
	}
	sslConfig := &db.SSLConfig{}
	assert.Equal(t, mongod.DBConfig(sslConfig), &db.Config{
		DialInfo: &mgo.DialInfo{
			Addrs:    []string{"test1234:123456"},
			Direct:   true,
			FailFast: true,
			Timeout:  db.DefaultMongoDBTimeoutDuration,
		},
		SSL: sslConfig,
	}, "mongod.DBConfig() response is not valid")
}
