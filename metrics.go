package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/connorkuehl/factoid/promlabels"
)

type metrics struct {
	reqTotal                   prometheus.Counter
	rspTotal                   *prometheus.CounterVec
	inflight                   prometheus.Gauge
	totalProcessingTimeSeconds prometheus.Counter
	reqLatencySeconds          *prometheus.HistogramVec
}

func newMetrics(reg prometheus.Registerer) *metrics {
	namespace := "factoid"

	m := &metrics{
		reqTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_requests_total",
			Help:      "Total number of requests received",
		}),
		rspTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_responses_total",
			Help:      "Total number of responses sent",
		}, []string{"status"}),
		inflight: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "http_requests_inflight",
			Help:      "Number of requests being executed right now",
		}),
		totalProcessingTimeSeconds: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_total_request_time_seconds",
			Help:      "Time spent servicing all completed requests in seconds",
		}),
		reqLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "http_request_latency_seconds",
			Help:      "Request latencies by outcome",
		}, []string{"status"}),
	}

	reg.MustRegister(m.reqTotal)
	reg.MustRegister(m.rspTotal)
	reg.MustRegister(m.inflight)
	reg.MustRegister(m.totalProcessingTimeSeconds)
	reg.MustRegister(m.reqLatencySeconds)

	return m
}

func (m *metrics) TotalRequestsInc() {
	m.reqTotal.Inc()
}

func (m *metrics) TotalResponsesInc(status string) {
	m.rspTotal.With(prometheus.Labels{"status": status}).Inc()
}

func (m *metrics) InflightAdd(inc float64) {
	m.inflight.Add(inc)
}

func (m *metrics) RequestTimeInc(inc time.Duration) {
	m.totalProcessingTimeSeconds.Add(float64(inc.Seconds()))
}

func (m *metrics) RequestLatency(status promlabels.RequestStatus, latency time.Duration) {
	m.reqLatencySeconds.With(prometheus.Labels{"status": string(status)}).
		Observe(float64(latency.Seconds()))
}
