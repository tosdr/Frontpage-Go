package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"tosdrgo/internal/db"
	"tosdrgo/models"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func ServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]
	serviceID := vars["serviceID"]

	serviceID = regexp.MustCompile("[^0-9]").ReplaceAllString(serviceID, "")

	cacheKey := fmt.Sprintf("service_%s_%s", lang, serviceID)
	if cachedPage, found := pageCache.Get(cacheKey); found {
		w.Header().Set(ContentType, ContentTypeHtml)
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := parseTemplates("templates/contents/service.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the service page", err)
		return
	}

	intServiceID, err := strconv.Atoi(serviceID)
	if err != nil {
		log.Printf("Error parsing service ID: %v", err)
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to parse service ID", err)
		return
	}

	service, err := db.FetchServiceData(intServiceID, lang)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusNotFound, "Service not found", err)
		return
	}

	data := struct {
		Title     string
		Beta      bool
		Lang      string
		Service   models.Service
		Languages map[string]string
	}{
		Title:     service.Name + " - ToS;DR",
		Beta:      isBeta,
		Lang:      lang,
		Service:   *service,
		Languages: SupportedLanguages,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the service page", err)
		return
	}

	pageCache.Set(cacheKey, buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(buf.Bytes())
}
