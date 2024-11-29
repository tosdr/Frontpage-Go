package middleware

import (
	"net/http"
	"tosdrgo/metrics"
	"tosdrgo/ratelimit"
)

func RateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.Header.Get("X-Real-IP")
		if ip == "" {
			ip = r.RemoteAddr
		}

		println("The IP is: ", ip)

		if !ratelimit.Limiter.Allow(ip) {
			metrics.RateLimitExceeded.WithLabelValues("search").Inc()
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
