package httpapi

import (
	"net/http"
	"strconv"
	"time"
)

func (s *Server) instrument(method, route string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		s.metrics.requests.WithLabelValues(method, route, strconv.Itoa(rw.code)).Inc()
		s.metrics.duration.WithLabelValues(method, route).Observe(duration.Seconds())
		s.logger.Info("request completed",
			"method", method,
			"route", route,
			"status_code", rw.code,
			"duration_ms", duration.Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}
