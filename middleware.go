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
	RequestLatency(status requestStatus, latency time.Duration)
}

func withMetrics(m goldenSignaller, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		m.TotalRequestsInc()
		m.InflightAdd(1)

		snoop := newSpyResponseWriter(w)
		next.ServeHTTP(snoop, r)

		elapsed := time.Since(start)

		statusInt := snoop.Status()
		status := strconv.Itoa(statusInt)

		var result requestStatus
		switch {
		case statusInt >= 500:
			result = requestFail
		case statusInt >= 400 && statusInt < 500:
			result = requestReject
		default:
			result = requestSuccess
		}

		m.InflightAdd(-1)
		m.TotalResponsesInc(status)
		m.RequestTimeInc(elapsed)
		m.RequestLatency(result, elapsed)
	})
}
