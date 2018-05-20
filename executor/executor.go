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
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
	mgostatsd "github.com/scullxbones/mgo-statsd"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

// BackgroundJob is an interface for background backgroundJobs to be executed against the Daemon
type BackgroundJob interface {
	Name() string
	DoRun() bool
	IsRunning() bool
	Run(quit *chan bool) error
}

// Daemon is an interface for the mongodb (mongod or mongos) daemon
type Daemon interface {
	IsStarted() bool
	Start() error
	Wait()
	Kill() error
}

type Executor struct {
	Config         *Config
	backgroundJobs []BackgroundJob
	signals        chan os.Signal
	quit           chan bool
}

func New(config *Config) *Executor {
	return &Executor{
		Config:         config,
		backgroundJobs: make([]BackgroundJob, 0),
		signals:        make(chan os.Signal),
		quit:           make(chan bool),
	}
}

func (e *Executor) waitForSession() (*mgo.Session, error) {
	return db.WaitForSession(
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
		e.addBackgroundJob(metrics.New(e.Config.Metrics, session.Copy(), metrics.NewStatsdPusher(statsdCnf)))
	} else {
		log.Info("Skipping DC/OS Metrics client executor")
	}

	// Percona PMM
	if e.Config.PMM.Enabled {
		e.addBackgroundJob(pmm.New(e.Config.PMM, e.Config.FrameworkName))
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

func (e *Executor) Run(daemon Daemon) error {
	go e.backgroundJobRunner()

	log.Infof("Running %s daemon", e.Config.NodeType)
	err := daemon.Start()
	if err != nil {
		daemon.Kill()
		return err
	}

	signal.Notify(e.signals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-e.signals
	log.Infof("Received %s signal, killing %s daemon and jobs", sig, e.Config.NodeType)

	e.quit <- true
	return daemon.Kill()
}
