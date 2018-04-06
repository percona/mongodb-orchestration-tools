package executor

import (
	"errors"

	"github.com/percona/dcos-mongo-tools/executor/metrics"
	"github.com/percona/dcos-mongo-tools/executor/pmm"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	config  *Config
	mongod  *Mongod
	pmm     *pmm.PMM
	metrics *metrics.Metrics
}

func New(config *Config) *Executor {
	return &Executor{
		config:  config,
		mongod:  NewMongod(config),
		pmm:     pmm.New(config.PMM, config.FrameworkName),
		metrics: metrics.New(config.Metrics),
	}
}

func (e *Executor) Start() error {
	var err error

	// Percona PMM
	if e.pmm.DoStart() {
		go e.pmm.Start()
	} else {
		log.Info("Skipping Percona PMM client executor")
	}

	// DC/OS Metrics
	if e.metrics.DoStart() {
		go e.metrics.Start()
	} else {
		log.Info("Skipping DC/OS Metrics client executore")
	}

	switch e.config.NodeType {
	case NodeTypeMongod:
		if e.mongod == nil {
			errors.New("unsupported mongod config!")
		}
		err = e.mongod.Start()
		if err != nil {
			return err
		}
		e.mongod.Wait()
		return nil
	case NodeTypeMongos:
		return errors.New("mongos is not supported yet!")
	default:
		return errors.New("did not start anything, this is unexpected")
	}
}
