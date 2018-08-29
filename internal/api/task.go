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

	"github.com/percona/dcos-mongo-tools/internal"
)

type PodTaskState string

var (
	AutoIPDnsSuffix                   = "autoip.dcos.thisdcos.directory"
	PodTaskStateError    PodTaskState = "TASK_ERROR"
	PodTaskStateFailed   PodTaskState = "TASK_FAILED"
	PodTaskStateFinished PodTaskState = "TASK_FINISHED"
	PodTaskStateKilled   PodTaskState = "TASK_KILLED"
	PodTaskStateLost     PodTaskState = "TASK_LOST"
	PodTaskStateRunning  PodTaskState = "TASK_RUNNING"
	PodTaskStateUnknown  PodTaskState = "UNKNOWN"
)

type PodTask interface {
	GetEnvVar(variableName string) (string, error)
	GetMongoHostname(frameworkName string) string
	GetMongoPort() (int, error)
	GetMongoReplsetName() (string, error)
	HasState() bool
	IsMongodTask() bool
	IsMongosTask() bool
	IsRemovedMongod() bool
	IsRunning() bool
	Name() string
	State() PodTaskState
}

type PodTaskHTTP struct {
	Info   *PodTaskInfo   `json:"info"`
	Status *PodTaskStatus `json:"status"`
}

func (task *PodTaskHTTP) Name() string {
	return task.Info.Name
}

func (task *PodTaskHTTP) HasState() bool {
	return task.Status != nil && task.Status.State != nil
}

func (task *PodTaskHTTP) State() PodTaskState {
	if task.HasState() {
		return *task.Status.State
	}
	return PodTaskStateUnknown
}

func (task *PodTaskHTTP) IsRunning() bool {
	return task.State() == PodTaskStateRunning
}

func (task *PodTaskHTTP) IsMongodTask() bool {
	if task.Info != nil && strings.HasSuffix(task.Info.Name, "-mongod") {
		return strings.Contains(task.Info.Command.Value, "mongodb-executor-")
	}
	return false
}

func (task *PodTaskHTTP) IsMongosTask() bool {
	if task.Info != nil && strings.HasSuffix(task.Info.Name, "-mongos") {
		return strings.Contains(task.Info.Command.Value, "mongodb-executor-")
	}
	return false
}

// Asking for a better way to detect a removed task here: https://github.com/mesosphere/dcos-mongo/issues/112
// for now we will use the lack of a task state to determine a task is intentionally removed (for scale-down, etc)
func (task *PodTaskHTTP) IsRemovedMongod() bool {
	return task.IsMongodTask() && task.HasState() == false
}

func (task *PodTaskHTTP) GetMongoHostname(frameworkName string) string {
	return task.Info.Name + "." + frameworkName + "." + AutoIPDnsSuffix
}

func (task *PodTaskHTTP) GetEnvVar(variableName string) (string, error) {
	if task.Info.Command != nil && task.Info.Command.Environment != nil {
		for _, variable := range task.Info.Command.Environment.Variables {
			if variable.Name == variableName {
				return variable.Value, nil
			}
		}
	}
	return "", errors.New("Could not find env variable: " + variableName)
}

func (task *PodTaskHTTP) GetMongoPort() (int, error) {
	portStr, err := task.GetEnvVar(common.EnvMongoDBPort)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(portStr)
}

func (task *PodTaskHTTP) GetMongoReplsetName() (string, error) {
	return task.GetEnvVar(common.EnvMongoDBReplset)
}
