package metrics

import (
	log "github.com/sirupsen/logrus"
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

func (m *Metrics) Close() {
	return
}

func (m *Metrics) DoRun() bool {
	return m.config.Enabled
}

func (m *Metrics) IsRunning() bool {
	return m.running
}

func (m *Metrics) Run() {
	if m.DoRun() == false {
		log.Warn("DC/OS Metrics client executor disabled! Skipping start")
		return
	}

	log.Info("Starting DC/OS Metrics client executor")
	m.running = true

	log.Warn("Do something here!")

	m.running = false
	log.Info("Completed DC/OS Metrics client executor")
}
