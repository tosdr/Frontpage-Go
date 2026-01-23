package handlers

import (
	"fmt"
	"net/http"
	"time"
	"tosdrgo/internal/logger"

	"github.com/gorilla/mux"
)

type SitemapURL struct {
	Loc        string
	LastMod    string
	ChangeFreq string
	Priority   string
}

func SitemapHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	// Default to English if no language specified
	if lang == "" {
		lang = "en"
	}

	baseURL := "https://tosdr.org"
	now := time.Now().Format("2006-01-02")

	// Static pages with their priorities and change frequencies
	urls := []SitemapURL{
		{Loc: baseURL + "/" + lang, LastMod: now, ChangeFreq: "daily", Priority: "1.0"},
		{Loc: baseURL + "/" + lang + "/about", LastMod: now, ChangeFreq: "monthly", Priority: "0.8"},
		{Loc: baseURL + "/" + lang + "/contact", LastMod: now, ChangeFreq: "monthly", Priority: "0.7"},
		{Loc: baseURL + "/" + lang + "/donate", LastMod: now, ChangeFreq: "monthly", Priority: "0.8"},
		{Loc: baseURL + "/" + lang + "/thanks", LastMod: now, ChangeFreq: "monthly", Priority: "0.6"},
		{Loc: baseURL + "/" + lang + "/services/A", LastMod: now, ChangeFreq: "daily", Priority: "0.9"},
		{Loc: baseURL + "/" + lang + "/services/B", LastMod: now, ChangeFreq: "daily", Priority: "0.9"},
		{Loc: baseURL + "/" + lang + "/services/C", LastMod: now, ChangeFreq: "daily", Priority: "0.9"},
		{Loc: baseURL + "/" + lang + "/services/D", LastMod: now, ChangeFreq: "daily", Priority: "0.9"},
		{Loc: baseURL + "/" + lang + "/services/E", LastMod: now, ChangeFreq: "daily", Priority: "0.9"},
	}

	// Build XML manually to avoid template escaping issues
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">`

	for _, url := range urls {
		xml += fmt.Sprintf(`
  <url>
    <loc>%s</loc>
    <lastmod>%s</lastmod>
    <changefreq>%s</changefreq>
    <priority>%s</priority>
  </url>`, url.Loc, url.LastMod, url.ChangeFreq, url.Priority)
	}

	xml += `
</urlset>`

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")

	_, err := w.Write([]byte(xml))
	if err != nil {
		logger.LogError(err, "Failed to write sitemap")
	}
}

func RobotsHandler(w http.ResponseWriter, r *http.Request) {
	robotsTxt := `User-agent: *
Allow: /
Disallow: /api/
Disallow: /dashboard/
Disallow: /login
Disallow: /logout
Disallow: /auth/

Sitemap: https://tosdr.org/sitemap.xml`

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Cache-Control", "public, max-age=86400")
	_, _ = w.Write([]byte(robotsTxt))
}
