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

package pod

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

//func TestInternalAPIGetPodURL(t *testing.T) {
//	assert.Equal(t, testAPI.GetPodURL(), testAPI.scheme.String()+testAPI.config.Host+"/"+APIVersion+"/pod", "api.GetPodURL() is incorrect")
//}

func TestInternalPodHasPod(t *testing.T) {
	pods := Pods{"test1"}
	assert.True(t, pods.HasPod("test1"))
	assert.False(t, pods.HasPod("not here"))
}

func TestInternalPodActivePods(t *testing.T) {
	activePods := NewActivePods()
	assert.Len(t, *activePods.Get(), 0)
	activePods.Set(&Pods{"test"})

	pods := *activePods.Get()
	assert.Len(t, pods, 1)
	assert.Equal(t, "test", pods[0])

	assert.True(t, activePods.Has("test"))
	activePods.Set(&Pods{"false"})
	assert.False(t, activePods.Has("test"))
}
