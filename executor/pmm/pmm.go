package pmm

import (
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	pmmAdminCommand  = "pmm-admin"
	pmmAdminRunUser  = "root"
	pmmAdminRunGroup = "root"
	qanServiceName   = "mongodb:queries"
)

type PMM struct {
	config        *Config
	configFile    string
	frameworkName string
	maxRetries    uint
	retrySleep    time.Duration
	running       bool
}

func New(config *Config, frameworkName string) *PMM {
	return &PMM{
		config:        config,
		configFile:    filepath.Join(config.ConfigDir, "pmm.yml"),
		frameworkName: frameworkName,
		maxRetries:    5,
		retrySleep:    time.Duration(5) * time.Second,
	}
}

func (p *PMM) Name() string {
	return "Percona PMM"
}

func (p *PMM) DoRun() bool {
	return p.config.Enabled
}

func (p *PMM) IsRunning() bool {
	return p.running
}

func (p *PMM) DoStartQueryAnalytics() bool {
	return p.config.EnableQueryAnalytics
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

func (p *PMM) Run() error {
	if p.DoRun() == false {
		log.Warn("PMM client executor disabled! Skipping start")
		return nil
	}

	log.WithFields(log.Fields{
		"config": p.configFile,
	}).Info("Starting PMM client executor")
	p.running = true

	err := p.Repair()
	if err != nil {
		log.Errorf("Error repairing PMM services: %s", err)
		return err
	}

	err = p.StartMetrics()
	if err != nil {
		log.Errorf("PMM metrics services did not start: %s", err)
		return err
	}

	if p.DoStartQueryAnalytics() {
		err = p.StartQueryAnalytics()
		if err != nil {
			log.Errorf("Could not enable PMM Query Analytics (QAN) agent service: %s", err)
			return err
		}
	} else {
		log.Info("PMM Query Analytics (QAN) disabled, skipping")
	}

	p.running = false
	log.Info("Completed PMM client executor")

	return nil
}
