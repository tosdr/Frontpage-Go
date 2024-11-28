package logger

import (
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
	DebugLogger *log.Logger
)

func init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)
	DebugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime)
}

// sanitizeURL removes potential PII from URLs
func sanitizeURL(url string) string {
	// Remove query parameters
	if idx := strings.Index(url, "?"); idx != -1 {
		url = url[:idx]
	}

	// Remove specific path parameters that might contain PII
	parts := strings.Split(url, "/")
	for i, part := range parts {
		if strings.Contains(part, "@") || len(part) > 20 {
			parts[i] = "[REDACTED]"
		}
	}
	return strings.Join(parts, "/")
}

// getCallerInfo returns the file and line number of the caller
func getCallerInfo() string {
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		return "unknown:0"
	}
	// Get just the file name, not the full path
	parts := strings.Split(file, "/")
	return parts[len(parts)-1] + ":" + string(rune(line))
}

// LogRequest logs incoming HTTP requests
func LogRequest(r *http.Request, duration time.Duration) {
	InfoLogger.Printf("[%s] %s %s %s (%.2fms)",
		getCallerInfo(),
		r.Method,
		sanitizeURL(r.URL.Path),
		r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")],
		float64(duration.Microseconds())/1000)
}

// LogError logs error messages with context
func LogError(err error, context string) {
	ErrorLogger.Printf("[%s] %s: %v", getCallerInfo(), context, err)
}

// LogDebug logs debug messages
func LogDebug(format string, v ...interface{}) {
	DebugLogger.Printf("[%s] "+format, append([]interface{}{getCallerInfo()}, v...)...)
}
