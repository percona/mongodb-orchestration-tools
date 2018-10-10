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
	"testing"

	"github.com/percona/mongodb-orchestration-tools/pkg/pod"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPkgPodK8STask(t *testing.T) {
	assert.Implements(t, (*pod.Task)(nil), &Task{})

	task := NewTask(corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name: t.Name(),
		},
		Status: corev1.PodStatus{},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Env:   []corev1.EnvVar{},
					Ports: []corev1.ContainerPort{},
				},
			},
		},
	}, "mongodb")

	assert.NotNil(t, task)
	assert.Equal(t, t.Name(), task.Name())

	// task types
	assert.True(t, task.IsTaskType(pod.TaskTypeMongod))
	assert.False(t, task.IsTaskType(pod.TaskTypeMongos))

	// test empty state
	assert.False(t, task.HasState())

	// test non-running state
	task.pod.Status.Phase = corev1.PodPending
	assert.True(t, task.HasState())
	assert.False(t, task.IsRunning())

	// test running state
	task.pod.Status.Phase = corev1.PodRunning
	assert.True(t, task.IsRunning())
	assert.Equal(t, "Running", task.State().String())

	// empty mongo addr
	_, err := task.GetMongoAddr()
	assert.Error(t, err)

	// set mongo addr
	task.pod.Spec.Containers[0].Ports = []corev1.ContainerPort{{
		Name:     "mongodb",
		HostPort: int32(27017),
	}}
	addr, err := task.GetMongoAddr()
	assert.NoError(t, err)
	assert.Equal(t, t.Name(), addr.Host)
	assert.Equal(t, 27017, addr.Port)

	// empty replset name
	_, err = task.GetMongoReplsetName()
	assert.Error(t, err)

	// set replset name
	task.pod.Spec.Containers[0].Env = []corev1.EnvVar{
		{
			Name:  "MONGODB_REPLSET",
			Value: t.Name(),
		},
	}
	rsName, err := task.GetMongoReplsetName()
	assert.NoError(t, err)
	assert.Equal(t, t.Name(), rsName)
}
