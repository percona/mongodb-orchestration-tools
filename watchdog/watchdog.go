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
	"github.com/percona/dcos-mongo-tools/internal/pod"
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
	podSource      pod.Source
	watcherManager watcher.Manager
	quit           *chan bool
	activePods     *watcher.Pods
}

func New(config *config.Config, quit *chan bool, podSource pod.Source) *Watchdog {
	activePods := watcher.NewPods()
	return &Watchdog{
		config:         config,
		podSource:      podSource,
		watcherManager: watcher.NewManager(config, quit, activePods),
		quit:           quit,
		activePods:     activePods,
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

func (w *Watchdog) podMongodFetcher(podName string, wg *sync.WaitGroup) {
	defer wg.Done()

	log.WithFields(log.Fields{
		"pod": podName,
	}).Info("Getting tasks for pod")

	tasks, err := w.podSource.GetPodTasks(podName)
	if err != nil {
		log.WithFields(log.Fields{
			"pod":   podName,
			"error": err,
		}).Error("Error fetching DCOS pod tasks")
		return
	}
	apiFetches.With(prometheus.Labels{"type": "get_pod_tasks"}).Inc()

	for _, task := range tasks {
		if !task.IsTaskType(pod.TaskTypeMongod) {
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

		// ensure the replset has a watcher started
		if !w.watcherManager.HasWatcher(mongod.Replset) {
			rs := replset.New(w.config, mongod.Replset)
			w.watcherManager.Watch(rs)
		}

		// send the update to the watcher for the given replset
		w.watcherManager.Get(mongod.Replset).UpdateMongod(mongod)
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

func (w *Watchdog) fetchPods() {
	log.WithFields(log.Fields{
		"url": w.podSource.GetPodURL(),
	}).Info("Getting pods from url")

	pods, err := w.podSource.GetPods()
	if err != nil {
		log.WithFields(log.Fields{
			"url":   w.podSource.GetPodURL(),
			"error": err,
		}).Error("Error fetching DCOS pod list")
		return
	}
	apiFetches.With(prometheus.Labels{"type": "get_pods"}).Inc()

	if pods == nil {
		return
	}
	w.activePods.Set(pods)

	// get updated pods list
	var wg sync.WaitGroup
	for _, podName := range w.activePods.Get() {
		if w.doIgnorePod(podName) {
			continue
		}
		wg.Add(1)
		go w.podMongodFetcher(podName, &wg)
	}
	wg.Wait()
}

func (w *Watchdog) Run() {
	log.WithFields(log.Fields{
		"version":   dcosmongotools.Version,
		"framework": w.config.FrameworkName,
		"go":        runtime.Version(),
		"source":    w.podSource.Name(),
	}).Info("Starting watchdog")

	// run the prometheus metrics server
	prometheus.MustRegister(apiFetches)
	go w.runPrometheusMetricsServer()

	ticker := time.NewTicker(w.config.APIPoll)
	for {
		select {
		case <-ticker.C:
			w.fetchPods()
		case <-*w.quit:
			log.Info("Stopping watchers")
			ticker.Stop()
			return
		}
	}
}
