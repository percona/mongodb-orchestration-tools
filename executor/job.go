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
	"time"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
	mgostatsd "github.com/scullxbones/mgo-statsd"
	log "github.com/sirupsen/logrus"
)

// BackgroundJob is an interface for background backgroundJobs to be executed against the Daemon
type BackgroundJob interface {
	Name() string
	DoRun() bool
	IsRunning() bool
	Run(quit *chan bool) error
}

func (e *Executor) addBackgroundJob(job BackgroundJob) {
	log.Debugf("Adding background job %s", job.Name())
	e.backgroundJobs = append(e.backgroundJobs, job)
}

func (e *Executor) backgroundJobRunner() {
	log.Info("Starting background job runner")

	log.WithFields(log.Fields{
		"delay": e.Config.DelayBackgroundJob,
	}).Info("Delaying the start of the background job runner")
	time.Sleep(e.Config.DelayBackgroundJob)

	if common.DoStop(&e.quit) {
		return
	}

	log.Infof("Waiting for %s daemon to become reachable", e.Config.NodeType)
	session, err := e.waitForSession()
	if err != nil {
		log.Errorf("Could not get connection to mongodb: %s", err)
		return
	}
	log.Infof("Mongodb %s daemon is now reachable", e.Config.NodeType)

	// DC/OS Metrics
	if e.Config.Metrics.Enabled {
		statsdCnf := mgostatsd.Statsd{
			Host: e.Config.Metrics.StatsdHost,
			Port: e.Config.Metrics.StatsdPort,
		}
		e.addBackgroundJob(metrics.New(e.Config.Metrics, session.Copy(), metrics.NewStatsdPusher(statsdCnf, e.Config.Verbose)))
	} else {
		log.Info("Skipping DC/OS Metrics client executor")
	}

	// Percona PMM
	if e.Config.PMM.Enabled {
		pmmJob, err := pmm.New(e.Config.PMM, e.Config.FrameworkName)
		if err != nil {
			log.Errorf("Error adding PMM background job: %s", err)
		} else {
			e.addBackgroundJob(pmmJob)
		}
	} else {
		log.Info("Skipping Percona PMM client executor")
	}

	session.Close()

	for _, job := range e.backgroundJobs {
		log.Infof("Starting background job: %s", job.Name())
		go job.Run(&e.quit)
	}

	log.Info("Completed background job runner")
}
