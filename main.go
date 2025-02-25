package main

import (
	"log"
	"net/http"
	"strings"
	"time"
	"tosdrgo/handlers"
	"tosdrgo/handlers/auth"
	"tosdrgo/handlers/metrics"
	"tosdrgo/handlers/middleware"
	"tosdrgo/internal/config"
	db2 "tosdrgo/internal/db"
	"tosdrgo/internal/email"
	"tosdrgo/internal/logger"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var IsBeta = false

func init() {
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	auth.Init(
		config.AppConfig.Login.Domain,
		config.AppConfig.Login.ClientID,
		config.AppConfig.Login.ClientSecret,
		config.AppConfig.Login.RedirectURI,
		config.AppConfig.Login.SessionKey,
		config.AppConfig.Login.LogoutReturn,
	)

	handlers.InitContact(config.AppConfig.Webhook)

	if err := email.Init(); err != nil {
		log.Fatalf("Failed to initialize email client: %v", err)
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

// Add basic auth middleware for metrics
func basicAuthMiddleware(username, password string) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || user != username || pass != password {
				w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func main() {
	if err := db2.InitDB(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db2.CloseDB()

	// Initial indexing
	db2.IndexSearch()

	// Start background indexing
	go func() {
		ticker := time.NewTicker(12 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			logger.LogDebug("Running scheduled search index update")
			db2.IndexSearch()
		}
	}()

	r := mux.NewRouter()

	// Add metrics middleware to all routes
	r.Use(metrics.MetricsMiddleware)
	r.Use(loggingMiddleware)

	// Create a subrouter for metrics with authentication
	metricsRouter := r.PathPrefix("/metrics").Subrouter()
	metricsRouter.Use(basicAuthMiddleware(
		config.AppConfig.MetricsUsername,
		config.AppConfig.MetricsPassword,
	))
	metricsRouter.Handle("", promhttp.Handler())

	r.HandleFunc("/v1/health", handlers.HealthCheckHandler).Methods("GET").Name("health")

	// Serve static files with content type middleware and minification for CSS
	r.PathPrefix("/static/css/").Handler(handlers.MinifyMiddlewareHandler(
		setCSSContentType(http.StripPrefix("/static/css/", http.FileServer(http.Dir("static/css"))))))
	r.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static/"))))

	// Shield endpoints (without language prefix)
	r.HandleFunc("/{lang:[a-z]{2}}/shield/{serviceID}", handlers.ShieldHandler).Methods("GET")
	r.HandleFunc("/legacyshield/{lang:[a-z]{2}}_{serviceID}.svg", handlers.ShieldHandler).Methods("GET")

	// Add redirects for non-language-prefixed routes
	r.HandleFunc("/service/{serviceID}", handlers.DetectLanguageAndRedirectWithPath).Methods("GET")
	r.HandleFunc("/about", handlers.DetectLanguageAndRedirectWithPath).Methods("GET")
	r.HandleFunc("/donate", handlers.RedirectDonate).Methods("GET")
	r.HandleFunc("/thanks", handlers.DetectLanguageAndRedirectWithPath).Methods("GET")
	r.HandleFunc("/sites/{sitename}", handlers.DetectLanguageAndRedirectWithPath).Methods("GET")

	// Root redirect to browser language
	r.HandleFunc("/", handlers.DetectLanguageAndRedirect)

	// forward old donate page to new
	r.HandleFunc("/{lang:[a-z]{2}}/sites/donate", handlers.RedirectDonate)

	// Language-prefixed routes
	r.HandleFunc("/{lang:[a-z]{2}}", handlers.MinifyMiddleware(handlers.HomeHandler)).Name("home")
	r.HandleFunc("/{lang:[a-z]{2}}/", handlers.MinifyMiddleware(handlers.HomeHandler))
	r.HandleFunc("/{lang:[a-z]{2}}/about", handlers.MinifyMiddleware(handlers.AboutHandler)).Name("about")
	r.HandleFunc("/{lang:[a-z]{2}}/donate", handlers.MinifyMiddleware(handlers.DonateHandler)).Name("about")
	r.HandleFunc("/{lang:[a-z]{2}}/thanks", handlers.MinifyMiddleware(handlers.ThanksHandler)).Name("thanks")
	r.HandleFunc("/{lang:[a-z]{2}}/service/{serviceID}", handlers.MinifyMiddleware(handlers.ServiceHandler)).Name("service")
	r.HandleFunc("/{lang:[a-z]{2}}/sites/{sitename}", handlers.MinifyMiddleware(handlers.SiteHandler))
	r.HandleFunc("/{lang:[a-z]{2}}/new_service", handlers.MinifyMiddleware(handlers.NewServiceHandler)).Methods("GET", "POST").Name("new_service")
	r.HandleFunc("/{lang:[a-z]{2}}/services/{grade}", handlers.MinifyMiddleware(handlers.GradedServicesHandler))
	r.HandleFunc("/{lang:[a-z]{2}}/contact", handlers.MinifyMiddleware(handlers.ContactHandler)).Name("contact")

	searchRouter := r.PathPrefix("/{lang:[a-z]{2}}/search").Subrouter()
	searchRouter.Use(middleware.RateLimitMiddleware)
	searchRouter.HandleFunc("/{term:.*}", handlers.MinifyMiddleware(handlers.SearchHandler))

	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Track 404 paths
		metrics.NotFoundPaths.WithLabelValues(r.URL.Path, r.Method).Inc()
		handlers.RenderErrorPage(w, "en", http.StatusNotFound, "The requested page was not found", nil)
	})

	//goland:noinspection GoBoolExpressions
	handlers.SetIsBeta(IsBeta)

	// Auth routes
	r.HandleFunc("/login", handlers.LoginHandler).Methods("GET").Name("login")
	r.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET").Name("logout")
	r.HandleFunc("/auth/callback", handlers.CallbackHandler).Methods("GET").Name("auth_callback")
	r.HandleFunc("/{lang:[a-z]{2}}/profile", handlers.ProfileHandler).Methods("GET").Name("profile")

	// Dashboard route
	r.HandleFunc("/{lang:[a-z]{2}}/dashboard", handlers.DashboardHandler).Methods("GET").Name("dashboard")
	r.HandleFunc("/{lang:[a-z]{2}}/dashboard/{term}", handlers.MinifyMiddleware(handlers.DashboardSearchHandler))

	r.HandleFunc("/api/submissions/{id}/{action}", handlers.HandleSubmissionAction).Methods("POST").Name("submission_action")

	r.HandleFunc("/api/teams", handlers.HandleTeamAction).Methods("GET")

	// Start the server
	log.Printf("Server starting on 0.0.0.0:80")
	log.Fatal(http.ListenAndServe("0.0.0.0:80", r))
}
