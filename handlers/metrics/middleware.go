package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create custom response writer to capture status code
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

		// Get handler name from route
		var handler string
		if currentRoute := mux.CurrentRoute(r); currentRoute != nil {
			if name := currentRoute.GetName(); name != "" {
				handler = name
			} else {
				handler = "unknown"
			}
		} else {
			handler = "unknown"
		}

		// Call next handler
		next.ServeHTTP(rw, r)

		// Record metrics
		duration := time.Since(start).Seconds()

		// Get language from URL if present
		vars := mux.Vars(r)
		lang := vars["lang"]
		if lang == "" {
			lang = "none"
		}

		PageRenderTime.WithLabelValues(handler, lang).Observe(duration)
		RequestCounter.WithLabelValues(
			handler,
			r.Method,
			strconv.Itoa(rw.status),
		).Inc()
	})
}
