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
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

// BackgroundJob is an interface for background backgroundJobs to be executed against the Daemon
type BackgroundJob interface {
	Name() string
	DoRun() bool
	IsRunning() bool
	Run() error
}

// Daemon is an interface for the mongodb (mongod or mongos) daemon
type Daemon interface {
	IsStarted() bool
	Start() error
	Wait()
}

type Executor struct {
	Config         *Config
	PMM            *pmm.PMM
	Metrics        *metrics.Metrics
	backgroundJobs []BackgroundJob
}

func New(config *Config) *Executor {
	e := &Executor{
		Config:         config,
		PMM:            pmm.New(config.PMM, config.FrameworkName),
		Metrics:        metrics.New(config.Metrics),
		backgroundJobs: make([]BackgroundJob, 0),
	}

	// Percona PMM
	if e.PMM.DoRun() {
		e.addBackgroundJob(e.PMM)
	} else {
		log.Info("Skipping Percona PMM client executor")
	}

	// DC/OS Metrics
	if e.Metrics.DoRun() {
		e.addBackgroundJob(e.Metrics)
	} else {
		log.Info("Skipping DC/OS Metrics client executor")
	}

	return e
}

func (e *Executor) waitForSession() (*mgo.Session, error) {
	return common.WaitForSession(
		e.Config.DB,
		0,
		e.Config.ConnectRetrySleep,
	)
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

	log.Infof("Waiting for %s daemon to become reachable", e.Config.NodeType)
	session, err := e.waitForSession()
	if err != nil {
		log.Errorf("Could not get connection to mongodb: %s", err)
		return
	}
	log.Infof("Mongodb %s daemon is now reachable", e.Config.NodeType)
	session.Close()

	for _, job := range e.backgroundJobs {
		log.Infof("Starting background job: %s", job.Name())
		err := job.Run()
		if err != nil {
			log.Errorf("Background job %s failed: %s", job.Name(), err)
		}
	}

	log.Info("Completed background job runner")
}

func (e *Executor) Run(daemon Daemon) error {
	log.Infof("Running %s daemon", e.Config.NodeType)

	if len(e.backgroundJobs) > 0 {
		go e.backgroundJobRunner()
	} else {
		log.Info("Skipping start of background job runner, no jobs to run")
	}

	err := daemon.Start()
	if err != nil {
		return err
	}

	daemon.Wait()
	return nil
}
