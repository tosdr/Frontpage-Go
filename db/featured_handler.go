package db

import (
	"log"
	"time"

	"tosdrgo/cache"
	"tosdrgo/config"
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

	for _, serviceID := range config.AppConfig.FeaturedServices {
		service, err := FetchServiceData(serviceID)
		if err != nil {
			log.Printf("Error fetching service data for ID %d: %v", serviceID, err)
			continue
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

		featuredServices.Services = append(featuredServices.Services, featuredService)
	}

	cache.SetFeaturedServices(featuredServices)

	return featuredServices, nil
}
