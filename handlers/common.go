package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"
	"tosdrgo/handlers/localization"
	"tosdrgo/internal/logger"
	"tosdrgo/models"

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

func GenerateOGDescription(service models.Service, lang string) string {
	ogDesc := localization.Get(lang, "service.og.rated_by")
	ogDesc = fmt.Sprintf(ogDesc, service.Name, service.Rating) + " "

	if len(service.Points) > 0 {
		goodPoints := 0
		badPoints := 0

		for _, point := range service.Points {
			if point.Case != nil {
				if point.Case.Classification == "good" {
					goodPoints++
				} else if point.Case.Classification == "bad" || point.Case.Classification == "blocker" {
					badPoints++
				}
			}
		}

		analysisPrefix := localization.Get(lang, "service.og.analysis_prefix") + " "
		goodKey := "service.og.good_point"
		if goodPoints != 1 {
			goodKey = "service.og.good_points"
		}
		badKey := "service.og.bad_point"
		if badPoints != 1 {
			badKey = "service.og.bad_points"
		}

		ogDesc += analysisPrefix
		ogDesc += fmt.Sprintf(localization.Get(lang, goodKey), goodPoints) + " and "
		ogDesc += fmt.Sprintf(localization.Get(lang, badKey), badPoints) + ". "
	} else {
		ogDesc += localization.Get(lang, "service.og.no_points") + " "
	}

	return strings.TrimSpace(ogDesc)
}

func GenerateOGImageURLBasic(title string) string {
	return fmt.Sprintf("https://tosdr.org/og/tosdr/?title=%s", title)
}

func GenerateOGImageURLService(title string, service_icon_url string, grade string, good int, bad int) string {
	return fmt.Sprintf("http://tosdr.org/og/tosdr/?title=%s&icon=%s&grade=%s&good=%d&bad=%d", title, service_icon_url, grade, good, bad)
}

func parseTemplates(contentTemplate string, lang string, r *http.Request) (*template.Template, error) {
	templates := append([]string{}, baseTemplates...)
	templates = append(templates, contentTemplate)

	var canonicalURL string
	if r != nil {
		canonicalURL = "https://tosdr.org" + r.URL.Path
	}

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
			return currentMonth == time.March ||
				currentMonth == time.June ||
				currentMonth == time.September ||
				currentMonth == time.December
		},
		"canonicalURL": func() string {
			return canonicalURL
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
