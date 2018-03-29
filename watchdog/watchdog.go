package watchdog

import (
	"runtime"
	"sync"
	"time"

	"github.com/mesosphere/dcos-mongo/mongodb_tools"
	"github.com/mesosphere/dcos-mongo/mongodb_tools/common/api"
	"github.com/mesosphere/dcos-mongo/mongodb_tools/watchdog/config"
	"github.com/mesosphere/dcos-mongo/mongodb_tools/watchdog/replset"
	"github.com/mesosphere/dcos-mongo/mongodb_tools/watchdog/replset/watcher"
	log "github.com/sirupsen/logrus"
)

type Watchdog struct {
	startTime      time.Time
	config         *config.Config
	api            *api.Api
	replsetManager *replset.Manager
	watcherManager *watcher.Manager
}

func New(config *config.Config) *Watchdog {
	return &Watchdog{
		config:         config,
		startTime:      time.Now(),
		api:            api.New(config.FrameworkName, config.API),
		replsetManager: replset.NewManager(config),
		watcherManager: watcher.NewManager(config),
	}
}

func (w *Watchdog) runtimeDuration() time.Duration {
	return time.Since(w.startTime)
}

func (w *Watchdog) startWatchers() {
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
		"version":   tools.Version,
		"framework": w.config.FrameworkName,
		"go":        runtime.Version(),
	}).Info("Starting watchdog")

	// run the mongod updater in a goroutine
	updateMongod := make(chan *replset.Mongod)
	go w.mongodUpdater(updateMongod)

	for {
		log.WithFields(log.Fields{
			"url": w.api.GetPodUrl(),
		}).Info("Getting pods from url")
		pods, err := w.api.GetPods()
		if err != nil {
			log.WithFields(log.Fields{
				"url":   w.api.GetPodUrl(),
				"error": err,
			}).Error("Error fetching DCOS pod list")
			time.Sleep(w.config.APIPoll)
			continue
		}

		var wg sync.WaitGroup
		wg.Add(len(*pods))
		for _, podName := range *pods {
			go w.podMongodFetcher(podName, &wg, updateMongod)
		}
		wg.Wait()
		w.startWatchers()

		log.WithFields(log.Fields{
			"sleep": w.config.APIPoll,
		}).Info("Waiting to refresh pod info")
		time.Sleep(w.config.APIPoll)
	}

	log.Info("Stopping watchers")
	w.stopWatchers()
}
