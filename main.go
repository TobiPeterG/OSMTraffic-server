package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "redis:6379"
	}

	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	// Create a new router using Chi
	r := chi.NewRouter()

	// Define routes
	r.Get("/traffic", TrafficHandler)

	// Start the HTTP server
	fmt.Println("Starting server on :8080...")
	http.ListenAndServe(":8080", r)
}

// TrafficHandler handles requests to the /traffic endpoint
func TrafficHandler(w http.ResponseWriter, r *http.Request) {
	// Fetch traffic data (integrating Datex2 API)
	trafficData := FetchTrafficData()

	// Send the traffic data to the client in GeoJSON format
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(trafficData))
}

// FetchTrafficData fetches traffic data, checks Redis cache first
func FetchTrafficData() string {
	ctx := context.Background()

	// Try to get data from Redis cache
	val, err := rdb.Get(ctx, "traffic_data").Result()
	if err == redis.Nil {
		// Cache miss, so fetch from the data source (Datex2)
		fmt.Println("Cache miss. Fetching from Datex2...")
		trafficData := GetTrafficData()

		// Store in cache with an expiration time (e.g., 5 minutes)
		rdb.Set(ctx, "traffic_data", trafficData, 5*time.Minute)

		return trafficData
	} else if err != nil {
		// Handle other Redis errors
		log.Println("Redis error:", err)
		return "[]"
	}

	// Cache hit, return cached traffic data
	fmt.Println("Cache hit.")
	return val
}

// GetTrafficData fetches the traffic data and formats it into GeoJSON
func GetTrafficData() string {
	// Define the API endpoint
	url := "https://verkehr.autobahn.de/o/autobahn/A1/services/warning"

	// Create an HTTP client and make a GET request to the API
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		log.Println("Error fetching traffic data:", err)
		return `{"error": "Failed to fetch traffic data"}`
	}
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK {
		log.Printf("Error: received non-200 status code: %d\n", resp.StatusCode)
		return `{"error": "Invalid response from traffic API"}`
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading response body:", err)
		return `{"error": "Failed to read traffic data"}`
	}

	// Convert the raw traffic data into a structured format
	var rawTrafficData map[string][]map[string]interface{}
	if err := json.Unmarshal(body, &rawTrafficData); err != nil {
		log.Println("Error parsing JSON response:", err)
		return `{"error": "Failed to parse traffic data"}`
	}

	// Prepare GeoJSON format
	geoJSON := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": []map[string]interface{}{},
	}

	// Loop through the traffic data and format it as GeoJSON features
	for _, item := range rawTrafficData["warning"] {
		// Extract point coordinates
		pointCoordinates := item["point"].(string) // Example: "49.89161871011858,6.851523798786332"
		var lat, lon float64
		fmt.Sscanf(pointCoordinates, "%f,%f", &lat, &lon)

		// Extract LineString coordinates
		geometry, ok := item["geometry"].(map[string]interface{})
		if !ok {
			continue
		}
		lineCoordinates := [][]float64{}
		for _, coord := range geometry["coordinates"].([]interface{}) {
			coordPair := coord.([]interface{})
			longitude := coordPair[0].(float64)
			latitude := coordPair[1].(float64)
			lineCoordinates = append(lineCoordinates, []float64{longitude, latitude})
		}

		// Create a LineString feature
		lineFeature := map[string]interface{}{
			"type": "Feature",
			"geometry": map[string]interface{}{
				"type":        "LineString",
				"coordinates": lineCoordinates,
			},
			"properties": map[string]interface{}{
				"title":               item["title"],
				"subtitle":            item["subtitle"],
				"abnormalTrafficType": item["abnormalTrafficType"],
				"averageSpeed":        item["averageSpeed"],
				"startTimestamp":      item["startTimestamp"],
				"description":         item["description"],
			},
		}

		// Add the LineString feature to the features list
		geoJSON["features"] = append(geoJSON["features"].([]map[string]interface{}), lineFeature)

		// Create a Point feature
		pointFeature := map[string]interface{}{
			"type": "Feature",
			"geometry": map[string]interface{}{
				"type":        "Point",
				"coordinates": []float64{lon, lat},
			},
			"properties": map[string]interface{}{
				"title":               item["title"],
				"subtitle":            item["subtitle"],
				"abnormalTrafficType": item["abnormalTrafficType"],
				"averageSpeed":        item["averageSpeed"],
				"startTimestamp":      item["startTimestamp"],
				"description":         item["description"],
			},
		}

		// Add the Point feature to the features list
		geoJSON["features"] = append(geoJSON["features"].([]map[string]interface{}), pointFeature)
	}

	// Convert the GeoJSON structure to a JSON string
	geoJSONBytes, err := json.Marshal(geoJSON)
	if err != nil {
		log.Println("Error marshalling GeoJSON:", err)
		return `{"error": "Failed to generate GeoJSON"}`
	}

	return string(geoJSONBytes)
}
