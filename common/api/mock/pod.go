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

package mock

import (
	"errors"
	"path/filepath"

	"github.com/percona/dcos-mongo-tools/common/api"
)

func (a *API) GetPodUrl() string {
	return "http://localhost/" + api.APIVersion + "/pod"
}

func (a *API) GetPods() (*api.Pods, error) {
	if SimulateError {
		return nil, errors.New("simulating a .GetPods() error")
	}
	pods := &api.Pods{}
	err := loadJSONFile(apiFilePath(filepath.Join("pod.json")), pods)
	return pods, err
}

func (a *API) GetPodTasks(podName string) ([]*api.PodTask, error) {
	if SimulateError {
		return nil, errors.New("simulating a .GetPodTasks() error")
	}

	var tasks []*api.PodTask
	err := loadJSONFile(apiFilePath(filepath.Join("pod", podName+".info.json")), &tasks)
	return tasks, err
}
