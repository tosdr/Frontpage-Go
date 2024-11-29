package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	PageRenderTime = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "tosdr_page_render_time_seconds",
		Help:    "Time taken to render pages",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"handler", "lang"})

	RequestCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tosdr_requests_total",
		Help: "Total number of HTTP requests",
	}, []string{"handler", "method", "status"})

	ErrorCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tosdr_errors_total",
		Help: "Total number of errors",
	}, []string{"handler", "error_type"})

	ErrorDetailsCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tosdr_error_details_total",
		Help: "Detailed breakdown of errors by status code and message",
	}, []string{"handler", "status_code", "error_message"})

	CacheHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tosdr_cache_hits_total",
		Help: "Total number of cache hits",
	}, []string{"cache_type"})

	CacheMisses = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "tosdr_cache_misses_total",
		Help: "Total number of cache misses",
	}, []string{"cache_type"})
)
