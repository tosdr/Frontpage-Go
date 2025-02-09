package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"time"
	"tosdrgo/handlers/localization"
	"tosdrgo/internal/db"
	"tosdrgo/internal/logger"
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

	featured, err := db.FetchFeaturedServicesData(lang)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to fetch featured services", err)
		return
	}

	data := struct {
		Title           string
		Beta            bool
		Lang            string
		LastFetchedTime string
		Featured        []models.FeaturedService
		Languages       map[string]string
	}{
		Title:           localization.Get(lang, "page.home"),
		Beta:            isBeta,
		Lang:            lang,
		LastFetchedTime: time.Now().Format(time.RFC850),
		Featured:        featured.Services,
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
