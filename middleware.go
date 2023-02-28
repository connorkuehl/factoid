package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/connorkuehl/factoid/internal/promlabels"
)

type goldenSignaller interface {
	TotalRequestsInc()
	TotalResponsesInc(code string)
	InflightAdd(float64)
	RequestTimeInc(time.Duration)
	RequestLatency(status promlabels.RequestStatus, latency time.Duration)
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

		var result promlabels.RequestStatus
		switch {
		case statusInt >= 500:
			result = promlabels.RequestFail
		case statusInt >= 400 && statusInt < 500:
			result = promlabels.RequestReject
		default:
			result = promlabels.RequestSuccess
		}

		m.InflightAdd(-1)
		m.TotalResponsesInc(status)
		m.RequestTimeInc(elapsed)
		m.RequestLatency(result, elapsed)
	})
}
