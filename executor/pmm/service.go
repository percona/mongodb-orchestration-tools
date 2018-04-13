package pmm

import (
	"strconv"
	"strings"
	"time"

	"github.com/percona/dcos-mongo-tools/common/command"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	ConfigFile string
	Name       string
	Port       uint
	Args       []string
}

func NewService(configFile string, name string, port uint, args []string) *Service {
	return &Service{
		ConfigFile: configFile,
		Name:       name,
		Port:       port,
		Args:       args,
	}
}

func (s *Service) Add() error {
	args := append(
		[]string{"add"},
		s.Name,
		"--config-file="+s.ConfigFile,
	)
	if int(s.Port) > 0 {
		args = append(args, "--service-port="+strconv.Itoa(int(s.Port)))
	}
	args = append(args, s.Args...)

	cmd, err := command.New(
		pmmAdminCommand,
		args,
		pmmAdminRunUser,
		pmmAdminRunGroup,
	)
	if err != nil {
		return err
	}

	out, err := cmd.CombinedOutput()
	trimmed := strings.TrimSpace(string(out))
	if err != nil {
		log.Errorf("Failed to add PMM service %s! Error: '%s'", s.Name, trimmed)
		return err
	}
	log.Infof("Added PMM service %s, pmm-admin out: '%s'", s.Name, trimmed)

	return nil
}

func (s *Service) AddWithRetry(maxRetries uint, retrySleep time.Duration) error {
	var err error
	var tries uint
	for tries <= maxRetries {
		err = s.Add()
		if err == nil {
			return nil
		}
		log.Errorf("Received error adding PMM service %s: %s, retrying", s.Name, err)
		time.Sleep(retrySleep)
		tries += 1
	}
	return err
}

func (p *PMM) Repair() error {
	log.Info("Repairing all PMM client services")

	cmd, err := command.New(
		pmmAdminCommand,
		[]string{"repair", "--config-file=" + p.configFile},
		pmmAdminRunUser,
		pmmAdminRunGroup,
	)
	if err != nil {
		return err
	}

	return cmd.Run()
}
