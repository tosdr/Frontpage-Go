package handlers

import (
	"net/http"
	"strings"
)

var SupportedLanguages = map[string]string{
	"en": "English",
	"de": "German",
	"es": "Spanish",
	"fr": "French",
	"hi": "Hindi",
	"id": "Indonesian",
	"ja": "Japanese",
	"pl": "Polish",
	"pt": "Portuguese",
	"ru": "Russian",
	"zh": "Chinese",
}

func DetectLanguageAndRedirect(w http.ResponseWriter, r *http.Request) {
	acceptLang := r.Header.Get("Accept-Language")
	lang := "en" // default to English

	if acceptLang != "" {
		// Take the first language preference
		prefLang := strings.Split(strings.Split(acceptLang, ",")[0], "-")[0]
		// You might want to add validation here to ensure the language is supported
		if len(prefLang) == 2 {
			lang = strings.ToLower(prefLang)
		}
	}

	http.Redirect(w, r, "/"+lang, http.StatusSeeOther)
}
