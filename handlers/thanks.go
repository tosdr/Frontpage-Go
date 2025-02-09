package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"tosdrgo/handlers/localization"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

type Sponsor struct {
	Name        string `json:"name"`
	Logo        string `json:"logo"`
	Link        string `json:"link"`
	Description string `json:"description"`
	ClassName   string
}

func ThanksHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	cacheKey := "thanks_" + lang
	if cachedPage, found := pageCache.Get(cacheKey); found {
		w.Header().Set(ContentType, ContentTypeHtml)
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := parseTemplates("templates/contents/thanks.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the thanks page", err)
		return
	}

	jsonData, err := os.ReadFile("assets/thanks.json")
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load sponsors data", err)
		return
	}

	var sponsors []Sponsor
	if err := json.Unmarshal(jsonData, &sponsors); err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to parse sponsors data", err)
		return
	}

	for i := range sponsors {
		className := strings.ToLower(sponsors[i].Name)
		className = strings.ReplaceAll(className, " ", "")
		sponsors[i].ClassName = className
	}

	data := struct {
		Title     string
		Beta      bool
		Lang      string
		Sponsors  []Sponsor
		Languages map[string]string
	}{
		Title:     localization.Get(lang, "page.thanks"),
		Beta:      isBeta,
		Lang:      lang,
		Sponsors:  sponsors,
		Languages: SupportedLanguages,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the thanks page", err)
		return
	}

	pageCache.Set(cacheKey, buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(buf.Bytes())
}
