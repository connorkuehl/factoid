package main

import (
	"net/http"
	"strconv"
	"time"
)

type goldenSignaller interface {
	TotalRequestsInc()
	TotalResponsesInc(code string)
	InflightAdd(float64)
	RequestTimeInc(time.Duration)
	RequestLatency(method, status string, latency time.Duration)
}

func withMetrics(m goldenSignaller, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		m.TotalRequestsInc()
		m.InflightAdd(1)

		snoop := newSpyResponseWriter(w)
		next.ServeHTTP(snoop, r)

		elapsed := time.Since(start)

		status := strconv.Itoa(snoop.Status())

		m.InflightAdd(-1)
		m.TotalResponsesInc(status)
		m.RequestTimeInc(elapsed)
		m.RequestLatency(r.Method, status, elapsed)
	})
}
