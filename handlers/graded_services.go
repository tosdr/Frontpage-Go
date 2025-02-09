package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"tosdrgo/internal/db"
	"tosdrgo/models"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func GradedServicesHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]
	grade := vars["grade"]

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if grade != "" && grade != "A" && grade != "B" && grade != "C" && grade != "D" && grade != "E" {
		http.Error(w, "Invalid grade provided", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("graded_%s_%s_%d", lang, grade, page)
	if cachedPage, found := pageCache.Get(cacheKey); found {
		w.Header().Set(ContentType, ContentTypeHtml)
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := parseTemplates("templates/contents/graded_services.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the graded services page", err)
		return
	}

	results, total, code, err := db.FetchServicesByGrade(grade, page, 24)
	if err != nil {
		RenderErrorPage(w, lang, code, "Failed to fetch results\n"+err.Error(), err)
		return
	}

	totalPages := (total + 23) / 24
	hasNext := page < totalPages
	hasPrev := page > 1

	data := struct {
		Title      string
		Beta       bool
		Lang       string
		Grade      string
		Results    []models.SearchResult
		Page       int
		TotalPages int
		HasNext    bool
		HasPrev    bool
		Languages  map[string]string
	}{
		Title:      fmt.Sprintf("Grade %s Services", grade),
		Beta:       isBeta,
		Lang:       lang,
		Grade:      grade,
		Results:    results,
		Page:       page,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
		Languages:  SupportedLanguages,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the results", err)
		return
	}

	pageCache.Set(cacheKey, buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(buf.Bytes())
}
