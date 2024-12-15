package db

import (
	"fmt"
	"time"
	"tosdrgo/handlers/cache"
	"tosdrgo/internal/config"
	"tosdrgo/internal/logger"

	"tosdrgo/models"
)

// Helper struct to hold result data
type fetchResult struct {
	index   int
	service models.FeaturedService
}

// fetchServiceConcurrently handles the concurrent fetching of a single service
func fetchServiceConcurrently(idx int, serviceID int, lang string, resultChan chan<- fetchResult, errorChan chan<- error) {
	service, err := FetchServiceData(serviceID, lang)
	if err != nil {
		logger.LogError(err, fmt.Sprintf("Error fetching service data for ID %d", serviceID))
		errorChan <- err
		return
	}

	featuredService := models.FeaturedService{
		ID:    service.ID,
		Name:  service.Name,
		Icon:  service.Image,
		Grade: service.Rating,
	}

	// Get first 5 points
	for i := 0; i < len(service.Points) && i < 5; i++ {
		featuredService.Points = append(featuredService.Points, service.Points[i])
	}

	resultChan <- fetchResult{idx, featuredService}
}

// compactServices removes any nil entries from the services slice
func compactServices(services []models.FeaturedService) []models.FeaturedService {
	compactedServices := make([]models.FeaturedService, 0, len(services))
	for _, service := range services {
		if service.ID != 0 { // Check if it's not a zero value
			compactedServices = append(compactedServices, service)
		}
	}
	return compactedServices
}

func FetchFeaturedServicesData(lang string) (*models.FeaturedServices, error) {
	if cachedServices, found := cache.GetFeaturedServices(lang); found {
		return cachedServices, nil
	}

	featuredServices := &models.FeaturedServices{
		Services:  make([]models.FeaturedService, len(config.AppConfig.FeaturedServices)),
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Create channels for results and errors
	resultChan := make(chan fetchResult, len(config.AppConfig.FeaturedServices))
	errorChan := make(chan error, len(config.AppConfig.FeaturedServices))

	// Launch goroutines for parallel fetching
	for i, serviceID := range config.AppConfig.FeaturedServices {
		go fetchServiceConcurrently(i, serviceID, lang, resultChan, errorChan)
	}

	// Collect results
	validResults := 0
	for i := 0; i < len(config.AppConfig.FeaturedServices); i++ {
		select {
		case result := <-resultChan:
			featuredServices.Services[result.index] = result.service
			validResults++
		case err := <-errorChan:
			logger.LogError(err, "Error collecting featured service")
		}
	}

	// Trim any nil entries if needed
	if validResults < len(featuredServices.Services) {
		featuredServices.Services = compactServices(featuredServices.Services)
	}

	cache.SetFeaturedServices(lang, featuredServices)
	return featuredServices, nil
}
