package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

func init() {
	gob.Register(&A0User{})
	gob.Register(&oauth2.Token{})
}

var (
	store        *sessions.CookieStore
	config       oauth2.Config
	logoutReturn string
	loginDomain  string
)

const cookieName = "auth-session"

type A0User struct {
	Sub           string `json:"sub"`
	Nickname      string `json:"nickname"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
}

func InitStore(sessionKey string) {
	if sessionKey == "" {
		key := make([]byte, 32)
		if _, err := rand.Read(key); err != nil {
			panic(fmt.Errorf("failed to generate random session key: %v", err))
		}
		sessionKey = base64.StdEncoding.EncodeToString(key)
	}
	store = sessions.NewCookieStore([]byte(sessionKey))
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func Init(domain, clientID, clientSecret, redirectURI, sessionKey string, logoutReturnUrl string) {
	InitStore(sessionKey)

	config = oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURI,
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  domain + "/authorize",
			TokenURL: domain + "/oauth/token",
		},
	}
	loginDomain = domain
	logoutReturn = logoutReturnUrl
}

func GetLoginURL(state string) string {
	return config.AuthCodeURL(state, oauth2.AccessTypeOffline)
}

func Exchange(code string) (*oauth2.Token, error) {
	return config.Exchange(context.Background(), code)
}

func GetUserInfo(token *oauth2.Token) (*A0User, error) {
	client := config.Client(context.Background(), token)
	resp, err := client.Get(config.Endpoint.AuthURL + "/userinfo")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var user A0User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

func SaveUserSession(w http.ResponseWriter, r *http.Request, user *A0User, token *oauth2.Token) error {
	session, err := store.Get(r, cookieName)
	if err != nil {
		return fmt.Errorf("failed to get session: %v", err)
	}

	session.Values["user"] = user
	session.Values["token"] = token

	return session.Save(r, w)
}

func GetUserSession(r *http.Request) (*A0User, error) {
	session, _ := store.Get(r, cookieName)
	if user, ok := session.Values["user"].(*A0User); ok {
		return user, nil
	}
	return nil, errors.New("no user in session")
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, _ := store.Get(r, cookieName)
	session.Options.MaxAge = -1
	return session.Save(r, w)
}

func GetStore() *sessions.CookieStore {
	return store
}

func Logout(w http.ResponseWriter, r *http.Request) (string, error) {
	if err := ClearSession(w, r); err != nil {
		return "", fmt.Errorf("failed to clear session: %v", err)
	}

	logoutURL := fmt.Sprintf(
		"%s/v2/logout?client_id=%s&returnTo=%s",
		loginDomain,
		config.ClientID,
		logoutReturn,
	)

	return logoutURL, nil
}
