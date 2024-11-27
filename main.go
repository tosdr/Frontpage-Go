package main

import (
	"log"
	"net/http"
	"strings"
	"tosdrgo/config"
	"tosdrgo/db"
	"tosdrgo/handlers"

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

func main() {
	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	r := mux.NewRouter()

	// Health check endpoint
	r.HandleFunc("/v1/health", handlers.HealthCheckHandler).Methods("GET")

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
	r.HandleFunc("/about", handlers.MinifyMiddleware(handlers.AboutHandler))
	r.HandleFunc("/thanks", handlers.MinifyMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("TBD"))
	}))
	r.HandleFunc("/service/{serviceID}", handlers.MinifyMiddleware(handlers.ServiceHandler))
	r.HandleFunc("/sites/{sitename}", handlers.MinifyMiddleware(handlers.SiteHandler))
	r.HandleFunc("/search/{term}", handlers.MinifyMiddleware(handlers.SearchHandler))
	r.HandleFunc("/shield/{serviceID}", handlers.ShieldHandler).Methods("GET")

	// legacy shield -> we route shields.tosdr.org/en_XYZ.svg to here
	r.HandleFunc("/legacyshield/en_{serviceID}.svg", handlers.ShieldHandler).Methods("GET")

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlers.RenderErrorPage(w, http.StatusNotFound, "The requested page was not found")
	})

	//goland:noinspection GoBoolExpressions
	handlers.SetIsBeta(IsBeta)

	// Start the server
	log.Printf("Server starting on 0.0.0.0:80")
	log.Fatal(http.ListenAndServe("0.0.0.0:80", r))
}
