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
	"net/http"
	"runtime"
	"sync"
	"time"

	dcosmongotools "github.com/percona/dcos-mongo-tools"
	"github.com/percona/dcos-mongo-tools/internal/api"
	"github.com/percona/dcos-mongo-tools/watchdog/config"
	"github.com/percona/dcos-mongo-tools/watchdog/replset"
	"github.com/percona/dcos-mongo-tools/watchdog/watcher"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

const (
	metricsPath        = "/metrics"
	DefaultMetricsPort = "8080"
)

var (
	apiFetches = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "api",
		Name:      "fetches_total",
		Help:      "API fetches",
	}, []string{"type"})
)

type Watchdog struct {
	config         *config.Config
	api            api.Client
	watcherManager watcher.Manager
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

func (w *Watchdog) runPrometheusMetricsServer() {
	log.WithFields(log.Fields{
		"port": w.config.MetricsPort,
		"path": metricsPath,
	}).Info("Starting Prometheus metrics server")
	http.Handle(metricsPath, promhttp.Handler())
	log.Fatal(http.ListenAndServe(":"+w.config.MetricsPort, nil))
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
	apiFetches.With(prometheus.Labels{"type": "get_pod_tasks"}).Inc()

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

func (w *Watchdog) doIgnorePod(podName string) bool {
	for _, ignorePodName := range w.config.IgnorePods {
		if podName == ignorePodName {
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
	apiFetches.With(prometheus.Labels{"type": "get_pods"}).Inc()

	var wg sync.WaitGroup
	for _, podName := range *pods {
		if w.doIgnorePod(podName) {
			continue
		}
		wg.Add(1)
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

	// run the prometheus metrics server
	prometheus.MustRegister(apiFetches)
	go w.runPrometheusMetricsServer()

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
			return
		}
	}
}
