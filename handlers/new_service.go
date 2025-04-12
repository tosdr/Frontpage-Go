package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"tosdrgo/handlers/localization"
	"tosdrgo/internal/db"
	"tosdrgo/internal/logger"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"gorm.io/gorm"
)

var store = sessions.NewCookieStore([]byte("tosdr-secret-key"))

type ServiceRequest struct {
	ID        uint   `gorm:"primarykey"`
	Name      string `gorm:"column:name"`
	Domains   string `gorm:"column:domains"`
	Wikipedia string `gorm:"column:wikipedia"`
	Email     string `gorm:"column:email"`
	Note      string `gorm:"column:note"`
	Count     int    `gorm:"column:count;default:1"`
}

func (ServiceRequest) TableName() string {
	return "service_requests_new"
}

type ServiceForm struct {
	ServiceName  string
	ServiceURL   string
	WikipediaURL string
	EmailAddress string
	Note         string
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

	form := &ServiceForm{
		ServiceName:  strings.TrimSpace(r.FormValue("service_name")),
		ServiceURL:   strings.TrimSpace(r.FormValue("service_url")),
		WikipediaURL: strings.TrimSpace(r.FormValue("wikipedia_url")),
		EmailAddress: strings.TrimSpace(r.FormValue("email")),
		Note:         strings.TrimSpace(r.FormValue("note")),
		Errors:       make(map[string]string),
	}

	logger.LogDebug("Form values received - Name: %s, URL: %s, Wiki: %s, Email: %s, Note: %s",
		form.ServiceName, form.ServiceURL, form.WikipediaURL, form.EmailAddress, form.Note)

	// Validate form
	if !validateForm(form) {
		logger.LogDebug("Form validation failed: %v", form.Errors)
		renderNewServiceForm(w, r, lang, form)
		return
	}

	// Process domains
	domains := strings.Split(form.ServiceURL, ",")
	for i := range domains {
		domains[i] = strings.TrimSpace(domains[i])
	}
	domainsStr := strings.Join(domains, ",")

	// Create submission object
	submission := &ServiceRequest{
		Name:      form.ServiceName,
		Domains:   domainsStr,
		Wikipedia: form.WikipediaURL,
		Email:     form.EmailAddress,
		Note:      form.Note,
	}

	// Check if service already exists
	var existing ServiceRequest
	result := db.SubDB.Where("domains LIKE ?", "%"+domains[0]+"%").First(&existing)
	if result.Error == nil {
		// Service exists, increment count
		if err := db.SubDB.Model(&existing).Update("count", gorm.Expr("count + ?", 1)).Error; err != nil {
			logger.LogError(err, "Failed to update submission count")
			form.Errors["general"] = "Failed to process submission. Please try again later."
			renderNewServiceForm(w, r, lang, form)
			return
		}

		session, _ := store.Get(r, "flash-session")
		session.AddFlash("Service already exists. Your submission has been counted.")
		session.Save(r, w)
		http.Redirect(w, r, "/"+lang+"/new_service", http.StatusSeeOther)
		return
	}

	// Add new submission
	if err := db.SubDB.Create(submission).Error; err != nil {
		logger.LogError(err, "Database submission failed")
		form.Errors["general"] = "Failed to submit service. Please try again later."
		renderNewServiceForm(w, r, lang, form)
		return
	}

	// Add success message to session
	session, _ := store.Get(r, "flash-session")
	session.AddFlash("Service submitted successfully!")
	session.Save(r, w)

	logger.LogDebug("Service submission successful, redirecting")
	http.Redirect(w, r, "/"+lang+"/new_service", http.StatusSeeOther)
}

func validateForm(form *ServiceForm) bool {
	isValid := true

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

func renderNewServiceForm(w http.ResponseWriter, r *http.Request, lang string, form *ServiceForm) {
	tmpl, err := parseTemplates("templates/contents/new_service.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load the new service form", err)
		return
	}

	session, _ := store.Get(r, "flash-session")
	flashes := session.Flashes()
	session.Save(r, w)

	data := struct {
		Title         string
		Beta          bool
		Lang          string
		Form          *ServiceForm
		Languages     map[string]string
		FlashMessages []interface{}
	}{
		Title:         localization.Get(lang, "page.newservice"),
		Beta:          isBeta,
		Lang:          lang,
		Form:          form,
		Languages:     SupportedLanguages,
		FlashMessages: flashes,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the new service form", err)
		return
	}

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(buf.Bytes())
}
