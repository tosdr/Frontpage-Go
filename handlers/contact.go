package handlers

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"strings"
	"tosdrgo/handlers/localization"
	"tosdrgo/handlers/metrics"
	"tosdrgo/handlers/ratelimit"

	"github.com/patrickmn/go-cache"

	"github.com/gorilla/mux"
)

const (
	maxContactFieldLen   = 2000
	maxContactMessageLen = 5000
)

func clientIP(r *http.Request) string {
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func firstFormValue(form *multipart.Form, key string) string {
	if form == nil {
		return ""
	}
	if vals, ok := form.Value[key]; ok && len(vals) > 0 {
		return vals[0]
	}
	return ""
}

type ContactCategory struct {
	ID          string
	Title       string
	Description string
	Email       string
}

var webhookUrl = ""

func InitContact(webhook string) {
	webhookUrl = webhook
}

// ContactHandler handles the contact page
func ContactHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	if r.Method == "POST" {
		// Per-IP rate limit to stop automated submission floods.
		if !ratelimit.ContactLimiter.Allow(clientIP(r)) {
			metrics.RateLimitExceeded.WithLabelValues("contact").Inc()
			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// Cap the in-memory form size; oversized requests are rejected.
		err := r.ParseMultipartForm(1 << 20)
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		if strings.TrimSpace(firstFormValue(r.MultipartForm, "contact_hp")) != "" {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"status": "success"}`))
			return
		}

		category := strings.TrimSpace(firstFormValue(r.MultipartForm, "category"))
		name := strings.TrimSpace(firstFormValue(r.MultipartForm, "name"))
		email := strings.TrimSpace(firstFormValue(r.MultipartForm, "email"))
		company := strings.TrimSpace(firstFormValue(r.MultipartForm, "company"))
		message := strings.TrimSpace(firstFormValue(r.MultipartForm, "message"))

		if category == "" || name == "" || email == "" || message == "" {
			http.Error(w, "All fields are required.", http.StatusBadRequest)
			return
		}
		if len(name) > maxContactFieldLen || len(email) > maxContactFieldLen ||
			len(category) > maxContactFieldLen || len(company) > maxContactFieldLen ||
			len(message) > maxContactMessageLen {
			http.Error(w, "Submission too large.", http.StatusBadRequest)
			return
		}

		// Prepare Mattermost webhook payload
		companyText := ""
		if company != "" {
			companyText = "**Company:** " + company + "\n"
		}

		payload := map[string]interface{}{
			"text": "### New Contact Form Submission\n" +
				"**Message:**\n" + message + "\n\n" +
				"**Category:** " + category + "\n" +
				"**Name:** " + name + "\n" +
				"**Email:** " + email + "\n" +
				companyText,
			"username": "Form",
		}

		jsonData, err := json.Marshal(payload)
		if err != nil {
			println(err.Error())
			http.Error(w, "Failed to create webhook payload", http.StatusInternalServerError)
			return
		}

		if webhookUrl != "" {
			resp, err := http.Post(webhookUrl, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				println(err.Error())
				http.Error(w, "Failed to send webhook", http.StatusInternalServerError)
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
				http.Error(w, "Webhook returned error", http.StatusInternalServerError)
				return
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status": "success"}`))
		return
	}

	cacheKey := "contact_" + lang
	if cachedPage, found := pageCache.Get(cacheKey); found {
		w.Header().Set(ContentType, ContentTypeHtml)
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := parseTemplates("templates/contents/contact.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the contact page", err)
		return
	}

	data := struct {
		Title     string
		Beta      bool
		Lang      string
		Languages map[string]string
	}{
		Title:     localization.Get(lang, "page.contact"),
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
