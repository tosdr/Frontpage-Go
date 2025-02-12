package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"tosdrgo/handlers/localization"

	"github.com/patrickmn/go-cache"

	"github.com/gorilla/mux"
)

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
		err := r.ParseMultipartForm(32 << 20)
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		// Get form data using r.MultipartForm
		category := r.MultipartForm.Value["category"][0]
		name := r.MultipartForm.Value["name"][0]
		email := r.MultipartForm.Value["email"][0]
		company := r.MultipartForm.Value["company"][0]
		message := r.MultipartForm.Value["message"][0]

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
