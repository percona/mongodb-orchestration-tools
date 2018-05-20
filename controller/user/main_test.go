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

package user

import (
	"bytes"
	"os"
	gotesting "testing"
	"time"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/percona/dcos-mongo-tools/controller"
)

var (
	testController       *Controller
	testLogBuffer        = new(bytes.Buffer)
	testControllerConfig = &controller.Config{
		SSL: &db.SSLConfig{},
		User: &controller.ConfigUser{
			EndpointName:    "mongo-port",
			MaxConnectTries: 1,
			RetrySleep:      time.Second,
		},
		FrameworkName:     "test",
		Replset:           testing.MongodbReplsetName,
		UserAdminUser:     testing.MongodbAdminUser,
		UserAdminPassword: testing.MongodbAdminPassword,
	}
)

func TestMain(m *gotesting.M) {
	common.SetupLogger(nil, common.GetLogFormatter("test"), testLogBuffer)
	defer func() {
		if testController != nil {
			testController.Close()
		}
	}()
	os.Exit(m.Run())
}
