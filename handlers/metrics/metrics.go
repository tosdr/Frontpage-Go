package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// RequestCounter Request metrics
	RequestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tosdr_http_requests_total",
			Help: "Total number of HTTP requests by handler, method, and status code",
		},
		[]string{"handler", "method", "status"},
	)

	// PageRenderTime Page render timing
	PageRenderTime = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tosdr_page_render_duration_seconds",
			Help:    "Time taken to render pages",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "lang"},
	)

	// ErrorCounter Error tracking
	ErrorCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tosdr_errors_total",
			Help: "Total number of errors by type and message",
		},
		[]string{"type", "message"},
	)

	// ErrorDetailsCounter Detailed error tracking
	ErrorDetailsCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tosdr_error_details_total",
			Help: "Detailed error information including status codes",
		},
		[]string{"type", "status_code", "message"},
	)

	// CacheHits Cache performance
	CacheHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tosdr_cache_hits_total",
			Help: "Total number of cache hits by cache type",
		},
		[]string{"cache_type"},
	)

	CacheMisses = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tosdr_cache_misses_total",
			Help: "Total number of cache misses by cache type",
		},
		[]string{"cache_type"},
	)

	// SearchLatency Search latency
	SearchLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "tosdr_search_duration_seconds",
			Help:    "Time taken to perform searches",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		},
		[]string{"result_count"},
	)

	// RateLimitExceeded Rate limit metrics
	RateLimitExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "tosdr_rate_limit_exceeded_total",
			Help: "Total number of rate limit exceeded events",
		},
		[]string{"handler"},
	)
)
