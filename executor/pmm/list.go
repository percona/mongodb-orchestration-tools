package pmm

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

var (
	ErrListServicesFailed = errors.New("Server-side error when listing PMM services!")
	ErrMsgNoServices      = "No services under monitoring."
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
	Error         string               `json:"Err,omitempty"`
}

func (pl *ListClient) HasError() bool {
	if pl.Error != "" && strings.TrimSpace(pl.Error) != ErrMsgNoServices {
		return true
	}
	return false
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
	cmd := exec.Command("pmm-admin", "list", "--json", "--config-file="+p.configFile)
	cmd.Stderr = os.Stderr
	bytes, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	list := &ListClient{}
	err = json.Unmarshal(bytes, list)
	if err != nil {
		return nil, err
	}
	if list.HasError() {
		return nil, ErrListServicesFailed
	}

	return list, nil
}
