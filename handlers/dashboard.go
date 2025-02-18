package handlers

import (
	"bytes"
	"net/http"
	"strconv"
	"tosdrgo/handlers/auth"
	"tosdrgo/handlers/localization"
	"tosdrgo/internal/db"
	"tosdrgo/internal/logger"

	"github.com/gorilla/mux"
)

type DashboardData struct {
	Submissions []db.ServiceSubmission
	Page        int
	TotalPages  int
	HasNext     bool
	HasPrev     bool
}

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	// Check if user is logged in
	user, err := auth.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	// Get page number from query params
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Fetch submissions with pagination
	submissions, total, err := db.GetSubmissions(page, 50)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to fetch submissions", err)
		return
	}

	// Calculate pagination info
	totalPages := (total + 49) / 50 // Round up division
	hasNext := page < totalPages
	hasPrev := page > 1

	data := struct {
		Title      string
		Beta       bool
		Lang       string
		User       *auth.A0User
		Dashboard  DashboardData
		SearchTerm *string
		Languages  map[string]string
	}{
		Title: localization.Get(lang, "page.dashboard"),
		Beta:  isBeta,
		Lang:  lang,
		User:  user,
		Dashboard: DashboardData{
			Submissions: submissions,
			Page:        page,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrev:     hasPrev,
		},
		SearchTerm: nil,
		Languages:  SupportedLanguages,
	}

	tmpl, err := parseTemplates("templates/contents/dashboard.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load dashboard template", err)
		return
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, "layout", data); err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render dashboard", err)
		return
	}

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(buf.Bytes())
}

func DashboardSearchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]
	searchTerm := vars["term"]

	user, err := auth.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	submissions, total, err := db.SearchSubmissions(searchTerm, page, 50)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to fetch submissions", err)
		return
	}

	totalPages := (total + 49) / 50
	hasNext := page < totalPages
	hasPrev := page > 1

	data := struct {
		Title      string
		Beta       bool
		Lang       string
		User       *auth.A0User
		Dashboard  DashboardData
		SearchTerm string
		Languages  map[string]string
	}{
		Title:      localization.Get(lang, "page.dashboard"),
		Beta:       isBeta,
		Lang:       lang,
		User:       user,
		SearchTerm: searchTerm,
		Dashboard: DashboardData{
			Submissions: submissions,
			Page:        page,
			TotalPages:  totalPages,
			HasNext:     hasNext,
			HasPrev:     hasPrev,
		},
		Languages: SupportedLanguages,
	}

	tmpl, err := parseTemplates("templates/contents/dashboard.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load dashboard template", err)
		return
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render the dashboard", err)
		return
	}

	w.Header().Set(ContentType, ContentTypeHtml)
	_, _ = w.Write(buf.Bytes())
}

func HandleSubmissionAction(w http.ResponseWriter, r *http.Request) {
	// Check if user is logged in
	user, err := auth.GetUserSession(r)

	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if !isPartOfTeam(user.Sub) {
		LogoutHandler(w, r)
	}

	vars := mux.Vars(r)
	id := vars["id"]
	action := vars["action"]

	// Validate action
	if action != "accept" && action != "deny" {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// Update submission status in database
	err = db.UpdateSubmissionStatus(id, action)
	if err != nil {
		logger.LogError(err, "Failed to update submission")
		http.Error(w, "Failed to update submission", http.StatusInternalServerError)
		return
	}

	logger.LogDebug("Submission accepted by " + user.Email)

	w.WriteHeader(http.StatusOK)
}
