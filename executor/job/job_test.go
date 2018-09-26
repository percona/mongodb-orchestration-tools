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

package job

import (
	"testing"
	"time"

	"github.com/percona/mongodb-orchestration-tools/executor/config"
	"github.com/percona/mongodb-orchestration-tools/executor/metrics"
	"github.com/percona/mongodb-orchestration-tools/executor/mocks"
	"github.com/percona/mongodb-orchestration-tools/executor/pmm"
	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestExecutorJobAdd(t *testing.T) {
	mockJob := &mocks.BackgroundJob{}
	mockJob.On("Name").Return(t.Name())
	r := &Runner{jobs: make([]BackgroundJob, 0)}
	assert.Len(t, r.jobs, 0)
	r.add(mockJob)
	assert.Len(t, r.jobs, 1)
}

func TestExecutorJobRun(t *testing.T) {
	testutils.DoSkipTest(t)

	quit := make(chan bool, 1)
	config := &config.Config{
		DelayBackgroundJob: time.Millisecond,
		Metrics: &metrics.Config{
			Enabled:  false,
			Interval: 500 * time.Millisecond,
		},
		PMM: &pmm.Config{
			Enabled: false,
		},
	}
	r := New(config, testDBSession, &quit)

	// run with disabled jobs
	assert.NotPanics(t, func() { r.Run() })
	quit <- true

	// run with enabled jobs
	config.Metrics.Enabled = true
	config.PMM.Enabled = true
	r2 := New(config, testDBSession, &quit)
	assert.NotPanics(t, func() { r2.Run() })
	quit <- true
}
