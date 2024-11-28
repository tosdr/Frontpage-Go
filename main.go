package main

import (
	"log"
	"net/http"
	"strings"
	"time"
	"tosdrgo/config"
	"tosdrgo/db"
	"tosdrgo/handlers"
	"tosdrgo/logger"

	"github.com/gorilla/mux"
)

var IsBeta = true

func init() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
}

func setCSSContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".css") {
			w.Header().Set("Content-Type", "text/css")
		}
		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.LogRequest(r, time.Since(start))
	})
}

func main() {
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	r := mux.NewRouter()

	// Add logging middleware to all routes
	r.Use(loggingMiddleware)

	// Health check endpoint
	r.HandleFunc("/v1/health", handlers.HealthCheckHandler).Methods("GET")

	// Serve static files with content type middleware and minification for CSS
	r.PathPrefix("/static/css/").Handler(handlers.MinifyMiddlewareHandler(
		setCSSContentType(http.StripPrefix("/static/css/", http.FileServer(http.Dir("static/css"))))))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Shield endpoints (without language prefix)
	r.HandleFunc("/shield/{serviceID}", handlers.ShieldHandler).Methods("GET")
	r.HandleFunc("/legacyshield/en_{serviceID}.svg", handlers.ShieldHandler).Methods("GET")

	// Root redirect to browser language
	r.HandleFunc("/", handlers.DetectLanguageAndRedirect)

	// Language-prefixed routes
	r.HandleFunc("/{lang:[a-z]{2}}", handlers.MinifyMiddleware(handlers.HomeHandler))
	r.HandleFunc("/{lang:[a-z]{2}}/", handlers.MinifyMiddleware(handlers.HomeHandler))
	r.HandleFunc("/{lang:[a-z]{2}}/about", handlers.MinifyMiddleware(handlers.AboutHandler))
	r.HandleFunc("/{lang:[a-z]{2}}/thanks", handlers.MinifyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("TBD"))
	}))
	r.HandleFunc("/{lang:[a-z]{2}}/service/{serviceID}", handlers.MinifyMiddleware(handlers.ServiceHandler))
	r.HandleFunc("/{lang:[a-z]{2}}/sites/{sitename}", handlers.MinifyMiddleware(handlers.SiteHandler))
	r.HandleFunc("/{lang:[a-z]{2}}/search/{term}", handlers.MinifyMiddleware(handlers.SearchHandler))

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.RenderErrorPage(w, "en", http.StatusNotFound, "The requested page was not found", nil)
	})

	//goland:noinspection GoBoolExpressions
	handlers.SetIsBeta(IsBeta)

	// Start the server
	log.Printf("Server starting on 0.0.0.0:80")
	log.Fatal(http.ListenAndServe("0.0.0.0:80", r))
}
