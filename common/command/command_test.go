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

func TestCombinedOutput(t *gotesting.T) {
	_, err := testCommand.CombinedOutput()
	assert.NoError(t, err, ".CombinedOutput() should not return an error")
}
