package middleware

import (
	"net/http"
	"strconv"
	"time"

	"dZev1/character-gallery/internal/metrics"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		metrics.RequestsInFlight.WithLabelValues(req.Method).Inc()
		defer metrics.RequestsInFlight.WithLabelValues(req.Method).Dec()

		start := time.Now()
		sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sr, req)
		elapsed := time.Since(start).Seconds()

		path := req.Pattern
		if path == "" {
			path = req.URL.Path
		}

		metrics.RequestsTotal.WithLabelValues(req.Method, path, strconv.Itoa(sr.status)).Inc()
		metrics.RequestDuration.WithLabelValues(req.Method, path).Observe(elapsed)
	})
}