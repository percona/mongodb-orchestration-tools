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

package executor

import (
	"os"
	"testing"

	"github.com/percona/mongodb-orchestration-tools/executor/config"
	"github.com/percona/mongodb-orchestration-tools/executor/mocks"
)

var (
	testExecutor               *Executor
	testExecutorDaemon         *mocks.Daemon
	testExecutorDaemonNodeType = config.NodeType("MockDaemon")
	testExecutorConfig         = &config.Config{
		NodeType: testExecutorDaemonNodeType,
	}
	testQuitChan = make(chan bool)
)

func TestMain(m *testing.M) {
	exit := m.Run()
	if testExecutorDaemon != nil {
		testExecutorDaemon.Kill()
	}
	os.Exit(exit)
}
