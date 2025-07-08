package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// Listing represents the listing entity for inter-service communication.
// Note: This model should ideally be shared or a common contract defined.
type Listing struct {
	ID          int64  `json:"id"`
	UserID      int64  `json:"user_id"`
	ListingType string `json:"listing_type"`
	Price       int64  `json:"price"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// ListingServiceResponse is the expected structure for Listing Service API responses.
type ListingServiceResponse struct {
	Result   bool      `json:"result"`
	Listings []Listing `json:"listings,omitempty"`
	Listing  *Listing  `json:"listing,omitempty"`
	Error    string    `json:"error,omitempty"`
}

// ListingServiceClient handles communication with the Listing Service.
type ListingServiceClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewListingServiceClient creates a new ListingServiceClient.
func NewListingServiceClient(httpClient *http.Client, baseURL string) *ListingServiceClient {
	return &ListingServiceClient{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// CreateListing sends a POST request to the Listing Service to create a new listing.
func (c *ListingServiceClient) CreateListing(userID int64, listingType string, price int64) (*Listing, error) {
	// Prepare the form data for application/x-www-form-urlencoded
	formData := url.Values{}
	formData.Set("user_id", strconv.FormatInt(userID, 10))
	formData.Set("listing_type", listingType)
	formData.Set("price", strconv.FormatInt(price, 10))

	req, err := http.NewRequest("POST", c.baseURL+"/listings", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request to Listing Service: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Listing Service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Listing Service returned non-OK status: %s", resp.Status)
	}

	var apiResp ListingServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode Listing Service response: %w", err)
	}

	if !apiResp.Result {
		return nil, fmt.Errorf("Listing Service reported error: %s", apiResp.Error)
	}

	return apiResp.Listing, nil
}

// GetListings sends a GET request to the Listing Service to retrieve listings.
func (c *ListingServiceClient) GetListings(pageNum, pageSize int, userID string) ([]Listing, error) {
	// Build query parameters
	params := url.Values{}
	params.Set("page_num", strconv.Itoa(pageNum))
	params.Set("page_size", strconv.Itoa(pageSize))
	if userID != "" {
		params.Set("user_id", userID)
	}

	requestURL := fmt.Sprintf("%s/listings?%s", c.baseURL, params.Encode())

	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to Listing Service: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Listing Service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Listing Service returned non-OK status: %s", resp.Status)
	}

	var apiResp ListingServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode Listing Service response: %w", err)
	}

	if !apiResp.Result {
		return nil, fmt.Errorf("Listing Service reported error: %s", apiResp.Error)
	}

	return apiResp.Listings, nil
}
