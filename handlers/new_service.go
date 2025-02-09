package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"tosdrgo/handlers/localization"
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

	// check if already exists (domain in db)
	existing, err := db.GetServiceSubmissionByDomain(form.ServiceURL)
	if err != nil {
		logger.LogError(err, "Failed to check if service already exists")
	}

	if existing != 0 {
		form.Errors["service_url"] = "Service already in submission queue!"
		renderNewServiceForm(w, r, lang, form)
		return
	}

	urls := strings.Split(form.ServiceURL, ",")
	for _, url := range urls {
		url = strings.TrimSpace(url)
	}

	trimmedUrls := strings.Join(urls, ",")
	logger.LogDebug("Form validation passed, creating submission")

	// Create submission
	submission := &db.ServiceSubmission{
		Name:      form.ServiceName,
		Domains:   trimmedUrls,
		Documents: documentsJSON,
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

	// Validate each field using separate functions
	if !validateServiceName(form) {
		isValid = false
	}
	if !validateServiceURL(form) {
		isValid = false
	}
	if !validateWikipediaURL(form) {
		isValid = false
	}
	if !validateEmail(form) {
		isValid = false
	}
	if !validateDocuments(form) {
		isValid = false
	}

	return isValid
}

func validateServiceName(form *ServiceForm) bool {
	if len(form.ServiceName) < 2 || len(form.ServiceName) > 100 {
		form.Errors["service_name"] = "Service name must be between 2 and 100 characters"
		return false
	}
	return true
}

func validateServiceURL(form *ServiceForm) bool {
	if form.ServiceURL == "" {
		form.Errors["service_url"] = "Service URL is required"
		return false
	}

	domains := strings.Split(form.ServiceURL, ",")
	for _, domain := range domains {
		domain = strings.TrimSpace(domain)
		if domain == "" {
			continue
		}
		if err := validateDomain(domain); err != nil {
			form.Errors["service_url"] = err.Error()
			return false
		}
	}
	return true
}

func validateDomain(domain string) error {
	if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
		return fmt.Errorf("domains must not include protocols (http:// or https://)")
	}
	if strings.HasPrefix(domain, "www.") {
		return fmt.Errorf("domains must not include www prefix")
	}
	if !strings.Contains(domain, ".") || len(domain) < 4 {
		return fmt.Errorf("invalid domain format")
	}
	return nil
}

func validateWikipediaURL(form *ServiceForm) bool {
	if form.WikipediaURL != "" && !strings.HasPrefix(form.WikipediaURL, "https://en.wikipedia.org/wiki/") {
		form.Errors["wikipedia_url"] = "Wikipedia URL must be from English Wikipedia (https://en.wikipedia.org/wiki/)"
		return false
	}
	return true
}

func validateEmail(form *ServiceForm) bool {
	if form.EmailAddress != "" && (!strings.Contains(form.EmailAddress, "@") || !strings.Contains(form.EmailAddress, ".")) {
		form.Errors["email"] = "Invalid email address"
		return false
	}
	return true
}

func validateDocuments(form *ServiceForm) bool {
	if len(form.Documents) == 0 {
		form.Errors["documents"] = "At least one document is required"
		return false
	}

	for i, doc := range form.Documents {
		if err := validateDocument(i, doc); err != nil {
			form.Errors["documents"] = err.Error()
			return false
		}
	}
	return true
}

func validateDocument(index int, doc Document) error {
	if strings.TrimSpace(doc.Name) == "" {
		return fmt.Errorf("document %d: Name is required", index+1)
	}
	if strings.TrimSpace(doc.URL) == "" {
		return fmt.Errorf("document %d: URL is required", index+1)
	}
	if !strings.HasPrefix(doc.URL, "http://") && !strings.HasPrefix(doc.URL, "https://") {
		return fmt.Errorf("document %d: URL must start with http:// or https://", index+1)
	}
	return nil
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
		Title:     localization.Get(lang, "page.newservice"),
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
