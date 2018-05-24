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

// GetPodURL returns a string representing the full HTTP URI to the 'GET /<version>/pod' API call
func (c *ClientHTTP) GetPodURL() string {
	return c.scheme.String() + c.getBaseURL() + "/" + APIVersion + "/pod"
}

// GetPods returns a slice of existing Pods in the DC/OS SDK
func (c *ClientHTTP) GetPods() (*Pods, error) {
	pods := &Pods{}
	err := c.get(c.GetPodURL(), pods)
	return pods, err
}

// GetPodTasks returns a slice of PodTask for a given DC/OS SDK Pod by name
func (c *ClientHTTP) GetPodTasks(podName string) ([]*PodTask, error) {
	podURL := c.GetPodURL() + "/" + podName + "/info"
	var tasks []*PodTask
	err := c.get(podURL, &tasks)
	return tasks, err
}
