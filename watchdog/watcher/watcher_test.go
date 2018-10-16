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

package watcher

import (
	"testing"

	"github.com/percona/mongodb-orchestration-tools/pkg/pod"
	"github.com/percona/mongodb-orchestration-tools/watchdog/replset"
	"github.com/stretchr/testify/assert"
	rsConfig "github.com/timvaillancourt/go-mongodb-replset/config"
	rsStatus "github.com/timvaillancourt/go-mongodb-replset/status"
)

func TestGetScaledDownMembers(t *testing.T) {
	rs := replset.New(nil, "test")
	rs.UpdateMember(&replset.Mongod{
		Host:    "scaled-down",
		Port:    27017,
		PodName: "testPod",
	})

	pods := pod.NewPods()
	pods.Set([]string{"testPod"})
	w := &Watcher{
		activePods: pods,
		replset:    rs,
		state: &replset.State{
			Config: &rsConfig.Config{
				Members: []*rsConfig.Member{{
					Host: "scaled-down:27017",
				}},
			},
			Status: &rsStatus.Status{
				Members: []*rsStatus.Member{{
					Name:  "scaled-down:27017",
					State: rsStatus.MemberStatePrimary,
				}},
			},
		},
	}

	// test empty result (zero down hosts)
	assert.Len(t, w.getScaledDownMembers(), 0)

	// test empty result (1 down host but pod still exists in 'activePods')
	// this test simulates an unplanned member/task failure
	w.state.Status.Members[0].State = rsStatus.MemberStateDown
	assert.Len(t, w.getScaledDownMembers(), 0)

	// test scaled down (1 down host AND pod doesnt exist in 'activePods')
	// this test simulates a scaling-down of replset members/tasks
	w.activePods.Set([]string{})
	scaledDown := w.getScaledDownMembers()
	assert.Len(t, scaledDown, 1)
	assert.Equal(t, "scaled-down:27017", scaledDown[0].Host)
}
