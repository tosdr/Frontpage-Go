package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"tosdrgo/internal/db"
	"tosdrgo/internal/logger"

	"github.com/essentialkaos/go-badge"
	"github.com/gorilla/mux"
	"github.com/patrickmn/go-cache"
)

var (
	shieldCache = cache.New(24*time.Hour, 24*time.Hour)
)

const (
	contentType = "Content-Type"
	svgType     = "image/svg+xml"
)

func ShieldHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	serviceID := vars["serviceID"]

	// Check cache first
	if cachedShield, found := shieldCache.Get(serviceID); found {
		w.Header().Set(contentType, svgType)
		_, _ = w.Write(cachedShield.([]byte))
		return
	}

	// Parse service ID and fetch data
	intServiceID, err := strconv.Atoi(serviceID)
	if err != nil {
		logger.LogError(err, "Invalid service ID in shield handler")
		svg, _ := returnShield("Error", "Service not found!", gradeToHexColor("E"))
		w.Header().Set(contentType, svgType)
		_, _ = w.Write(svg)
		return
	}

	service, err := db.FetchServiceData(intServiceID)
	if err != nil {
		svg, _ := returnShield("Error", "Service not found!", gradeToHexColor("E"))
		w.Header().Set(contentType, svgType)
		_, _ = w.Write(svg)
		return
	}

	svg, err := returnShield(service.Name, fmt.Sprintf("Grade %s", service.Rating), gradeToHexColor(service.Rating))
	if err != nil {
		http.Error(w, "Error creating badge generator", http.StatusNotFound)
		return
	}

	// Cache the result
	shieldCache.Set(serviceID, svg, cache.DefaultExpiration)

	// Send response
	w.Header().Set(contentType, svgType)
	_, _ = w.Write(svg)
}

func returnShield(title string, subtitle string, color string) ([]byte, error) {
	// Initialize badge generator
	g, err := badge.NewGenerator("assets/opensans.ttf", 11)
	if err != nil {
		log.Printf("Error creating badge generator: %v", err)
		return nil, err
	}

	// Generate SVG badge
	return g.GenerateFlat(
		title,
		subtitle,
		color,
	), nil
}

// Helper to convert grade to hex color
func gradeToHexColor(grade string) string {
	colors := map[string]string{
		"A":   "#198754",
		"B":   "#79b752",
		"C":   "#ffc107",
		"D":   "#d66f2c",
		"E":   "#dc3545",
		"N/A": "#6c757d",
	}
	return colors[grade]
}
