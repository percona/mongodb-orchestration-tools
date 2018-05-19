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

package command

import (
	gotesting "testing"

	ps "github.com/mitchellh/go-ps"
	"github.com/stretchr/testify/assert"
)

func TestNewCommand(t *gotesting.T) {
	var err error
	testCommand, err = New("echo", []string{"hello", "world"}, testCurrentUser.Name, testCurrentGroup.Name)
	assert.NoError(t, err, ".New() should not return an error")
	assert.Equal(t, "echo", testCommand.Bin, ".New() has incorrect Bin")
	assert.Equal(t, testCurrentUser.Name, testCommand.User, ".New() has incorrect User")
	assert.Equal(t, testCurrentGroup.Name, testCommand.Group, ".New() has incorrect Group")
}

func TestIsRunningFalse(t *gotesting.T) {
	assert.False(t, testCommand.IsRunning(), ".IsRunning() should be false")
}

func TestStart(t *gotesting.T) {
	assert.NoError(t, testCommand.Start(), ".Start() should not return an error")
}

func TestIsRunning(t *gotesting.T) {
	assert.True(t, testCommand.IsRunning(), ".IsRunning() should be true")
}

func TestWait(t *gotesting.T) {
	testCommand.Wait()
	assert.False(t, testCommand.IsRunning(), ".IsRunning() should be false after .Wait()")
}

func TestKill(t *gotesting.T) {
	killCommand, err := New("sleep", []string{"120"}, testCurrentUser.Name, testCurrentGroup.Name)
	assert.NoError(t, err, ".New() should not return an error")
	assert.NoError(t, killCommand.Start(), ".Start() should not return an error")

	// check process started
	killCommandProc := killCommand.command.Process
	proc, _ := ps.FindProcess(killCommandProc.Pid)
	assert.NotNil(t, proc)

	// kill the process before it's done
	killCommand.Kill()

	// check the process died
	proc, err = ps.FindProcess(killCommandProc.Pid)
	assert.Nil(t, err)
	assert.Nil(t, proc)
}

func TestCombinedOutput(t *gotesting.T) {
	coCommand, err := New("echo", []string{"hello", "world"}, testCurrentUser.Name, testCurrentGroup.Name)
	bytes, err := coCommand.CombinedOutput()
	assert.NoError(t, err, ".CombinedOutput() should not return an error")
	assert.NotEmpty(t, bytes, ".CombinedOutput() should not return empty bytes")
	assert.Equal(t, "hello world\n", string(bytes), ".CombinedOutput() has unexpected output")
}
