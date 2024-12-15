package cache

import (
	"fmt"
	"time"
	"tosdrgo/handlers/metrics"
	"tosdrgo/models"

	"github.com/patrickmn/go-cache"
)

var (
	c = cache.New(5*time.Minute, 10*time.Minute)
)

func GetService(id int, lang string) (*models.Service, bool) {
	if x, found := c.Get(getServiceKey(id, lang)); found {
		metrics.CacheHits.WithLabelValues("service").Inc()
		return x.(*models.Service), true
	}
	metrics.CacheMisses.WithLabelValues("service").Inc()
	return nil, false
}

func SetService(id int, lang string, service *models.Service) {
	c.Set(getServiceKey(id, lang), service, cache.DefaultExpiration)
}

func GetFeaturedServices(lang string) (*models.FeaturedServices, bool) {
	if x, found := c.Get(getFeaturedServicesKey(lang)); found {
		metrics.CacheHits.WithLabelValues("featured").Inc()
		return x.(*models.FeaturedServices), true
	}
	metrics.CacheMisses.WithLabelValues("featured").Inc()
	return nil, false
}

func SetFeaturedServices(lang string, services *models.FeaturedServices) {
	c.Set(getFeaturedServicesKey(lang), services, 24*time.Hour)
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

func getServiceKey(id int, lang string) string {
	return fmt.Sprintf("service_%d_%s", id, lang)
}

func getSearchKey(term string) string {
	return fmt.Sprintf("search_%s", term)
}

func getFeaturedServicesKey(lang string) string {
	return fmt.Sprintf("featured_services_%s", lang)
}
