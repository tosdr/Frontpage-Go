package handlers

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/yuin/goldmark"
	meta "github.com/yuin/goldmark-meta"
	"github.com/yuin/goldmark/parser"
	"html/template"
	"net/http"
	"os"
	"tosdrgo/handlers/localization"
	"tosdrgo/internal/logger"
)

func SiteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	site := vars["sitename"]
	lang := vars["lang"]

	if err := localization.LoadTranslations(lang); err != nil {
		logger.LogError(err, fmt.Sprintf("Failed to load translations for %s", lang))
	}

	cacheKey := fmt.Sprintf("view_%s_%s", lang, site)
	if cachedPage, found := pageCache.Get(cacheKey); found {
		w.Header().Set(ContentType, ContentTypeHtml)
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := parseTemplates("templates/contents/markdown.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the page", err)
		return
	}

	content, err := os.ReadFile(fmt.Sprintf("md/%s/%s.md", lang, site))
	if err != nil {
		content, err = os.ReadFile(fmt.Sprintf("md/en/%s.md", site))
		if err != nil {
			if os.IsNotExist(err) {
				RenderErrorPage(w, lang, http.StatusNotFound, "The requested page was not found", err)
			} else {
				RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the page content", err)
			}
			return
		}
	}

	markdown := goldmark.New(
		goldmark.WithExtensions(
			meta.Meta,
		),
	)

	context := parser.NewContext()
	var buf bytes.Buffer

	if err := markdown.Convert(content, &buf, parser.WithContext(context)); err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to parse markdown", err)
		return
	}

	metaData := meta.Get(context)
	title := ""
	if metaData != nil {
		if t, ok := metaData["Title"].(string); ok {
			title = t
		}
	}

	data := struct {
		Title     string
		Beta      bool
		Lang      string
		Content   template.HTML
		Languages map[string]string
	}{
		Title:     title,
		Beta:      isBeta,
		Lang:      lang,
		Content:   template.HTML(buf.String()),
		Languages: SupportedLanguages,
	}

	var renderBuf bytes.Buffer
	err = tmpl.ExecuteTemplate(&renderBuf, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the page", err)
		return
	}

	pageCache.Set(cacheKey, renderBuf.Bytes(), cache.DefaultExpiration)

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(renderBuf.Bytes())
}
