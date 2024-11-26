package handlers

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
	"tosdrgo/db"
	"tosdrgo/models"

	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
	"github.com/yuin/goldmark"
)

var (
	pageCache  = cache.New(4*time.Hour, 10*time.Minute)
	isBeta     bool
	apiBaseURL string
)

func SetIsBeta(value bool) {
	isBeta = value
}

func SetAPIBaseURL(value string) {
	apiBaseURL = value
}

func RedirectToRoot(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func HomeHandler(w http.ResponseWriter, _ *http.Request) {
	// Check if the page is in cache
	if cachedPage, found := pageCache.Get("home"); found {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.gohtml",
		"templates/header.gohtml",
		"templates/footer.gohtml",
		"templates/contents/home.gohtml",
	)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to load the home page")
		return
	}

	featured, err := db.FetchFeaturedServicesData()
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to fetch featured services")
		return
	}

	data := struct {
		Title           string
		Beta            bool
		LastFetchedTime string
		Featured        []models.FeaturedService
	}{
		Title:           "Home Page",
		Beta:            isBeta,
		LastFetchedTime: time.Now().Format(time.RFC850),
		Featured:        featured.Services,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to render the home page")
		return
	}

	// Cache the rendered page
	pageCache.Set("home", buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write(buf.Bytes())
}

func RenderMarkdown(description string) template.HTML {
	md := goldmark.New()
	var buf bytes.Buffer
	if err := md.Convert([]byte(description), &buf); err != nil {
		return ""
	}
	return template.HTML(buf.String())
}

func SiteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	site := vars["sitename"]

	// Check if the page is in cache
	if cachedPage, found := pageCache.Get("view_" + site); found {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.gohtml",
		"templates/header.gohtml",
		"templates/footer.gohtml",
		"templates/contents/markdown.gohtml",
	)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to load the page")
		return
	}

	content, err := os.ReadFile("md/" + site + ".md")
	if err != nil {
		if os.IsNotExist(err) {
			RenderErrorPage(w, http.StatusNotFound, "The requested page was not found")
		} else {
			RenderErrorPage(w, http.StatusInternalServerError, "Failed to load the page content")
		}
		return
	}

	// Extract title from the comment
	title := ""
	lines := bytes.Split(content, []byte("\n"))
	if len(lines) > 0 && bytes.HasPrefix(lines[0], []byte("[//]: <> (TITLE:")) {
		titleStart := bytes.Index(lines[0], []byte("TITLE:")) + 6
		titleEnd := bytes.LastIndex(lines[0], []byte(")"))
		if titleStart < titleEnd {
			title = string(bytes.Trim(lines[0][titleStart:titleEnd], " \""))
		}
	}

	data := struct {
		Content template.HTML
		Title   string
		Beta    bool
	}{
		Content: RenderMarkdown(string(content)),
		Title:   title,
		Beta:    isBeta,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to render the page")
		return
	}

	// Cache the rendered page
	pageCache.Set("view_"+site, buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write(buf.Bytes())
}

func ServiceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["serviceID"]

	tmpl, err := template.ParseFiles(
		"templates/layout.gohtml",
		"templates/header.gohtml",
		"templates/footer.gohtml",
		"templates/contents/service.gohtml",
	)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to load the service page")
		return
	}

	intServiceID, err := strconv.Atoi(serviceID)
	if err != nil {
		log.Printf("Error parsing service ID: %v", err)
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to parse service ID")
		return
	}

	// Fetch service data from API
	service, err := db.FetchServiceData(intServiceID)
	if err != nil {
		RenderErrorPage(w, http.StatusNotFound, "Service not found")
		return
	}

	data := struct {
		Title   string
		Beta    bool
		Service models.Service
	}{
		Title:   service.Name + " - ToS;DR",
		Beta:    isBeta,
		Service: *service,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to render the service page")
		return
	}

	// Cache the rendered page
	pageCache.Set("service_"+serviceID, buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write(buf.Bytes())
}

func SearchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	searchTerm := vars["term"]

	// Check if the page is in cache
	if cachedPage, found := pageCache.Get("search_" + searchTerm); found {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write(cachedPage.([]byte))
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/layout.gohtml",
		"templates/header.gohtml",
		"templates/footer.gohtml",
		"templates/contents/search.gohtml",
	)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to load the search page")
		return
	}

	// Fetch search results from API
	searchResults, err := db.SearchServices(searchTerm)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to fetch search results\n"+err.Error())
		return
	}

	data := struct {
		Title         string
		Beta          bool
		SearchTerm    string
		SearchResults []models.SearchResult
	}{
		Title:         "Search Results - ToS;DR",
		Beta:          isBeta,
		SearchTerm:    searchTerm,
		SearchResults: searchResults,
	}

	var buf bytes.Buffer
	err = tmpl.ExecuteTemplate(&buf, "layout", data)
	if err != nil {
		RenderErrorPage(w, http.StatusInternalServerError, "Failed to render the search results")
		return
	}

	// Cache the rendered page
	pageCache.Set("search_"+searchTerm, buf.Bytes(), cache.DefaultExpiration)

	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write(buf.Bytes())
}

func RenderErrorPage(w http.ResponseWriter, errorCode int, errorMessage string) {
	tmpl, err := template.ParseFiles(
		"templates/layout.gohtml",
		"templates/header.gohtml",
		"templates/footer.gohtml",
		"templates/contents/error.gohtml",
	)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	data := struct {
		Title        string
		Beta         bool
		ErrorCode    int
		ErrorMessage string
	}{
		Title:        fmt.Sprintf("Error %d", errorCode),
		Beta:         isBeta,
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	}

	w.WriteHeader(errorCode)
	err = tmpl.ExecuteTemplate(w, "layout", data)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
