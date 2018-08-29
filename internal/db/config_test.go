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

package db

import (
	gotesting "testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/mgo.v2"
)

func TestCommonDBUri(t *gotesting.T) {
	config := &Config{
		DialInfo: &mgo.DialInfo{
			Addrs:    []string{"test:1234"},
			Username: "admin",
			Password: "123456",
		},
		SSL: &SSLConfig{
			Enabled: false,
		},
	}
	assert.Equal(t, "mongodb://admin:123456@test:1234", config.Uri(), ".Uri() returned invalid uri")

	config.SSL.Enabled = true
	assert.Equal(t, "mongodb://admin:123456@test:1234?ssl=true", config.Uri(), ".Uri() returned invalid uri")

	config.DialInfo.ReplicaSetName = "test"
	assert.Equal(t, "mongodb://admin:123456@test:1234?replicaSet=test&ssl=true", config.Uri(), ".Uri() returned invalid uri")
}
