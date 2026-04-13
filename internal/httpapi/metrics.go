package httpapi

import "github.com/prometheus/client_golang/prometheus"

type metrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func newMetrics(registry *prometheus.Registry) *metrics {
	requests := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "job_search",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Total number of HTTP requests.",
		},
		[]string{"method", "route", "code"},
	)
	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "job_search",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "HTTP request duration in seconds.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	registry.MustRegister(requests, duration)
	return &metrics{requests: requests, duration: duration}
}
