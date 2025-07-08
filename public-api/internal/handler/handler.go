package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"public-api-layer/internal/client"
)

// PublicAPIHandler handles public-facing HTTP requests.
type PublicAPIHandler struct {
	userServiceClient    *client.UserServiceClient
	listingServiceClient *client.ListingServiceClient
}

// NewPublicAPIHandler creates a new instance of PublicAPIHandler.
func NewPublicAPIHandler(
	userServiceClient *client.UserServiceClient,
	listingServiceClient *client.ListingServiceClient,
) *PublicAPIHandler {
	return &PublicAPIHandler{
		userServiceClient:    userServiceClient,
		listingServiceClient: listingServiceClient,
	}
}

// PublicUserResponse represents the structure for public user creation response.
type PublicUserResponse struct {
	User *client.User `json:"user"`
}

// PublicListing represents a listing with embedded user information for public API.
type PublicListing struct {
	ID          int64        `json:"id"`
	ListingType string       `json:"listing_type"`
	Price       int64        `json:"price"`
	CreatedAt   int64        `json:"created_at"`
	UpdatedAt   int64        `json:"updated_at"`
	User        *client.User `json:"user"` // Embedded user object
}

// PublicListingsResponse represents the structure for public listings response.
type PublicListingsResponse struct {
	Result   bool            `json:"result"`
	Listings []PublicListing `json:"listings"`
	Error    string          `json:"error,omitempty"`
}

// CreatePublicUser handles POST /public-api/users requests.
// It proxies the request to the internal User Service.
func (h *PublicAPIHandler) CreatePublicUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Request body for public API is JSON
	var requestBody struct {
		Name string `json:"name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	if requestBody.Name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User name is required"})
		return
	}

	user, err := h.userServiceClient.CreateUser(requestBody.Name)
	if err != nil {
		log.Printf("Error creating user via User Service: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create user"})
		return
	}

	json.NewEncoder(w).Encode(PublicUserResponse{User: user})
}

// CreatePublicListing handles POST /public-api/listings requests.
// It proxies the request to the internal Listing Service.
func (h *PublicAPIHandler) CreatePublicListing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Request body for public API is JSON
	var requestBody struct {
		UserID      int64  `json:"user_id"`
		ListingType string `json:"listing_type"`
		Price       int64  `json:"price"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	// Basic validation for required fields
	if requestBody.UserID == 0 || requestBody.ListingType == "" || requestBody.Price <= 0 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "User ID, listing type, and price are required and valid"})
		return
	}
	if requestBody.ListingType != "rent" && requestBody.ListingType != "sale" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Listing type must be 'rent' or 'sale'"})
		return
	}

	listing, err := h.listingServiceClient.CreateListing(requestBody.UserID, requestBody.ListingType, requestBody.Price)
	if err != nil {
		log.Printf("Error creating listing via Listing Service: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Failed to create listing"})
		return
	}

	// The public API response format for create listing is just the listing object
	json.NewEncoder(w).Encode(map[string]*client.Listing{"listing": listing})
}

// GetPublicListings handles GET /public-api/listings requests.
// It aggregates data from Listing Service and User Service.
func (h *PublicAPIHandler) GetPublicListings(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse query parameters for pagination and user_id filter
	pageNumStr := r.URL.Query().Get("page_num")
	pageSizeStr := r.URL.Query().Get("page_size")
	userIDFilter := r.URL.Query().Get("user_id") // Optional user_id filter

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1 // Default
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10 // Default
	}

	// 1. Get listings from Listing Service
	listings, err := h.listingServiceClient.GetListings(pageNum, pageSize, userIDFilter)
	if err != nil {
		log.Printf("Error getting listings from Listing Service: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(PublicListingsResponse{Result: false, Error: "Failed to retrieve listings"})
		return
	}

	if len(listings) == 0 {
		json.NewEncoder(w).Encode(PublicListingsResponse{Result: true, Listings: []PublicListing{}})
		return
	}

	// 2. Extract unique user IDs from listings
	uniqueUserIDs := make(map[int64]struct{})
	for _, listing := range listings {
		uniqueUserIDs[listing.UserID] = struct{}{}
	}

	// 3. Concurrently fetch user details for unique user IDs
	userMap := make(map[int64]*client.User)
	var wg sync.WaitGroup
	var mu sync.Mutex // Mutex to protect userMap concurrent writes
	errorsChan := make(chan error, len(uniqueUserIDs))

	for userID := range uniqueUserIDs {
		wg.Add(1)
		go func(id int64) {
			defer wg.Done()
			user, err := h.userServiceClient.GetUserByID(id)
			if err != nil {
				// Log the error but don't fail the entire request if one user lookup fails
				log.Printf("Error fetching user %d from User Service: %v", id, err)
				errorsChan <- fmt.Errorf("failed to fetch user %d: %w", id, err)
				return
			}
			if user != nil {
				mu.Lock()
				userMap[id] = user
				mu.Unlock()
			}
		}(userID)
	}

	wg.Wait()         // Wait for all goroutines to complete
	close(errorsChan) // Close the channel after all goroutines are done

	// Check for any errors encountered during user fetching
	for err := range errorsChan {
		if err != nil {
			// Decide how to handle this:
			// Option 1: Return 500 if any user lookup fails (stricter)
			// log.Printf("Aggregate error during user fetching: %v", err)
			// w.WriteHeader(http.StatusInternalServerError)
			// json.NewEncoder(w).Encode(PublicListingsResponse{Result: false, Error: "Failed to retrieve all user details"})
			// return
			// Option 2: Continue, but listings without user data will have nil user (more resilient)
			// For this exercise, we'll proceed and let user be nil if not found/error.
		}
	}

	// 4. Aggregate listings with user details
	publicListings := make([]PublicListing, 0, len(listings))
	for _, listing := range listings {
		publicListing := PublicListing{
			ID:          listing.ID,
			ListingType: listing.ListingType,
			Price:       listing.Price,
			CreatedAt:   listing.CreatedAt,
			UpdatedAt:   listing.UpdatedAt,
			User:        userMap[listing.UserID], // Will be nil if user not found/error
		}
		publicListings = append(publicListings, publicListing)
	}

	json.NewEncoder(w).Encode(PublicListingsResponse{Result: true, Listings: publicListings})
}
