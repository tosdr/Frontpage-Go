package handlers

import (
	"net/http"
	"time"
	"tosdrgo/db"
	"tosdrgo/logger"
)

func HealthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	start := time.Now()

	err := db.DB.Ping()
	if err != nil {
		logger.LogError(err, "Health check failed - database connection error")
		w.Header().Set(ContentType, ContentTypeJson)
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status": "unhealthy", "message": "database connection failed"}`))
		return
	}

	w.Header().Set(ContentType, ContentTypeJson)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"status": "healthy"}`))

	logger.LogDebug("Health check completed in %.2fms", time.Since(start).Seconds()*1000)
}
