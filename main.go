package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"tosdrgo/handlers"

	"github.com/gorilla/mux"
)

var IsBeta = true
var apiBaseURL string
var serverPort int

func init() {
	if err := LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	apiBaseURL = AppConfig.APIBaseURL
	serverPort = AppConfig.ServerPort
}

func setCSSContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()

	// Serve static files with content type middleware and minification for CSS
	r.PathPrefix("/static/css/").Handler(handlers.MinifyMiddlewareHandler(
		setCSSContentType(http.StripPrefix("/static/css/", http.FileServer(http.Dir("static/css"))))))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Define routes with minify middleware
	r.HandleFunc("/", handlers.MinifyMiddleware(handlers.HomeHandler))
	r.HandleFunc("/{lang:[a-z]{2}}", handlers.RedirectToRoot)
	r.HandleFunc("/{lang:[a-z]{2}}/{path:.*}", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		http.Redirect(w, r, "/"+vars["path"], http.StatusSeeOther)
	})
	r.HandleFunc("/service/{serviceID}", handlers.MinifyMiddleware(handlers.ServiceHandler))
	r.HandleFunc("/sites/{sitename}", handlers.MinifyMiddleware(handlers.SiteHandler))
	r.HandleFunc("/search/{term}", handlers.MinifyMiddleware(handlers.SearchHandler))

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.RenderErrorPage(w, http.StatusNotFound, "The requested page was not found")
	})

	//goland:noinspection GoBoolExpressions
	handlers.SetIsBeta(IsBeta)

	handlers.SetAPIBaseURL(apiBaseURL)

	// Start the server
	log.Printf("Server starting on :%d", serverPort)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", serverPort), r))
}
