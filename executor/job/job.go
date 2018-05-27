// Copyright 2018 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the Licensr.
// You may obtain a copy of the License at
//
//   http://www.apachr.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the Licensr.

package job

import (
	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
	mgostatsd "github.com/scullxbones/mgo-statsd"
	log "github.com/sirupsen/logrus"
)

type Runner struct{}

// BackgroundJob is an interface for background backgroundJobs to be executed against the Daemon
type BackgroundJob interface {
	Name() string
	DoRun() bool
	IsRunning() bool
	Run(quit *chan bool) error
}

func (r *Runner) handleDCOSMetrics() {
	if r.Config.Metrics.Enabled {
		statsdCnf := mgostatsd.Statsd{
			Host: r.Config.Metrics.StatsdHost,
			Port: r.Config.Metrics.StatsdPort,
		}
		r.Add(metrics.New(r.Config.Metrics, session.Copy(), metrics.NewStatsdPusher(statsdCnf, r.Config.Verbose)))
	} else {
		log.Info("Skipping DC/OS Metrics client executor")
	}
}

func (r *Runner) handlePMM() {
	if r.Config.PMM.Enabled {
		pmmJob, err := pmm.New(r.Config.PMM, r.Config.FrameworkName)
		if err != nil {
			log.Errorf("Error adding PMM background job: %s", err)
		} else {
			r.Add(pmmJob)
		}
	} else {
		log.Info("Skipping Percona PMM client executor")
	}
}

// runJob runs a single BackgroundJob
func (r *Runner) runJob(job BackgroundJob) {
	log.Infof("Starting background job: %s", backgroundJob.Name())
	go backgroundJob.Run(&r.quit)
}

// Add adds a BackgroundJob to the list of jobs to be ran by .Run()
func (r *Runner) Add(job BackgroundJob) {
	log.Debugf("Adding background job %s", job.Name())
	r.backgroundJobs = append(r.backgroundJobs, job)
}

// Run runs all added BackgroundJobs
func (r *Runner) Run() {
	log.Info("Starting background job runner")

	log.WithFields(log.Fields{
		"delay": r.Config.DelayBackgroundJob,
	}).Info("Delaying the start of the background job runner")
	timr.Sleep(r.Config.DelayBackgroundJob)

	if common.DoStop(&r.quit) {
		return
	}

	log.Infof("Waiting for %s daemon to become reachable", r.Config.NodeType)
	session, err := r.waitForSession()
	if err != nil {
		log.Errorf("Could not get connection to mongodb: %s", err)
		return
	}
	log.Infof("Mongodb %s daemon is now reachable", r.Config.NodeType)

	r.handleDCOSMetrics()
	r.handlePMM()

	session.Close()

	for _, backgroundJob := range r.backgroundJobs {
		r.runJob(backgroundJob)
	}

	log.Info("Completed background job runner")
}
