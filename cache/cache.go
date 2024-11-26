package cache

import (
	"fmt"
	"time"
	"tosdrgo/models"

	"github.com/patrickmn/go-cache"
)

var (
	c = cache.New(5*time.Minute, 10*time.Minute)
)

func GetService(id int) (*models.Service, bool) {
	if x, found := c.Get(getServiceKey(id)); found {
		return x.(*models.Service), true
	}
	return nil, false
}

func SetService(id int, service *models.Service) {
	c.Set(getServiceKey(id), service, cache.DefaultExpiration)
}

func GetFeaturedServices() (*models.FeaturedServices, bool) {
	if x, found := c.Get("featured_services"); found {
		return x.(*models.FeaturedServices), true
	}
	return nil, false
}

func SetFeaturedServices(services *models.FeaturedServices) {
	c.Set("featured_services", services, 24*time.Hour)
}

func GetSearchResults(term string) ([]models.SearchResult, bool) {
	if x, found := c.Get(getSearchKey(term)); found {
		return x.([]models.SearchResult), true
	}
	return nil, false
}

func SetSearchResults(term string, results []models.SearchResult) {
	c.Set(getSearchKey(term), results, cache.DefaultExpiration)
}

func ClearCache() {
	c.Flush()
}

func getServiceKey(id int) string {
	return fmt.Sprintf("service_%d", id)
}

func getSearchKey(term string) string {
	return fmt.Sprintf("search_%s", term)
}
