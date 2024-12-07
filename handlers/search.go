package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"tosdrgo/handlers/metrics"
	"tosdrgo/internal/db"
	"tosdrgo/models"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]
	searchTerm := vars["term"]

	cacheKey := fmt.Sprintf("search_%s_%s", lang, searchTerm)
	if cachedPage, found := pageCache.Get(cacheKey); found {
		w.Header().Set(ContentType, ContentTypeHtml)
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := parseTemplates("templates/contents/search.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the search page", err)
		return
	}

	start := time.Now()
	searchResults, code, err := db.SearchServices(searchTerm)
	if err != nil {
		RenderErrorPage(w, lang, code, "Failed to fetch search results\n"+err.Error(), err)
		return
	}

	searchDuration := time.Since(start).Seconds()
	metrics.SearchLatency.WithLabelValues(strconv.Itoa(len(searchResults))).Observe(searchDuration)

	data := struct {
		Title         string
		Beta          bool
		Lang          string
		SearchTerm    string
		SearchResults []models.SearchResult
		Languages     map[string]string
	}{
		Title:         "Search Results - ToS;DR",
		Beta:          isBeta,
		Lang:          lang,
		SearchTerm:    searchTerm,
		SearchResults: searchResults,
		Languages:     SupportedLanguages,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the search results", err)
		return
	}

	pageCache.Set(cacheKey, buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(buf.Bytes())
}
