package handlers

import (
	"fmt"
	"net/http"
	"strings"
)

var SupportedLanguages = map[string]string{
	"de": "German",
	"en": "English",
	"es": "Spanish",
	"fr": "French",
	"ja": "Japanese",
	"pl": "Polish",
	"ru": "Russian",
	"zh": "Chinese",
}

func DetectLanguage(r *http.Request) string {
	acceptLang := r.Header.Get("Accept-Language")
	lang := "en" // default to English

	if acceptLang != "" {
		prefLang := strings.Split(strings.Split(acceptLang, ",")[0], "-")[0]
		if len(prefLang) == 2 && SupportedLanguages[prefLang] != "" {
			lang = strings.ToLower(prefLang)
		}
	}

	return lang
}

func DetectLanguageAndRedirect(w http.ResponseWriter, r *http.Request) {
	lang := DetectLanguage(r)
	http.Redirect(w, r, "/"+lang, http.StatusSeeOther)
}

func DetectLanguageAndRedirectWithPath(w http.ResponseWriter, r *http.Request) {
	lang := DetectLanguage(r)
	path := r.URL.Path
	http.Redirect(w, r, fmt.Sprintf("/%s%s", lang, path), http.StatusTemporaryRedirect)
}
