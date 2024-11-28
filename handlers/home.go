package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"
	"tosdrgo/db"
	"tosdrgo/localization"
	"tosdrgo/logger"
	"tosdrgo/models"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	vars := mux.Vars(r)
	lang := vars["lang"]

	if err := localization.LoadTranslations(lang); err != nil {
		logger.LogError(err, fmt.Sprintf("Failed to load translations for %s", lang))
	}

	cacheKey := fmt.Sprintf("home_%s", lang)
	if cachedPage, found := pageCache.Get(cacheKey); found {
		w.Header().Set(ContentType, ContentTypeHtml)
		_, _ = w.Write(cachedPage.([]byte))
		logger.LogDebug("Served cached home page for language %s in %.2fms", lang, time.Since(start).Seconds()*1000)
		return
	}

	tmpl, err := parseTemplates("templates/contents/home.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the home page", err)
		return
	}

	featured, err := db.FetchFeaturedServicesData()
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to fetch featured services", err)
		return
	}

	classificationContent, err := os.ReadFile(fmt.Sprintf("md/%s/classification.md", lang))
	if err != nil {
		logger.LogError(err, fmt.Sprintf("Failed to load classification content for %s, falling back to English", lang))
		classificationContent, err = os.ReadFile("md/en/classification.md")
	}

	rendered, err := RenderMarkdown(classificationContent)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render classification content", err)
		return
	}

	data := struct {
		Title           string
		Beta            bool
		Lang            string
		LastFetchedTime string
		Featured        []models.FeaturedService
		Classification  template.HTML
		Languages       map[string]string
	}{
		Title:           "Home Page",
		Beta:            isBeta,
		Lang:            lang,
		LastFetchedTime: time.Now().Format(time.RFC850),
		Featured:        featured.Services,
		Classification:  rendered,
		Languages:       SupportedLanguages,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the home page", err)
		return
	}

	pageCache.Set(cacheKey, buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(buf.Bytes())

	logger.LogDebug("Rendered home page for language %s in %.2fms", lang, time.Since(start).Seconds()*1000)
}
