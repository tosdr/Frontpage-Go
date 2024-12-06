package db

import (
	"fmt"
	"time"
	"tosdrgo/handlers/cache"
	"tosdrgo/internal/config"
	"tosdrgo/internal/logger"

	"tosdrgo/models"
)

func FetchFeaturedServicesData() (*models.FeaturedServices, error) {
	if cachedServices, found := cache.GetFeaturedServices(); found {
		return cachedServices, nil
	}

	featuredServices := &models.FeaturedServices{
		Services:  make([]models.FeaturedService, 0, len(config.AppConfig.FeaturedServices)),
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Create channels for results and errors
	resultChan := make(chan models.FeaturedService, len(config.AppConfig.FeaturedServices))
	errorChan := make(chan error, len(config.AppConfig.FeaturedServices))

	// Launch goroutines for parallel fetching
	for _, serviceID := range config.AppConfig.FeaturedServices {
		go func(id int) {
			service, err := FetchServiceData(id)
			if err != nil {
				logger.LogError(err, fmt.Sprintf("Error fetching service data for ID %d", id))
				errorChan <- err
				return
			}

			featuredService := models.FeaturedService{
				ID:    service.ID,
				Name:  service.Name,
				Icon:  service.Image,
				Grade: service.Rating,
			}

			for i := 0; i < len(service.Points) && i < 5; i++ {
				featuredService.Points = append(featuredService.Points, service.Points[i])
			}

			resultChan <- featuredService
		}(serviceID)
	}

	// Collect results
	for i := 0; i < len(config.AppConfig.FeaturedServices); i++ {
		select {
		case service := <-resultChan:
			featuredServices.Services = append(featuredServices.Services, service)
		case err := <-errorChan:
			// Log error but continue collecting other results
			logger.LogError(err, "Error collecting featured service")
		}
	}

	cache.SetFeaturedServices(featuredServices)
	return featuredServices, nil
}
