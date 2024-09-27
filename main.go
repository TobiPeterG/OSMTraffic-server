package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client

func main() {
	// Initialize Redis client
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
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
	// Fetch traffic data (later you'll integrate Datex2 API here)
	trafficData := FetchTrafficData()

	// Send the traffic data to the client
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(trafficData))
}

// FetchTrafficData fetches traffic data (will include Datex2 integration)
func FetchTrafficData() string {
	ctx := context.Background()

	// Try to get data from Redis cache
	val, err := rdb.Get(ctx, "traffic_data").Result()
	if err == redis.Nil {
		// Cache miss, so fetch from the data source (Datex2)
		fmt.Println("Cache miss. Fetching from Datex2...")
		trafficData := GetTrafficFromDatex2()

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

// Placeholder function for Datex2 integration
func GetTrafficFromDatex2() string {
	return `{"traffic": "sample data"}`
}
