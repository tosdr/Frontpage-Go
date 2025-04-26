package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"tosdrgo/handlers/auth"
	"tosdrgo/handlers/localization"

	"github.com/gorilla/mux"
)

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
		Title:     localization.Get(lang, "page.profile"),
		Beta:      isBeta,
		Lang:      lang,
		User:      user,
		Languages: SupportedLanguages,
	}

	w.Header().Set(ContentType, ContentTypeHtml)
	w.Header().Set("Cache-Control", "private, no-store, no-cache, must-revalidate, max-age=0")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to render profile page", err)
	}
}

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

	if !isPartOfTeam(user.Sub) {
		LogoutHandler(w, r)
		return
	}

	if err := auth.SaveUserSession(w, r, user, token); err != nil {
		RenderErrorPage(w, lang, http.StatusInternalServerError, "Failed to save user session", err)
		return
	}

	http.Redirect(w, r, "/en/profile", http.StatusTemporaryRedirect)
}

func isPartOfTeam(uid string) bool {
	url := fmt.Sprintf("https://id.tosdr.org/v1/orgs/%s", uid)
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	type Organization struct {
		Name string `json:"name"`
	}

	var organizations []Organization
	if err := json.NewDecoder(resp.Body).Decode(&organizations); err != nil {
		return false
	}

	for _, org := range organizations {
		if org.Name == "tosdrteam" || org.Name == "tosdrphoenix" {
			return true
		}
	}

	return false
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
