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

	dcosmongotools "github.com/percona/dcos-mongo-tools"
	"github.com/percona/dcos-mongo-tools/common/api"
	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"github.com/percona/dcos-mongo-tools/watchdog/replset"
	"github.com/percona/dcos-mongo-tools/watchdog/watcher"
	log "github.com/sirupsen/logrus"
)

type Watchdog struct {
	config         *config.Config
	api            api.Client
	watcherManager *watcher.Manager
	quit           *chan bool
}

func New(config *config.Config, quit *chan bool, client api.Client) *Watchdog {
	return &Watchdog{
		config:         config,
		api:            client,
		watcherManager: watcher.NewManager(config, quit),
		quit:           quit,
	}
}

func (w *Watchdog) mongodUpdateHandler(mongodUpdates <-chan *replset.Mongod) {
	for mongodUpdate := range mongodUpdates {
		// ensure the replset has a watcher started
		if !w.watcherManager.HasWatcher(mongodUpdate.Replset) {
			rs := replset.New(w.config, mongodUpdate.Replset)
			w.watcherManager.Watch(rs)
		}

		// send the update to the watcher for the given replset
		w.watcherManager.Get(mongodUpdate.Replset).UpdateMongod(mongodUpdate)
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

func (w *Watchdog) doSkipPod(podName string) bool {
	for _, skipPodName := range w.config.IgnorePods {
		if podName == skipPodName {
			return true
		}
	}
	return false
}

func (w *Watchdog) fetchPods(mongodUpdates chan *replset.Mongod) {
	log.WithFields(log.Fields{
		"url": w.api.GetPodURL(),
	}).Info("Getting pods from url")

	pods, err := w.api.GetPods()
	if err != nil {
		log.WithFields(log.Fields{
			"url":   w.api.GetPodURL(),
			"error": err,
		}).Error("Error fetching DCOS pod list")
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(*pods))
	for _, podName := range *pods {
		if w.doSkipPod(podName) {
			continue
		}
		go w.podMongodFetcher(podName, &wg, mongodUpdates)
	}
	wg.Wait()
}

func (w *Watchdog) Run() {
	log.WithFields(log.Fields{
		"version":   dcosmongotools.Version,
		"framework": w.config.FrameworkName,
		"go":        runtime.Version(),
	}).Info("Starting watchdog")

	// run the mongod update hander in a goroutine to receive updates
	mongodUpdates := make(chan *replset.Mongod)
	go w.mongodUpdateHandler(mongodUpdates)

	ticker := time.NewTicker(w.config.APIPoll)
	for {
		select {
		case <-ticker.C:
			w.fetchPods(mongodUpdates)
		case <-*w.quit:
			log.Info("Stopping watchers")
			ticker.Stop()
			break
		}
	}
}
