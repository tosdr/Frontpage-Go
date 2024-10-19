package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"tosdrgo/api/structs"
)

func SearchServices(term string) ([]structs.SearchResult, error) {
	url := fmt.Sprintf("http://localhost:8080/search/v1/%s", term)
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusBadRequest {
		return nil, fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
	}

	var apiResponse structs.SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&apiResponse)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %v", err)
	}

	if apiResponse.Status != "success" {
		return nil, fmt.Errorf("API returned error: %s", apiResponse.Message)
	}

	return apiResponse.Data, nil
}
