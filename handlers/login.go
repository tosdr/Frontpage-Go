package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"tosdrgo/handlers/auth"

	"github.com/gorilla/mux"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	// Generate random state
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to generate state", err)
		return
	}
	state := base64.StdEncoding.EncodeToString(b)

	// Store state in session
	session, _ := auth.GetStore().Get(r, "auth-session")
	session.Values["state"] = state
	err = session.Save(r, w)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to save state", err)
		return
	}

	// Redirect to Auth0
	loginURL := auth.GetLoginURL(state)
	http.Redirect(w, r, loginURL, http.StatusTemporaryRedirect)
}

func CallbackHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	session, err := auth.GetStore().Get(r, "auth-session")
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to get session", err)
		return
	}

	state := session.Values["state"]
	if state == nil {
		RenderErrorPage(w, lang, http.StatusBadRequest, "No state value in session", nil)
		return
	}

	if r.URL.Query().Get("state") != state.(string) {
		RenderErrorPage(w, lang, http.StatusBadRequest, "Invalid state parameter", nil)
		return
	}

	// Clear the state
	delete(session.Values, "state")
	err = session.Save(r, w)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to save session", err)
		return
	}

	code := r.URL.Query().Get("code")
	token, err := auth.Exchange(code)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to exchange token", err)
		return
	}

	user, err := auth.GetUserInfo(token)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to get user info", err)
		return
	}

	if err := auth.SaveUserSession(w, r, user, token); err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to save user session", err)
		return
	}

	http.Redirect(w, r, "/en/profile", http.StatusTemporaryRedirect)
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	user, err := auth.GetUserSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	tmpl, err := parseTemplates("templates/contents/profile.gohtml", lang, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to load profile page", err)
		return
	}

	data := struct {
		Title     string
		Beta      bool
		Lang      string
		User      *auth.A0User
		Languages map[string]string
	}{
		Title:     "Profile - ToS;DR",
		Beta:      isBeta,
		Lang:      lang,
		User:      user,
		Languages: SupportedLanguages,
	}

	w.Header().Set(ContentType, ContentTypeHtml)
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render profile page", err)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	lang := vars["lang"]

	logoutURL, err := auth.Logout(w, r)
	if err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to logout", err)
		return
	}

	// Redirect to Auth0 logout
	http.Redirect(w, r, logoutURL, http.StatusTemporaryRedirect)
}
