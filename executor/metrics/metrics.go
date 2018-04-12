package metrics

import (
	"github.com/percona/dcos-mongo-tools/common"
	log "github.com/sirupsen/logrus"
)

const (
	mgoStatsdRunUser              = "root"
	mgoStatsdRunGroup             = "root"
	mgoStatsdConfigUpdateInterval = "0"
)

type Metrics struct {
	config  *Config
	running bool
}

func New(config *Config) *Metrics {
	return &Metrics{
		config: config,
	}
}

func (m *Metrics) Name() string {
	return "DC/OS Metrics"
}

func (m *Metrics) DoRun() bool {
	return m.config.Enabled
}

func (m *Metrics) IsRunning() bool {
	return m.running
}

func (m *Metrics) Run() error {
	if m.DoRun() == false {
		log.Warn("DC/OS Metrics client executor disabled! Skipping start")
		return nil
	}

	cmd, err := common.NewCommand(
		m.config.MgoStatsdBin,
		[]string{
			"-configUpdateInterval", mgoStatsdConfigUpdateInterval,
			"-config", m.config.MgoStatsdConfigFile,
		},
		mgoStatsdRunUser,
		mgoStatsdRunGroup,
	)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"binary": m.config.MgoStatsdBin,
		"config": m.config.MgoStatsdConfigFile,
	}).Info("Starting DC/OS Metrics client executor")

	m.running = true
	err = cmd.Start()
	if err != nil {
		m.running = false
		return err
	}

	log.Info("Completed DC/OS Metrics client executor")

	return nil
}
