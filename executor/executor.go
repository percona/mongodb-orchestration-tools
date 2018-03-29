package executor

import (
	"errors"

	"github.com/mesosphere/dcos-mongo/mongodb_tools/executor/pmm"
	log "github.com/sirupsen/logrus"
)

type Executor struct {
	config *Config
	mongod *Mongod
	pmm    *pmm.PMM
}

func New(config *Config) *Executor {
	return &Executor{
		config: config,
		mongod: NewMongod(config),
		pmm:    pmm.New(config.PMM, config.FrameworkName),
	}
}

func (e *Executor) Start() error {
	var err error
	if e.pmm.DoStart() {
		go e.pmm.Start()
	} else {
		log.Info("Skipping Percona PMM client executor")
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
