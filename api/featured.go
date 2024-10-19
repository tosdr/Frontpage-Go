package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"tosdrgo/api/structs"
)

func FetchFeaturedServices(apiBaseURL string) (*structs.FeaturedServices, error) {
	url := apiBaseURL + "/featured/v1"
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var apiResponse structs.FeaturedResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if apiResponse.Status != "success" {
		return nil, fmt.Errorf("API returned error: %s", apiResponse.Message)
	}

	return &apiResponse.Data.FeaturedServices, nil
}
