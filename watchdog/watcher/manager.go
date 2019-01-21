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
	Watch(rs *replset.Replset)
}

type watcherInfo struct {
	rsName      string
	serviceName string
}

type WatcherManager struct {
	sync.Mutex
	config     *config.Config
	stop       *chan bool
	quitChans  map[watcherInfo]chan bool
	watchers   map[watcherInfo]*Watcher
	activePods *pod.Pods
}

func NewManager(config *config.Config, stop *chan bool, activePods *pod.Pods) *WatcherManager {
	return &WatcherManager{
		config:     config,
		stop:       stop,
		activePods: activePods,
		quitChans:  make(map[watcherInfo]chan bool),
		watchers:   make(map[watcherInfo]*Watcher),
	}
}

func (wm *WatcherManager) HasWatcher(serviceName, rsName string) bool {
	wm.Lock()
	defer wm.Unlock()

	wInfo := watcherInfo{
		serviceName: serviceName,
		rsName:      rsName,
	}
	if _, ok := wm.watchers[wInfo]; ok {
		return true
	}
	return false
}

func (wm *WatcherManager) Watch(rs *replset.Replset) {
	if !wm.HasWatcher(rs.ServiceName, rs.Name) {
		log.WithFields(log.Fields{
			"replset": rs.Name,
		}).Info("Starting replset watcher")

		wm.Lock()

		quitChan := make(chan bool)
		wInfo := watcherInfo{
			serviceName: rs.ServiceName,
			rsName:      rs.Name,
		}
		wm.quitChans[wInfo] = quitChan
		wm.watchers[wInfo] = New(rs, wm.config, &quitChan, wm.activePods)
		go wm.watchers[wInfo].Run()

		wm.Unlock()
	}
}

func (wm *WatcherManager) Get(serviceName, rsName string) *Watcher {
	wm.Lock()
	defer wm.Unlock()

	wInfo := watcherInfo{
		serviceName: serviceName,
		rsName:      rsName,
	}
	if _, ok := wm.watchers[wInfo]; ok {
		return wm.watchers[wInfo]
	}
	return nil
}

func (wm *WatcherManager) Stop(serviceName, rsName string) {
	if wm.HasWatcher(serviceName, rsName) {
		wm.Lock()
		defer wm.Unlock()

		wInfo := watcherInfo{
			serviceName: serviceName,
			rsName:      rsName,
		}
		close(wm.quitChans[wInfo])
		delete(wm.quitChans, wInfo)
	}
}

func (wm *WatcherManager) Close() {
	for wInfo := range wm.watchers {
		wm.Stop(wInfo.serviceName, wInfo.rsName)
	}
}
