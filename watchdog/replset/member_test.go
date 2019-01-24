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

	db "github.com/percona/mongodb-orchestration-tools/internal/db"
	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	"github.com/percona/mongodb-orchestration-tools/pkg"
	pkgDb "github.com/percona/mongodb-orchestration-tools/pkg/db"
	podDcos "github.com/percona/mongodb-orchestration-tools/pkg/pod/dcos"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod/mocks"
	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

// This test needs a mock of common/api PodTask
func TestWatchdogReplsetNewMongod(t *testing.T) {
	testutils.DoSkipTest(t)

	podTask := &mocks.Task{}
	port, _ := strconv.Atoi(testutils.MongodbPrimaryPort)
	podTask.On("GetMongoAddr").Return(&pkgDb.Addr{
		Host: "test." + pkg.DefaultServiceName + "." + podDcos.AutoIPDNSSuffix,
		Port: port,
	}, nil)
	podTask.On("GetMongoReplsetName").Return(testutils.MongodbReplsetName, nil)

	var err error
	testMongod, err = NewMongod(podTask, "mongo-"+testutils.MongodbReplsetName)
	assert.NoError(t, err, "replset.NewMongod() returned unexpected error")
	assert.NotNil(t, testMongod, "replset.NewMongod() should not return a nil Mongod")
}

func TestWatchdogReplsetMongodName(t *testing.T) {
	testutils.DoSkipTest(t)

	expected := "test." + pkg.DefaultServiceName + "." + podDcos.AutoIPDNSSuffix + ":" + testutils.MongodbPrimaryPort
	assert.Equal(t, expected, testMongod.Name(), ".Name() has unexpected output")
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
