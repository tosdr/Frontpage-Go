package handlers

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"tosdrgo/handlers/localization"
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

	if serviceID == "" {
		RenderErrorPage(w, lang, http.StatusBadRequest, "Service ID is required", nil)
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

	ogTitle := localization.Get(lang, "service.og.title")
	ogTitle = fmt.Sprintf(ogTitle, service.Name, service.Rating)

	goodPoints := 0
	badPoints := 0

	for _, point := range service.Points {
		if point.Case != nil {
			if point.Case.Classification == "good" {
				goodPoints++
			} else if point.Case.Classification == "bad" || point.Case.Classification == "blocker" {
				badPoints++
			}
		}
	}

	data := struct {
		Title     string
		Beta      bool
		Lang      string
		Service   models.Service
		Languages map[string]string
		Canonical string
		OGTitle   string
		OGDesc    string
		OGImage   string
		OGType    string
	}{
		Title:     service.Name + " - ToS;DR",
		Beta:      isBeta,
		Lang:      lang,
		Service:   *service,
		Languages: SupportedLanguages,
		Canonical: fmt.Sprintf("https://tosdr.org/%s/service/%d", lang, intServiceID),
		OGTitle:   ogTitle,
		OGDesc:    GenerateOGDescription(*service, lang),
		OGImage:   GenerateOGImageURLService(service.Name, service.Image, service.Rating, goodPoints, badPoints),
		OGType:    "website",
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
