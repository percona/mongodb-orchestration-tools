package pmm

import (
	"path/filepath"
	"time"

	"github.com/mesosphere/dcos-mongo/mongodb_tools/common"
	log "github.com/sirupsen/logrus"
)

const (
	qanServiceName = "mongodb:queries"
)

type PMM struct {
	config            *Config
	configFile        string
	frameworkName     string
	connectTries      uint
	connectRetrySleep time.Duration
	maxRetries        uint
	retrySleep        time.Duration
	started           bool
}

func New(config *Config, frameworkName string) *PMM {
	return &PMM{
		config:            config,
		configFile:        filepath.Join(config.ConfigDir, "pmm.yml"),
		frameworkName:     frameworkName,
		connectTries:      10,
		connectRetrySleep: time.Duration(5) * time.Second,
		maxRetries:        5,
		retrySleep:        time.Duration(5) * time.Second,
	}
}

func (p *PMM) DoStart() bool {
	return p.config.Enabled
}

func (p *PMM) DoStartQueryAnalytics() bool {
	return p.config.EnableQueryAnalytics
}

func (p *PMM) IsStarted() bool {
	return p.started
}

func (p *PMM) StartMetrics() error {
	list, err := p.List()
	if err != nil {
		log.Error("Got error listing PMM services: %s", err)
		return err
	}

	log.WithFields(log.Fields{
		"max_retries":  p.maxRetries,
		"linux_port":   p.config.LinuxMetricsExporterPort,
		"mongodb_port": p.config.MongoDBMetricsExporterPort,
	}).Info("Starting PMM metrics services")

	services := []*Service{
		NewService(
			p.configFile,
			"linux:metrics",
			p.config.LinuxMetricsExporterPort,
			[]string{},
		),
		NewService(
			p.configFile,
			"mongodb:metrics",
			p.config.MongoDBMetricsExporterPort,
			[]string{
				"--cluster=" + p.frameworkName,
				"--uri=" + p.config.DB.Uri(),
			},
		),
	}

	for _, service := range services {
		if list != nil && list.HasService(service.Name) {
			log.Warnf("Service %s already added! Skipping", service.Name)
			continue
		}
		err := service.AddWithRetry(p.maxRetries, p.retrySleep)
		if err != nil {
			log.Errorf("Could not add PMM service %s after %d retries: %s", service.Name, p.maxRetries, err)
		}
	}

	return nil
}

func (p *PMM) StartQueryAnalytics() error {
	list, err := p.List()
	if err != nil {
		log.Errorf("Got error listing PMM services: %s", err)
		return err
	}
	if list != nil && list.HasService(qanServiceName) {
		log.Warnf("Service %s already added! Skipping", qanServiceName)
		return nil
	}

	service := NewService(
		p.configFile,
		qanServiceName,
		0,
		[]string{"--uri=" + p.config.DB.Uri()},
	)

	log.WithFields(log.Fields{
		"max_retries": p.maxRetries,
	}).Info("Starting PMM Query Analytics (QAN) agent service")

	return service.AddWithRetry(p.maxRetries, p.retrySleep)
}

func (p *PMM) Wait() {
	log.Info("Waiting to start Percona PMM client")
	time.Sleep(p.config.DelayStart)
}

func (p *PMM) WaitForServer() error {
	session, err := common.WaitForSession(
		p.config.DB,
		p.connectTries,
		p.connectRetrySleep,
	)
	if err != nil {
		return err
	}
	session.Close()
	return nil
}

func (p *PMM) Start() {
	if p.DoStart() == false {
		log.Warn("PMM client executor disabled! Skipping start")
		return
	}

	log.WithFields(log.Fields{
		"config":      p.configFile,
		"delay_start": p.config.DelayStart,
	}).Info("Starting PMM client executor")

	p.Wait()
	err := p.WaitForServer()
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("MongoDB server is now reachable")

	err = p.Repair()
	if err != nil {
		log.Errorf("Error repairing PMM services: %s", err)
		return
	}

	err = p.StartMetrics()
	if err != nil {
		log.Errorf("PMM metrics services did not start: %s", err)
		return
	}

	if p.DoStartQueryAnalytics() {
		err = p.StartQueryAnalytics()
		if err != nil {
			log.Errorf("Could not enable PMM Query Analytics (QAN) agent service: %s", err)
			return
		}
	} else {
		log.Info("PMM Query Analytics (QAN) disabled, skipping")
	}

	log.Info("Completed PMM client executor")
	p.started = true
}
