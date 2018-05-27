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

	"github.com/percona/dcos-mongo-tools/common/db"
	"github.com/percona/dcos-mongo-tools/executor/job"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

type Executor struct {
	Config         *Config
	backgroundJobs []job.BackgroundJob
	signals        chan os.Signal
	quit           chan bool
}

func New(config *Config) *Executor {
	return &Executor{
		Config:         config,
		backgroundJobs: make([]job.BackgroundJob, 0),
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
