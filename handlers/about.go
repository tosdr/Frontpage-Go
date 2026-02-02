package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"tosdrgo/handlers/localization"
	"tosdrgo/internal/logger"
	"tosdrgo/models"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func renderTeamDescriptions(members []models.TeamMember) ([]models.TeamMember, error) {
	for i := range members {
		rendered, err := RenderMarkdown([]byte(members[i].Description))
		if err != nil {
			return nil, fmt.Errorf("failed to render team member description: %w", err)
		}
		members[i].Description = rendered
	}
	return members, nil
}

func AboutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	if err := localization.LoadTranslations(lang); err != nil {
		logger.LogError(err, fmt.Sprintf("Failed to load translations for %s", lang))
	}

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

	team.Founders, err = renderTeamDescriptions(team.Founders)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, err.Error(), err)
		return
	}

	team.Current, err = renderTeamDescriptions(team.Current)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, err.Error(), err)
		return
	}

	team.Past, err = renderTeamDescriptions(team.Past)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, err.Error(), err)
		return
	}

	// Try to load language-specific about content first, fall back to English
	mdPath := filepath.Join("md", lang, "about.md")
	mdContent, err := os.ReadFile(mdPath)
	if err != nil {
		// Fall back to English version
		mdPath = filepath.Join("md", "en", "about.md")
		mdContent, err = os.ReadFile(mdPath)
		if err != nil {
			// Fall back to assets/about.md as last resort
			mdContent, err = os.ReadFile("assets/about.md")
			if err != nil {
				RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load about content", err)
				return
			}
		}
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
		Title:     localization.Get(lang, "page.about"),
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
