package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"public-api-layer/internal/client"
	"public-api-layer/internal/handler"

	"github.com/gorilla/mux"
)

func main() {
	// Define command-line flags for port and service URLs
	port := flag.Int("port", 8000, "The port number to run the Public API Layer on")
	userServiceURL := flag.String("user-service-url", "http://localhost:7000", "URL of the User Service")
	listingServiceURL := flag.String("listing-service-url", "http://localhost:6000", "URL of the Listing Service")
	flag.Parse()

	// Initialize a custom HTTP client with timeouts for inter-service communication
	// This is crucial for resilience and preventing resource exhaustion.
	httpClient := client.NewHTTPClient(
		10*time.Second, // Overall request timeout
		5*time.Second,  // Dial timeout
		5*time.Second,  // TLS handshake timeout
		5*time.Second,  // Response header timeout
	)

	// Initialize service clients
	userServiceClient := client.NewUserServiceClient(httpClient, *userServiceURL)
	listingServiceClient := client.NewListingServiceClient(httpClient, *listingServiceURL)

	// Initialize the Public API handler
	publicAPIHandler := handler.NewPublicAPIHandler(userServiceClient, listingServiceClient)

	// Create a new Gorilla Mux router
	r := mux.NewRouter()

	// Define Public API Layer routes
	// GET /public-api/listings: Get all listings, enriched with user data
	r.HandleFunc("/public-api/listings", publicAPIHandler.GetPublicListings).Methods("GET")
	// POST /public-api/users: Create a new user
	r.HandleFunc("/public-api/users", publicAPIHandler.CreatePublicUser).Methods("POST")
	// POST /public-api/listings: Create a new listing
	r.HandleFunc("/public-api/listings", publicAPIHandler.CreatePublicListing).Methods("POST")

	// Configure HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      r,
		ReadTimeout:  15 * time.Second, // Max time to read request from client
		WriteTimeout: 15 * time.Second, // Max time to write response to client
		IdleTimeout:  60 * time.Second, // Max time for connections to remain idle
	}

	// Start the HTTP server
	log.Printf("Public API Layer starting on port %d", *port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on port %d: %v", *port, err)
	}
}
