package handlers

import (
	"bytes"
	"net/http"
	"strconv"
	"tosdrgo/auth"
	"tosdrgo/db"

	"github.com/gorilla/mux"
)

type DashboardData struct {
	Submissions []db.ServiceSubmissionWithStatus
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
		Title     string
		Beta      bool
		Lang      string
		User      *auth.A0User
		Dashboard DashboardData
		Languages map[string]string
	}{
		Title: "Dashboard - ToS;DR",
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
		Languages: SupportedLanguages,
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
