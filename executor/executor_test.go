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
	"errors"
	"testing"

	"github.com/percona/mongodb-orchestration-tools/executor/mocks"
	"github.com/stretchr/testify/assert"
)

func TestExecutorNew(t *testing.T) {
	testExecutor = New(testExecutorConfig, &testQuitChan)
	assert.NotNil(t, testExecutor, ".New() should not return nil")
}

func TestExecutorRun(t *testing.T) {
	testExecutorDaemon = &mocks.Daemon{}
	testExecutorDaemon.On("Kill").Return(nil)

	// test successful .Start()
	testExecutorDaemon.On("Start").Return(nil).Once()
	err := testExecutor.Run(testExecutorDaemon)
	assert.NoError(t, err, ".Run() should not return an error")

	// test failed .Start()
	testExecutorDaemon.On("Start").Return(errors.New("test failure"))
	err = testExecutor.Run(testExecutorDaemon)
	assert.Error(t, err, ".Run() should return an error after failed .Start()")
}
