// Copyright 2018 Percona LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package replset

import (
	"strconv"
	"strings"

	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/common/db"
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

func (m *Mongod) DBConfig() *db.Config {
	return &db.Config{
		DialInfo: &mgo.DialInfo{
			Addrs:    []string{m.Host + ":" + strconv.Itoa(m.Port)},
			Direct:   true,
			FailFast: true,
			Timeout:  db.DefaultMongoDBTimeoutDuration,
		},
	}
}
