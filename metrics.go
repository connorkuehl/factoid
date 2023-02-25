package main

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type metrics struct {
	reqTotal                   prometheus.Counter
	rspTotal                   *prometheus.CounterVec
	inflight                   prometheus.Gauge
	totalProcessingTimeSeconds prometheus.Counter
	reqLatencySeconds          *prometheus.HistogramVec
}

func newMetrics(reg prometheus.Registerer) *metrics {
	m := &metrics{
		reqTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of requests received",
		}),
		rspTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_responses_total",
			Help: "Total number of responses sent",
		}, []string{"status"}),
		inflight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_inflight",
			Help: "Number of requests being executed right now",
		}),
		totalProcessingTimeSeconds: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "http_total_request_time_seconds",
			Help: "Time spent servicing all completed requests in seconds",
		}),
		reqLatencySeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name: "http_request_latency_seconds",
			Help: "Request latencies by status code and method",
		}, []string{"status", "method"}),
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

func (m *metrics) RequestLatency(method, status string, latency time.Duration) {
	m.reqLatencySeconds.With(prometheus.Labels{"status": status, "method": method}).
		Observe(float64(latency.Seconds()))
}
