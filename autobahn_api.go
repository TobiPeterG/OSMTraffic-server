package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// Warning represents the structure of a warning item in the AutobahnAPI response.
type Warning struct {
	Identifier               string        `json:"identifier"`
	Icon                     string        `json:"icon"`
	IsBlocked                string        `json:"isBlocked"`
	Future                   bool          `json:"future"`
	Extent                   string        `json:"extent"`
	Point                    string        `json:"point"`
	StartLcPosition          string        `json:"startLcPosition"`
	DisplayType              string        `json:"display_type"`
	Subtitle                 string        `json:"subtitle"`
	Title                    string        `json:"title"`
	StartTimestamp           string        `json:"startTimestamp"`
	DelayTimeValue           string        `json:"delayTimeValue"`
	AbnormalTrafficType      string        `json:"abnormalTrafficType"`
	AverageSpeed             string        `json:"averageSpeed"`
	Coordinate               Coordinate    `json:"coordinate"`
	Description              []string      `json:"description"`
	RouteRecommendation      []interface{} `json:"routeRecommendation"`
	Footer                   []interface{} `json:"footer"`
	LorryParkingFeatureIcons []interface{} `json:"lorryParkingFeatureIcons"`
	Geometry                 Geometry      `json:"geometry"`
}

// Geometry represents the geometry information in the warning.
type Geometry struct {
	Type        string      `json:"type"`
	Coordinates [][]float64 `json:"coordinates"`
}

// AutobahnResponse represents the top-level structure of the API response.
type AutobahnResponse struct {
	Warnings []Warning `json:"warning"`
}

// APIURLs is a slice of URLs to fetch data from.
// You can add more URLs here if needed.
var APIURLs = []string{
	"https://verkehr.autobahn.de/o/autobahn/A1/services/warning",
	// Add more URLs as needed
}

// FetchWarnings fetches data from the AutobahnAPI and parses the relevant warnings.
func FetchWarnings() ([]Warning, error) {
	var allWarnings []Warning
	for _, url := range APIURLs {
		warnings, err := fetchWarningsFromURL(url)
		if err != nil {
			log.Printf("Error fetching from %s: %v", url, err)
			continue
		}
		allWarnings = append(allWarnings, warnings...)
	}
	return allWarnings, nil
}

// fetchWarningsFromURL fetches warnings from a single URL.
func fetchWarningsFromURL(url string) ([]Warning, error) {
	// Make the HTTP GET request.
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data from %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check for non-200 status codes.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code %d from %s", resp.StatusCode, url)
	}

	// Read the response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body from %s: %w", url, err)
	}

	// Parse the JSON response.
	var data AutobahnResponse
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON from %s: %w", url, err)
	}

	// Filter the warnings based on allowed event types.
	filteredWarnings := FilterWarnings(data.Warnings)
	return filteredWarnings, nil
}

// FilterWarnings filters warnings based on allowed Autobahn API event types.
func FilterWarnings(warnings []Warning) []Warning {
	allowedTypes := map[string]bool{
		"STATIONARY_TRAFFIC": true,
		"SLOW_TRAFFIC":       true,
		"HEAVY_TRAFFIC":      true,
		"TRAFFIC_BUILDUP":    true,
		"ACCIDENT":           true,
		"ROADWORKS":          true,
		"BLOCKAGE":           true,
		"HAZARD":             true,
		// Add more allowed types as needed
	}

	filtered := []Warning{}
	for _, warning := range warnings {
		if allowedTypes[warning.AbnormalTrafficType] {
			filtered = append(filtered, warning)
		} else {
			log.Printf("Filtering out warning with type: %s", warning.AbnormalTrafficType)
		}
	}
	return filtered
}
