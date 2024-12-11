package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"tosdrgo/models"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	cacheKey := fmt.Sprintf("about_%s", lang)
	if cachedPage, found := pageCache.Get(cacheKey); found {
		w.Header().Set(ContentType, ContentTypeHtml)
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := parseTemplates("templates/contents/about.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the about page", err)
		return
	}

	jsonData, err := os.ReadFile("assets/about.json")
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load team data", err)
		return
	}

	var team models.Team
	if err := json.Unmarshal(jsonData, &team); err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to parse team data", err)
		return
	}

	for i := range team.Founders {
		rendered, err := RenderMarkdown([]byte(team.Founders[i].Description))
		if err != nil {
			RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render team member description", err)
			return
		}
		team.Founders[i].Description = rendered
	}

	for i := range team.Current {
		rendered, err := RenderMarkdown([]byte(team.Current[i].Description))
		if err != nil {
			RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render team member description", err)
			return
		}
		team.Current[i].Description = rendered
	}

	for i := range team.Past {
		rendered, err := RenderMarkdown([]byte(team.Past[i].Description))
		if err != nil {
			RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render team member description", err)
			return
		}
		team.Past[i].Description = rendered
	}

	mdContent, err := os.ReadFile("assets/about.md")
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load about content", err)
		return
	}

	rendered, err := RenderMarkdown(mdContent)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render about content", err)
		return
	}

	data := struct {
		Title     string
		Beta      bool
		Lang      string
		Team      models.Team
		Content   template.HTML
		Languages map[string]string
	}{
		Title:     "About Us - ToS;DR",
		Beta:      isBeta,
		Lang:      lang,
		Team:      team,
		Content:   rendered,
		Languages: SupportedLanguages,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the about page", err)
		return
	}

	pageCache.Set(cacheKey, buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(buf.Bytes())
}
