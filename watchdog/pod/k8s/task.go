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

package k8s

import (
	"github.com/percona/dcos-mongo-tools/watchdog/pod"
)

type K8STask struct{}

func (task *K8STask) Name() string {
	return "test"
}

func (task *K8STask) IsRunning() bool {
	return false
}

func (task *K8STask) IsTaskType(taskType pod.TaskType) bool {
	return false
}

func (task *K8STask) GetMongoAddr() (*pod.MongoAddr, error) {
	addr := &pod.MongoAddr{}
	return addr, nil
}

func (task *K8STask) GetMongoReplsetName() (string, error) {
	return "rs", nil
}
