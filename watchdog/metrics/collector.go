package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = "watchdog"

type Collector struct {
	PodSourceErrorsTotal *prometheus.CounterVec
	PodSourceGetsTotal   *prometheus.CounterVec
}

func NewCollector() *Collector {
	return &Collector{
		PodSourceErrorsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "pod_source",
			Name:      "errors_total",
			Help:      "The total number of errors from polling the watchdog pod source",
		}, []string{"source"}),
		PodSourceGetsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "pod_source",
			Name:      "gets_total",
			Help:      "The total number of times the watchdog has polled a pod source",
		}, []string{"source"}),
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.PodSourceErrorsTotal.Collect(ch)
	c.PodSourceGetsTotal.Collect(ch)
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	c.PodSourceErrorsTotal.Describe(ch)
	c.PodSourceGetsTotal.Describe(ch)
}
