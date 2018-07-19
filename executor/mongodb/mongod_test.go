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

package mongodb

import (
	gotesting "testing"

	"github.com/stretchr/testify/assert"
)

var (
	testMongod *Mongod
	testConfig = &Config{
		ConfigDir: "/tmp",
		BinDir:    "/usr/bin",
		TmpDir:    "/tmp",
		User:      "nobody",
		Group:     "nogroup",
	}
)

func TestExecutorMongoDBNewMongod(t *gotesting.T) {
	testMongod = NewMongod(testConfig)
	assert.NotNil(t, testMongod, ".NewMongod() should not return nil")
	assert.Contains(t, testMongod.commandBin, testConfig.BinDir)
	assert.Contains(t, testMongod.configFile, testConfig.ConfigDir)
}
