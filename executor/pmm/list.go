package pmm

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/percona/dcos-mongo-tools/common/command"
	log "github.com/sirupsen/logrus"
)

var (
	ErrMsgNoServices = "No services under monitoring."
)

type ListClientService struct {
	Name     string `json:"Name"`
	Running  bool   `json:"Running"`
	Type     string `json:"Type"`
	Port     string `json:"Port"`
	Password string `json:"Password,omitempty"`
	DSN      string `json:"DSN"`
	Options  string `json:"Options,omitempty"`
	SSL      string `json:"SSL,omitempty"`
}

type ListClient struct {
	ClientName    string               `json:"ClientName"`
	ServerAddress string               `json:"ServerAddress"`
	Services      []*ListClientService `json:"Services"`
	serverErr     string               `json:"Err,omitempty"`
}

func (pl *ListClient) Error() error {
	if pl.serverErr != "" {
		return fmt.Errorf("PMM error: %s\n", strings.TrimSpace(pl.serverErr))
	}
	return nil
}

func (pl *ListClient) HasError() bool {
	err := pl.Error()
	return err != nil && err.Error() != ErrMsgNoServices
}

func (pl *ListClient) HasService(serviceName string) bool {
	for _, service := range pl.Services {
		if service.Type == serviceName && service.Running {
			return true
		}
	}
	return false
}

func (p *PMM) List() (*ListClient, error) {
	log.Info("Listing PMM services")

	cmd, err := command.New(
		pmmAdminCommand,
		[]string{"list", "--json", "--config-file=" + p.configFile},
		pmmAdminRunUser,
		pmmAdminRunGroup,
	)
	if err != nil {
		return nil, err
	}

	bytes, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}

	list := &ListClient{}
	err = json.Unmarshal(bytes, list)
	if err != nil {
		return nil, err
	}
	if list.HasError() {
		return nil, list.Error()
	}

	return list, nil
}
