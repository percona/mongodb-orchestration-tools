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
	"github.com/percona/dcos-mongo-tools/internal/db"
	"github.com/percona/dcos-mongo-tools/internal/pod"
)

type TaskState string

var (
	AutoIPDNSSuffix   string    = "autoip.dcos.thisdcos.directory"
	TaskStateError    TaskState = "TASK_ERROR"
	TaskStateFailed   TaskState = "TASK_FAILED"
	TaskStateFinished TaskState = "TASK_FINISHED"
	TaskStateKilled   TaskState = "TASK_KILLED"
	TaskStateLost     TaskState = "TASK_LOST"
	TaskStateRunning  TaskState = "TASK_RUNNING"
	TaskStateUnknown  TaskState = "UNKNOWN"
)

func (s TaskState) String() string {
	return string(s)
}

type Task struct {
	FrameworkName string
	Data          *TaskData
}

type TaskData struct {
	Info   *TaskInfo   `json:"info"`
	Status *TaskStatus `json:"status"`
}

type TaskCommandEnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type TaskCommandEnvironment struct {
	Variables []*TaskCommandEnvironmentVariable `json:"variables"`
}

type TaskCommand struct {
	Environment *TaskCommandEnvironment `json:"environment"`
	Value       string                  `json:"value"`
}

type TaskInfo struct {
	Name    string       `json:"name"`
	Command *TaskCommand `json:"command"`
}

type TaskStatus struct {
	State *TaskState `json:"state"`
}

func NewTask(data *TaskData, frameworkName string) *Task {
	return &Task{Data: data, FrameworkName: frameworkName}
}

func (task *Task) getEnvVar(variableName string) (string, error) {
	if task.Data.Info.Command != nil && task.Data.Info.Command.Environment != nil {
		for _, variable := range task.Data.Info.Command.Environment.Variables {
			if variable.Name == variableName {
				return variable.Value, nil
			}
		}
	}
	return "", errors.New("Could not find env variable: " + variableName)
}

func (task *Task) Name() string {
	return task.Data.Info.Name
}

func (task *Task) HasState() bool {
	return task.Data.Status != nil && task.Data.Status.State != nil
}

func (task *Task) State() pod.TaskState {
	if task.HasState() {
		return task.Data.Status.State
	}
	return &TaskStateUnknown
}

func (task *Task) IsRunning() bool {
	return task.State() == TaskStateRunning
}

func (task *Task) IsTaskType(taskType pod.TaskType) bool {
	if task.Data.Info != nil && strings.HasSuffix(task.Data.Info.Name, "-"+taskType.String()) {
		return strings.Contains(task.Data.Info.Command.Value, "mongodb-executor-")
	}
	return false
}

func (task *Task) GetMongoAddr() (*db.Addr, error) {
	addr := &db.Addr{
		Host: task.Data.Info.Name + "." + task.FrameworkName + "." + AutoIPDNSSuffix,
	}
	portStr, err := task.getEnvVar(internal.EnvMongoDBPort)
	if err != nil {
		return addr, err
	}
	addr.Port, err = strconv.Atoi(portStr)
	return addr, err
}

func (task *Task) GetMongoReplsetName() (string, error) {
	return task.getEnvVar(internal.EnvMongoDBReplset)
}
