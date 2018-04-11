package executor

import (
	"time"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
	log "github.com/sirupsen/logrus"
	"gopkg.in/mgo.v2"
)

type ExecutorJob interface {
	Name() string
	DoRun() bool
	IsRunning() bool
	Run()
	Close()
}

type Executor struct {
	Config  *Config
	Mongod  *Mongod
	PMM     *pmm.PMM
	Metrics *metrics.Metrics
	jobs    []ExecutorJob
}

func New(config *Config) *Executor {
	return &Executor{
		Config:  config,
		Mongod:  NewMongod(config),
		PMM:     pmm.New(config.PMM, config.FrameworkName),
		Metrics: metrics.New(config.Metrics),
		jobs:    make([]ExecutorJob, 0),
	}
}

func (e *Executor) WaitForSession() (*mgo.Session, error) {
	return common.WaitForSession(
		e.Config.DB,
		e.Config.ConnectTries,
		e.Config.ConnectRetrySleep,
	)
}

func (e *Executor) AddJob(job ExecutorJob) {
	log.Debugf("Adding background job %s\n", job.Name())
	e.jobs = append(e.jobs, job)
}

func (e *Executor) RunMongod() error {
	log.Info("Running mongod daemon")

	// Percona PMM
	if e.PMM.DoRun() {
		e.AddJob(e.PMM)
	} else {
		log.Info("Skipping Percona PMM client executor")
	}

	// DC/OS Metrics
	if e.Metrics.DoRun() {
		e.AddJob(e.Metrics)
	} else {
		log.Info("Skipping DC/OS Metrics client executor")
	}

	err := e.Mongod.Start()
	if err != nil {
		return err
	}

	if len(e.jobs) > 0 {
		log.Info("Waiting for MongoDB to become reachable")
		session, err := e.WaitForSession()
		if err != nil {
			log.Errorf("Could not get connection to mongodb: %s\n", err)
			return err
		}
		log.Info("Mongodb is now reachable")
		session.Close()

		log.Info("Starting background job runner")
		go e.BackgroundJobRunner()
	} else {
		log.Info("Skipping start of background job runner, no jobs to run")
	}

	e.Mongod.Wait()
	return nil
}

func (e *Executor) BackgroundJobRunner() {
	log.WithFields(log.Fields{
		"delay": e.Config.DelayBackgroundJob,
	}).Info("Delaying the start of the background job runner")
	time.Sleep(e.Config.DelayBackgroundJob)

	for _, job := range e.jobs {
		log.Infof("Starting job %s\n", job.Name())
		job.Run()
		job.Close()
	}
}
