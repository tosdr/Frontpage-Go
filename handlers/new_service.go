package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"tosdrgo/internal/db"
	"tosdrgo/internal/logger"

	"github.com/gorilla/mux"
)

type Document struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	XPath string `json:"xpath"`
}

type ServiceForm struct {
	ServiceName  string
	ServiceURL   string
	WikipediaURL string
	EmailAddress string
	Notes        string
	Documents    []Document
	Errors       map[string]string
}

func NewServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	if r.Method == "GET" {
		renderNewServiceForm(w, r, lang, &ServiceForm{})
		return
	}

	if r.Method == "POST" {
		handleServiceSubmission(w, r, lang)
		return
	}

	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

func handleServiceSubmission(w http.ResponseWriter, r *http.Request, lang string) {
	logger.LogDebug("Starting service submission handling")

	// Parse documents JSON from form
	var documents []Document
	documentsJSON := r.FormValue("documents")
	if documentsJSON != "" {
		if err := json.Unmarshal([]byte(documentsJSON), &documents); err != nil {
			logger.LogError(err, "Failed to parse documents JSON")
			renderNewServiceForm(w, r, lang, &ServiceForm{
				Errors: map[string]string{"documents": "Invalid document format"},
			})
			return
		}
	}

	form := &ServiceForm{
		ServiceName:  strings.TrimSpace(r.FormValue("service_name")),
		ServiceURL:   strings.TrimSpace(r.FormValue("service_url")),
		WikipediaURL: strings.TrimSpace(r.FormValue("wikipedia_url")),
		EmailAddress: strings.TrimSpace(r.FormValue("email")),
		Notes:        strings.TrimSpace(r.FormValue("notes")),
		Documents:    documents,
		Errors:       make(map[string]string),
	}

	logger.LogDebug("Form values received - Name: %s, URL: %s, Wiki: %s, Email: %s, Documents: %v",
		form.ServiceName, form.ServiceURL, form.WikipediaURL, form.EmailAddress, form.Documents)

	// Validate form
	if !validateForm(form) {
		logger.LogDebug("Form validation failed: %v", form.Errors)
		renderNewServiceForm(w, r, lang, form)
		return
	}

	logger.LogDebug("Form validation passed, creating submission")

	// Convert documents to JSON string
	documentsBytes, err := json.Marshal(form.Documents)
	if err != nil {
		logger.LogError(err, "Failed to marshal documents")
		form.Errors["general"] = "Failed to process documents"
		renderNewServiceForm(w, r, lang, form)
		return
	}

	// Create submission
	submission := &db.ServiceSubmission{
		Name:      form.ServiceName,
		Domains:   form.ServiceURL,
		Documents: string(documentsBytes),
		Wikipedia: form.WikipediaURL,
		Email:     form.EmailAddress,
		Note:      form.Notes,
	}

	// Add submission to database
	err = db.AddSubmission(submission)
	if err != nil {
		logger.LogError(err, "Database submission failed")
		form.Errors["general"] = "Failed to submit service. Please try again later."
		renderNewServiceForm(w, r, lang, form)
		return
	}

	logger.LogDebug("Service submission successful, redirecting")
	http.Redirect(w, r, "/"+lang+"/new_service", http.StatusSeeOther)
}

func validateForm(form *ServiceForm) bool {
	isValid := true

	if len(form.ServiceName) < 2 || len(form.ServiceName) > 100 {
		form.Errors["service_name"] = "Service name must be between 2 and 100 characters"
		isValid = false
	}

	if form.ServiceURL == "" {
		form.Errors["service_url"] = "Service URL is required"
		isValid = false
	} else {
		domains := strings.Split(form.ServiceURL, ",")
		for _, domain := range domains {
			domain = strings.TrimSpace(domain)
			if domain == "" {
				continue
			}
			if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
				form.Errors["service_url"] = "Domains must not include protocols (http:// or https://)"
				isValid = false
				break
			}
			if strings.HasPrefix(domain, "www.") {
				form.Errors["service_url"] = "Domains must not include www prefix"
				isValid = false
				break
			}
			// Basic domain format validation
			if !strings.Contains(domain, ".") || len(domain) < 4 {
				form.Errors["service_url"] = "Invalid domain format"
				isValid = false
				break
			}
		}
	}

	if form.WikipediaURL != "" && !strings.HasPrefix(form.WikipediaURL, "https://en.wikipedia.org/wiki/") {
		form.Errors["wikipedia_url"] = "Wikipedia URL must be from English Wikipedia (https://en.wikipedia.org/wiki/)"
		isValid = false
	}

	if form.EmailAddress != "" { // Only validate if email is provided
		if !strings.Contains(form.EmailAddress, "@") || !strings.Contains(form.EmailAddress, ".") {
			form.Errors["email"] = "Invalid email address"
			isValid = false
		}
	}

	// Validate documents
	if len(form.Documents) == 0 {
		form.Errors["documents"] = "At least one document is required"
		isValid = false
	}

	for i, doc := range form.Documents {
		if strings.TrimSpace(doc.Name) == "" {
			form.Errors["documents"] = fmt.Sprintf("Document %d: Name is required", i+1)
			isValid = false
			break
		}

		if strings.TrimSpace(doc.URL) == "" {
			form.Errors["documents"] = fmt.Sprintf("Document %d: URL is required", i+1)
			isValid = false
			break
		}

		if !strings.HasPrefix(doc.URL, "http://") && !strings.HasPrefix(doc.URL, "https://") {
			form.Errors["documents"] = fmt.Sprintf("Document %d: URL must start with http:// or https://", i+1)
			isValid = false
			break
		}
	}

	return isValid
}

func renderNewServiceForm(w http.ResponseWriter, r *http.Request, lang string, form *ServiceForm) {
	tmpl, err := parseTemplates("templates/contents/new_service.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the new service form", err)
		return
	}

	data := struct {
		Title     string
		Beta      bool
		Lang      string
		Form      *ServiceForm
		Languages map[string]string
	}{
		Title:     "Add New Service - ToS;DR",
		Beta:      isBeta,
		Lang:      lang,
		Form:      form,
		Languages: SupportedLanguages,
	}

	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the new service form", err)
		return
	}
}