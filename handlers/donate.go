package handlers

import (
	"bytes"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

func RedirectDonate(w http.ResponseWriter, r *http.Request) {
	lang := DetectLanguage(r)
	http.Redirect(w, r, "/"+lang+"/donate", http.StatusSeeOther)
}

func DonateHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	cacheKey := "donate_" + lang
	if cachedPage, found := pageCache.Get(cacheKey); found {
		w.Header().Set(ContentType, ContentTypeHtml)
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := parseTemplates("templates/contents/donate.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the donate page", err)
		return
	}

	data := struct {
		Title     string
		Beta      bool
		Lang      string
		Languages map[string]string
	}{
		Title:     "Donate",
		Beta:      isBeta,
		Lang:      lang,
		Languages: SupportedLanguages,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the donate page", err)
		return
	}

	pageCache.Set(cacheKey, buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(buf.Bytes())
}
