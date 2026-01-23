package handlers

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

func CacheControlMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		ext := strings.ToLower(filepath.Ext(path))

		switch ext {
		case ".css", ".js", ".svg", ".woff", ".woff2", ".ttf", ".eot", ".otf":
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			w.Header().Set("Expires", time.Now().Add(365*24*time.Hour).Format(http.TimeFormat))
		case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".avif", ".ico":
			w.Header().Set("Cache-Control", "public, max-age=604800, immutable")
			w.Header().Set("Expires", time.Now().Add(7*24*time.Hour).Format(http.TimeFormat))
		default:
			w.Header().Set("Cache-Control", "public, max-age=3600")
		}

		next.ServeHTTP(w, r)
	})
}

func NoCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", time.Now().Add(-1*time.Hour).Format(http.TimeFormat))
		next.ServeHTTP(w, r)
	})
}
