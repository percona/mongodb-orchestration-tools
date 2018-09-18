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

package watcher

import (
	"testing"

	"github.com/percona/dcos-mongo-tools/internal/api"
	"github.com/stretchr/testify/assert"
)

func TestWatchdogWatcherPods(t *testing.T) {
	pods := NewPods()
	pods.Set(&api.Pods{"test"})

	assert.Len(t, pods.Get(), 1)
	assert.Equal(t, "test", pods.Get()[0])

	assert.True(t, pods.Has("test"))
	pods.Set(&api.Pods{"false"})
	assert.False(t, pods.Has("test"))
}
