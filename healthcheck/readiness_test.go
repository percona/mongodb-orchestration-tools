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
	"gopkg.in/mgo.v2"
)

func TestReadinessCheck(t *gotesting.T) {
	testing.DoSkipTest(t)

	dialInfo := testing.PrimaryDialInfo()
	if dialInfo == nil {
		t.Error("Could not build dial info for Primary")
		return
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		t.Errorf("Database connection error: %s", err)
	}
	defer session.Close()

	err = session.Ping()
	if err != nil {
		t.Error("Database ping failed")
	}

	state, err := ReadinessCheck(session)
	if err != nil {
		t.Error("healthcheck.ReadinessCheck() returned an error: %s", err)
	} else if state != StateOk {
		t.Errorf("healthcheck.ReadinessCheck() returned non-ok state: %v", state)
	}
}
