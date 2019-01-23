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
	"sync"

	"github.com/percona/mongodb-orchestration-tools/pkg/pod"
	"github.com/percona/mongodb-orchestration-tools/watchdog/config"
	"github.com/percona/mongodb-orchestration-tools/watchdog/replset"
	log "github.com/sirupsen/logrus"
)

type Manager interface {
	Close()
	Get(serviceName, rsName string) *Watcher
	HasWatcher(serviceName, rsName string) bool
	Stop(serviceName, rsName string)
	Watch(serviceName string, rs *replset.Replset)
}

type watcherState struct {
	rsName      string
	serviceName string
	watcher     *Watcher
	quit        chan bool
}

type WatcherManager struct {
	sync.Mutex
	config     *config.Config
	stop       *chan bool
	watchers   []*watcherState
	activePods *pod.Pods
}

func NewManager(config *config.Config, stop *chan bool, activePods *pod.Pods) *WatcherManager {
	return &WatcherManager{
		config:     config,
		activePods: activePods,
		stop:       stop,
		watchers:   make([]*watcherState, 0),
	}
}

func (wm *WatcherManager) HasWatcher(serviceName, rsName string) bool {
	return wm.Get(serviceName, rsName) != nil
}

func (wm *WatcherManager) Watch(serviceName string, rs *replset.Replset) {
	if wm.HasWatcher(serviceName, rs.Name) {
		return
	}

	log.WithFields(log.Fields{
		"replset": rs.Name,
		"service": serviceName,
	}).Info("Starting replset watcher")

	wm.Lock()
	defer wm.Unlock()

	quitChan := make(chan bool)
	w := &watcherState{
		rsName:      rs.Name,
		serviceName: serviceName,
		quit:        quitChan,
		watcher:     New(rs, wm.config, &quitChan, wm.activePods),
	}
	go w.watcher.Run()

	wm.watchers = append(wm.watchers, w)
}

func (wm *WatcherManager) getState(serviceName, rsName string) *watcherState {
	for _, state := range wm.watchers {
		if state.serviceName != serviceName || state.rsName != rsName {
			continue
		}
		return state
	}
	return nil
}

func (wm *WatcherManager) Get(serviceName, rsName string) *Watcher {
	wm.Lock()
	defer wm.Unlock()

	state := wm.getState(serviceName, rsName)
	if state == nil || state.watcher == nil {
		return nil
	}
	return state.watcher
}

func (wm *WatcherManager) stopWatcher(serviceName, rsName string) {
	for i, state := range wm.watchers {
		if state.serviceName != serviceName || state.rsName != rsName {
			continue
		}
		close(state.quit)
		wm.watchers = append(wm.watchers[:i], wm.watchers[i+1:]...)
	}
}

func (wm *WatcherManager) Stop(serviceName, rsName string) {
	wm.Lock()
	defer wm.Unlock()
	wm.stopWatcher(serviceName, rsName)
}

func (wm *WatcherManager) Close() {
	wm.Lock()
	defer wm.Unlock()

	for _, state := range wm.watchers {
		wm.stopWatcher(state.serviceName, state.rsName)
	}
}
