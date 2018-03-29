package replset

import (
	"strconv"
	"strings"

	"github.com/percona/dcos-mongo-tools/common"
	"github.com/percona/dcos-mongo-tools/common/api"
	"gopkg.in/mgo.v2"
)

const (
	backupPodNamePrefix = "backup-"
)

type Mongod struct {
	Host          string
	Port          int
	Replset       string
	FrameworkName string
	PodName       string
	Task          *api.PodTask
}

func NewMongod(task *api.PodTask, frameworkName string, podName string) (*Mongod, error) {
	var err error
	mongod := &Mongod{
		FrameworkName: frameworkName,
		PodName:       podName,
		Task:          task,
		Host:          task.GetMongoHostname(frameworkName),
	}

	mongod.Port, err = task.GetMongoPort()
	if err != nil {
		return mongod, err
	}

	mongod.Replset, err = task.GetMongoReplsetName()
	if err != nil {
		return mongod, err
	}

	return mongod, err
}

func (m *Mongod) Name() string {
	return m.Host + ":" + strconv.Itoa(m.Port)
}

func (m *Mongod) IsBackupNode() bool {
	return strings.HasPrefix(m.PodName, backupPodNamePrefix)
}

func (m *Mongod) DBConfig() *common.DBConfig {
	return &common.DBConfig{
		DialInfo: &mgo.DialInfo{
			Addrs:    []string{m.Host + ":" + strconv.Itoa(m.Port)},
			Direct:   true,
			FailFast: true,
			Timeout:  common.DefaultMongoDBTimeoutDuration,
		},
	}
}
