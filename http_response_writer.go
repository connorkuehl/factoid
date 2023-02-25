package main

import "net/http"

type spyResponseWriter struct {
	interior   http.ResponseWriter
	statusCode int
	written    bool
}

func newSpyResponseWriter(w http.ResponseWriter) *spyResponseWriter {
	return &spyResponseWriter{interior: w}
}

func (w *spyResponseWriter) Status() int {
	return w.statusCode
}

func (w *spyResponseWriter) Header() http.Header {
	return w.interior.Header()
}

func (w *spyResponseWriter) WriteHeader(status int) {
	w.interior.WriteHeader(status)

	if !w.written {
		w.statusCode = status
		w.written = true
	}
}

func (w *spyResponseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}

	return w.interior.Write(b)
}

func (w *spyResponseWriter) Unwrap() http.ResponseWriter {
	return w.interior
}
