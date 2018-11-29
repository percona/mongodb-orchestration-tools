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
	"os"
	"testing"

	"github.com/percona/mongodb-orchestration-tools/pkg"
	"github.com/percona/mongodb-orchestration-tools/pkg/pod"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInternalPodK8SPods(t *testing.T) {
	assert.Implements(t, (*pod.Source)(nil), &Pods{})

	p := NewPods(pkg.DefaultServiceName, DefaultNamespace)
	assert.NotNil(t, p)

	pods, err := p.Pods()
	assert.NoError(t, err)
	assert.Len(t, pods, 0)

	corev1Pod := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: t.Name(),
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Env: []corev1.EnvVar{
							{
								Name:  pkg.EnvMongoDBPort,
								Value: t.Name(),
							},
						},
						Ports: []corev1.ContainerPort{
							{
								Name:     "mongodb",
								HostIP:   "1.2.3.4",
								HostPort: int32(27017),
							},
						},
					},
				},
			},
		},
	}
	p.Update(corev1Pod, nil)
	pods, _ = p.Pods()
	assert.Len(t, pods, 1)
	assert.Equal(t, t.Name(), pods[0])

	// test Succeeded pod is not listed by .Pods()
	corev1Pod[0].Status.Phase = corev1.PodSucceeded
	p.Update(corev1Pod, nil)
	pods, _ = p.Pods()
	assert.Len(t, pods, 0)

	assert.Equal(t, "k8s", p.Name())

	assert.Equal(t, "", p.URL())
	os.Setenv(EnvKubernetesHost, t.Name())
	os.Setenv(EnvKubernetesPort, "443")
	defer os.Unsetenv(EnvKubernetesHost)
	defer os.Unsetenv(EnvKubernetesPort)
	assert.Equal(t, "tcp://"+t.Name()+":443", p.URL())
}
