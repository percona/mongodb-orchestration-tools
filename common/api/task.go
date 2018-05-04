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

package api

import (
	"errors"
	"strconv"
	"strings"

	"github.com/percona/dcos-mongo-tools/common"
)

type PodTaskState string

var (
	autoIpDnsSuffix                   = "autoip.dcos.thisdcos.directory"
	PodTaskStateError    PodTaskState = "TASK_ERROR"
	PodTaskStateFailed   PodTaskState = "TASK_FAILED"
	PodTaskStateFinished PodTaskState = "TASK_FINISHED"
	PodTaskStateKilled   PodTaskState = "TASK_KILLED"
	PodTaskStateLost     PodTaskState = "TASK_LOST"
	PodTaskStateRunning  PodTaskState = "TASK_RUNNING"
	PodTaskStateUnknown  PodTaskState = "UNKNOWN"
)

type PodTaskHttp struct {
	Info   *PodTaskInfo   `json:"info"`
	Status *PodTaskStatus `json:"status"`
}

func (task *PodTaskHttp) Name() string {
	return task.Info.Name
}

func (task *PodTaskHttp) HasState() bool {
	return task.Status != nil && task.Status.State != nil
}

func (task *PodTaskHttp) State() PodTaskState {
	if task.HasState() {
		return *task.Status.State
	}
	return PodTaskStateUnknown
}

func (task *PodTaskHttp) IsRunning() bool {
	return task.State() == PodTaskStateRunning
}

func (task *PodTaskHttp) IsMongodTask() bool {
	if strings.HasSuffix(task.Info.Name, "-mongod") {
		return strings.Contains(task.Info.Command.Value, "mongodb-executor-")
	}
	return false
}

func (task *PodTaskHttp) IsMongosTask() bool {
	if strings.HasSuffix(task.Info.Name, "-mongos") {
		return strings.Contains(task.Info.Command.Value, "mongodb-executor-")
	}
	return false
}

// Asking for a better way to detect a removed task here: https://github.com/mesosphere/dcos-mongo/issues/112
// for now we will use the lack of a task state to determine a task is intentionally removed (for scale-down, etc)
func (task *PodTaskHttp) IsRemovedMongod() bool {
	return task.IsMongodTask() && task.HasState() == false
}

func (task *PodTaskHttp) GetMongoHostname(frameworkName string) string {
	return task.Info.Name + "." + frameworkName + "." + autoIpDnsSuffix
}

func (task *PodTaskHttp) GetEnvVar(variableName string) (string, error) {
	if task.Info.Command != nil && task.Info.Command.Environment != nil {
		for _, variable := range task.Info.Command.Environment.Variables {
			if variable.Name == variableName {
				return variable.Value, nil
			}
		}
	}
	return "", errors.New("Could not find env variable: " + variableName)
}

func (task *PodTaskHttp) GetMongoPort() (int, error) {
	portStr, err := task.GetEnvVar(common.EnvMongoDBPort)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(portStr)
}

func (task *PodTaskHttp) GetMongoReplsetName() (string, error) {
	return task.GetEnvVar(common.EnvMongoDBReplset)
}
