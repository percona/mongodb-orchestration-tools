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

package dcos

import (
	"errors"
	"strconv"
	"strings"

	"github.com/percona/dcos-mongo-tools/internal"
	"github.com/percona/dcos-mongo-tools/watchdog/pod"
)

type DCOSTaskState string

var (
	DCOSAutoIPDnsSuffix   string        = "autoip.dcos.thisdcos.directory"
	DCOSTaskStateError    DCOSTaskState = "TASK_ERROR"
	DCOSTaskStateFailed   DCOSTaskState = "TASK_FAILED"
	DCOSTaskStateFinished DCOSTaskState = "TASK_FINISHED"
	DCOSTaskStateKilled   DCOSTaskState = "TASK_KILLED"
	DCOSTaskStateLost     DCOSTaskState = "TASK_LOST"
	DCOSTaskStateRunning  DCOSTaskState = "TASK_RUNNING"
	DCOSTaskStateUnknown  DCOSTaskState = "UNKNOWN"
)

type DCOSTask struct {
	frameworkName string
	data          *DCOSTaskData
}

type DCOSTaskData struct {
	Info   *DCOSTaskInfo   `json:"info"`
	Status *DCOSTaskStatus `json:"status"`
}

type DCOSTaskCommandEnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DCOSTaskCommandEnvironment struct {
	Variables []*DCOSTaskCommandEnvironmentVariable `json:"variables"`
}

type DCOSTaskCommand struct {
	Environment *DCOSTaskCommandEnvironment `json:"environment"`
	Value       string                      `json:"value"`
}

type DCOSTaskInfo struct {
	Name    string           `json:"name"`
	Command *DCOSTaskCommand `json:"command"`
}

type DCOSTaskStatus struct {
	State *DCOSTaskState `json:"state"`
}

func (task *DCOSTask) Name() string {
	return task.data.Info.Name
}

func (task *DCOSTask) Framework() string {
	if task.frameworkName == "" {
		task.frameworkName = internal.DefaultFrameworkName
	}
	return task.frameworkName
}

func (task *DCOSTask) SetFramework(name string) {
	task.frameworkName = name
}

func (task *DCOSTask) HasState() bool {
	return task.data.Status != nil && task.data.Status.State != nil
}

func (task *DCOSTask) State() DCOSTaskState {
	if task.HasState() {
		return *task.data.Status.State
	}
	return DCOSTaskStateUnknown
}

func (task *DCOSTask) IsRunning() bool {
	return task.State() == DCOSTaskStateRunning
}

func (task *DCOSTask) IsTaskType(taskType pod.TaskType) bool {
	if task.data.Info != nil && strings.HasSuffix(task.data.Info.Name, "-"+taskType.String()) {
		return strings.Contains(task.data.Info.Command.Value, "mongodb-executor-")
	}
	return false
}

func (task *DCOSTask) getEnvVar(variableName string) (string, error) {
	if task.data.Info.Command != nil && task.data.Info.Command.Environment != nil {
		for _, variable := range task.data.Info.Command.Environment.Variables {
			if variable.Name == variableName {
				return variable.Value, nil
			}
		}
	}
	return "", errors.New("Could not find env variable: " + variableName)
}

func (task *DCOSTask) GetMongoAddr() (*pod.MongoAddr, error) {
	addr := &pod.MongoAddr{
		Host: task.data.Info.Name + "." + task.Framework() + "." + DCOSAutoIPDnsSuffix,
	}
	portStr, err := task.getEnvVar(internal.EnvMongoDBPort)
	if err != nil {
		return addr, err
	}
	addr.Port, err = strconv.Atoi(portStr)
	return addr, err
}

func (task *DCOSTask) GetMongoReplsetName() (string, error) {
	return task.getEnvVar(internal.EnvMongoDBReplset)
}
