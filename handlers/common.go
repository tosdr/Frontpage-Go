package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
	"tosdrgo/handlers/localization"
	"tosdrgo/internal/logger"

	"github.com/patrickmn/go-cache"
)

var (
	pageCache     = cache.New(4*time.Hour, 10*time.Minute)
	isBeta        bool
	baseTemplates = []string{
		"templates/layout.gohtml",
		"templates/header.gohtml",
		"templates/footer.gohtml",
	}
)

const (
	ContentType = "Content-Type"

	ContentTypeHtml = "text/html"
	ContentTypeJson = "application/json"
)

func init() {
	if err := localization.LoadTranslations("en"); err != nil {
		logger.LogError(err, "Failed to load English translations")
	}
}

func SetIsBeta(value bool) {
	isBeta = value
}

func parseTemplates(contentTemplate string, lang string, r *http.Request) (*template.Template, error) {
	templates := append([]string{}, baseTemplates...)
	templates = append(templates, contentTemplate)

	funcMap := template.FuncMap{
		"t": func(key string, args ...interface{}) string {
			translation := localization.Get(lang, key)
			if len(args) > 0 {
				return fmt.Sprintf(translation, args...)
			}
			return translation
		},
		"langURL": func(targetLang string) string {
			if r == nil {
				return "/" + targetLang
			}
			return "/" + targetLang + strings.TrimPrefix(r.URL.Path, "/"+lang)
		},
		"ToLower": strings.ToLower,
		"subtract": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"isDonationMonth": func() bool {
			currentMonth := time.Now().Month()
			return currentMonth == time.January || currentMonth == time.July
		},
	}

	tmpl := template.New("")
	tmpl.Funcs(funcMap)

	parsedTemplate, err := tmpl.ParseFiles(templates...)
	if err != nil {
		logger.LogError(err, fmt.Sprintf("Failed to parse templates for %s", contentTemplate))
		return nil, err
	}

	logger.LogDebug("Successfully parsed templates for %s", contentTemplate)
	return parsedTemplate, nil
}
