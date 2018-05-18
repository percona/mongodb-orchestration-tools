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

package watchdog

import (
	"runtime"
	"sync"
	"time"

	dcosMongoTools "github.com/percona/dcos-mongo-tools"
	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"github.com/percona/dcos-mongo-tools/watchdog/replset"
	"github.com/percona/dcos-mongo-tools/watchdog/replset/watcher"
	log "github.com/sirupsen/logrus"
)

type Watchdog struct {
	startTime      time.Time
	config         *config.Config
	api            api.Api
	replsetManager *replset.Manager
	watcherManager *watcher.Manager
}

func New(config *config.Config, dcosApi api.Api) *Watchdog {
	return &Watchdog{
		config:         config,
		startTime:      time.Now(),
		api:            dcosApi,
		replsetManager: replset.NewManager(config),
		watcherManager: watcher.NewManager(config),
	}
}

func (w *Watchdog) runtimeDuration() time.Duration {
	return time.Since(w.startTime)
}

func (w *Watchdog) ensureWatchers() {
	if w.runtimeDuration() < w.config.DelayWatcher {
		return
	}
	for _, rs := range w.replsetManager.GetAll() {
		w.watcherManager.Watch(rs)
	}
}

func (w *Watchdog) stopWatchers() {
	w.watcherManager.Stop()
}

func (w *Watchdog) mongodUpdater(mongodUpdates <-chan *replset.Mongod) {
	for mongod := range mongodUpdates {
		fields := log.Fields{
			"name":    mongod.Task.Name(),
			"state":   string(mongod.Task.State()),
			"replset": mongod.Replset,
			"host":    mongod.Name(),
		}
		if w.replsetManager.HasMember(mongod) {
			if mongod.Task.IsRemovedMongod() {
				log.WithFields(fields).Info("Removing completed mongod task")
				w.replsetManager.RemoveMember(mongod)
			} else {
				log.WithFields(fields).Info("Updating running mongod task")
				w.replsetManager.UpdateMember(mongod)
			}
		} else if mongod.Task.HasState() {
			log.WithFields(fields).Info("Adding new mongod task")
			w.replsetManager.UpdateMember(mongod)
		}
	}
}

func (w *Watchdog) podMongodFetcher(podName string, wg *sync.WaitGroup, updateMongod chan *replset.Mongod) {
	defer wg.Done()

	log.WithFields(log.Fields{
		"pod": podName,
	}).Info("Getting tasks for pod")
	tasks, err := w.api.GetPodTasks(podName)
	if err != nil {
		log.WithFields(log.Fields{
			"pod":   podName,
			"error": err,
		}).Error("Error fetching DCOS pod tasks")
		return
	}

	for _, task := range tasks {
		if task.IsMongodTask() != true {
			continue
		}
		mongod, err := replset.NewMongod(task, w.config.FrameworkName, podName)
		if err != nil {
			log.WithFields(log.Fields{
				"task":  task.Name(),
				"error": err,
			}).Error("Error creating mongod object")
			return
		}
		updateMongod <- mongod
	}
}

func (w *Watchdog) Run() {
	log.WithFields(log.Fields{
		"version":   dcosMongoTools.Version,
		"framework": w.config.FrameworkName,
		"go":        runtime.Version(),
	}).Info("Starting watchdog")

	// run the mongod updater in a goroutine
	updateMongod := make(chan *replset.Mongod)
	go w.mongodUpdater(updateMongod)

	ticker := time.NewTicker(w.config.APIPoll)
	for {
		select {
		case <-ticker.C:
			log.WithFields(log.Fields{
				"url": w.api.GetPodUrl(),
			}).Info("Getting pods from url")

			pods, err := w.api.GetPods()
			if err != nil {
				log.WithFields(log.Fields{
					"url":   w.api.GetPodUrl(),
					"error": err,
				}).Error("Error fetching DCOS pod list")
				continue
			}

			var wg sync.WaitGroup
			wg.Add(len(*pods))
			for _, podName := range *pods {
				go w.podMongodFetcher(podName, &wg, updateMongod)
			}
			wg.Wait()

			w.ensureWatchers()
		}
	}

	log.Info("Stopping watchers")
	w.stopWatchers()
}
