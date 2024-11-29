package handlers

import (
	"fmt"
	"net/http"
	"tosdrgo/logger"
	"tosdrgo/metrics"
)

func RenderErrorPage(w http.ResponseWriter, lang string, errorCode int, errorMessage string, err error) {
	metrics.ErrorCounter.WithLabelValues("render_error", errorMessage).Inc()
	metrics.ErrorDetailsCounter.WithLabelValues(
		"render_error",
		fmt.Sprintf("%d", errorCode),
		errorMessage,
	).Inc()

	logger.LogError(err, fmt.Sprintf("%s (HTTP %d)", errorMessage, errorCode))

	tmpl, err := parseTemplates("templates/contents/error.gohtml", lang, nil)
	if err != nil {
		logger.LogError(err, fmt.Sprintf("Failed to load error template: %s", errorMessage))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title        string
		Beta         bool
		Lang         string
		ErrorCode    int
		ErrorMessage string
		Languages    map[string]string
	}{
		Title:        fmt.Sprintf("Error %d", errorCode),
		Beta:         isBeta,
		Lang:         lang,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		Languages:    SupportedLanguages,
	}

	w.WriteHeader(errorCode)
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		logger.LogError(err, "Failed to render error template")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
