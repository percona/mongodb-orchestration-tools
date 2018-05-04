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

type Pods []string

type PodTask interface {
	Name() string
	HasState() bool
	State() PodTaskState
	IsRunning() bool
	IsMongodTask() bool
	IsMongosTask() bool
	IsRemovedMongod() bool
	GetMongoHostname(frameworkName string) string
	GetEnvVar(variableName string) (string, error)
	GetMongoPort() (int, error)
	GetMongoReplsetName() (string, error)
}

type PodTaskCommandEnvironmentVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type PodTaskCommandEnvironment struct {
	Variables []*PodTaskCommandEnvironmentVariable `json:"variables"`
}

type PodTaskCommand struct {
	Environment *PodTaskCommandEnvironment `json:"environment"`
	Value       string                     `json:"value"`
}

type PodTaskInfo struct {
	Name    string          `json:"name"`
	Command *PodTaskCommand `json:"command"`
}

type PodTaskStatus struct {
	State *PodTaskState `json:"state"`
}
