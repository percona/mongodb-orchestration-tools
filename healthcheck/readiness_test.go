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

package healthcheck

import (
	gotesting "testing"

	testing "github.com/percona/dcos-mongo-tools/common/testing"
	"github.com/percona/pmgo"
	"github.com/stretchr/testify/assert"
)

func TestReadinessCheck(t *gotesting.T) {
	testing.DoSkipTest(t)

	assert.NoError(t, testDBSession.Ping(), "Database ping error")
	state, err := ReadinessCheck(pmgo.NewSessionManager(testDBSession))
	assert.NoError(t, err, "healthcheck.ReadinessCheck() returned an error")
	assert.Equal(t, state, StateOk, "healthcheck.ReadinessCheck() returned incorrect state")
}
