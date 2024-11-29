package handlers

import (
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

func DetectLanguageAndRedirect(w http.ResponseWriter, r *http.Request) {
	acceptLang := r.Header.Get("Accept-Language")
	lang := "en" // default to English

	if acceptLang != "" {
		// Take the first language preference
		prefLang := strings.Split(strings.Split(acceptLang, ",")[0], "-")[0]
		if len(prefLang) == 2 && SupportedLanguages[prefLang] != "" {
			lang = strings.ToLower(prefLang)
		}
	}

	http.Redirect(w, r, "/"+lang, http.StatusSeeOther)
}
