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
	InfoLogger   *log.Logger
	ErrorLogger  *log.Logger
	DebugLogger  *log.Logger
	WarnLogger   *log.Logger
	currentLevel LogLevel
)

type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

var logLevelMap = map[string]LogLevel{
	"DEBUG": DebugLevel,
	"INFO":  InfoLevel,
	"WARN":  WarnLevel,
	"ERROR": ErrorLevel,
}

var ignoredErrorPatterns = []string{
	"Invalid service ID in shield handler",
	"The requested page was not found",
	"Failed to fetch search results",
	"search term must be at least 3 characters long",
}

func init() {
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	ErrorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)
	DebugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime)
	WarnLogger = log.New(os.Stdout, "WARN: ", log.Ldate|log.Ltime)

	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "WARN"
	}

	if lvl, ok := logLevelMap[strings.ToUpper(level)]; ok {
		currentLevel = lvl
	} else {
		currentLevel = WarnLevel
	}
}

func shouldLogError(err error, context string) bool {
	if err == nil && context == "" {
		return false
	}

	message := context
	if err != nil {
		message += err.Error()
	}

	for _, pattern := range ignoredErrorPatterns {
		if strings.Contains(message, pattern) {
			return false
		}
	}
	return true
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
	if currentLevel <= InfoLevel {
		InfoLogger.Printf("[%s] %s %s %s (%.2fms)",
			getCallerInfo(),
			r.Method,
			sanitizeURL(r.URL.Path),
			r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")],
			float64(duration.Microseconds())/1000)
	}
}

// LogError logs error messages with context
func LogError(err error, context string) {
	if currentLevel <= ErrorLevel && shouldLogError(err, context) {
		ErrorLogger.Printf("[%s] %s: %v", getCallerInfo(), context, err)
	}
}

// LogDebug logs debug messages
func LogDebug(format string, v ...interface{}) {
	if currentLevel <= DebugLevel {
		DebugLogger.Printf("[%s] "+format, append([]interface{}{getCallerInfo()}, v...)...)
	}
}

// LogWarn logs warning messages
func LogWarn(format string, v ...interface{}) {
	if currentLevel <= WarnLevel {
		WarnLogger.Printf("[%s] "+format, append([]interface{}{getCallerInfo()}, v...)...)
	}
}
