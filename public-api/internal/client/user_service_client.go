package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// User represents the user entity for inter-service communication.
// Note: This model should ideally be shared or a common contract defined.
type User struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// UserServiceResponse is the expected structure for User Service API responses.
type UserServiceResponse struct {
	Result bool   `json:"result"`
	Users  []User `json:"users,omitempty"`
	User   *User  `json:"user,omitempty"`
	Error  string `json:"error,omitempty"`
}

// UserServiceClient handles communication with the User Service.
type UserServiceClient struct {
	httpClient *http.Client
	baseURL    string
}

// NewUserServiceClient creates a new UserServiceClient.
func NewUserServiceClient(httpClient *http.Client, baseURL string) *UserServiceClient {
	return &UserServiceClient{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// CreateUser sends a POST request to the User Service to create a new user.
func (c *UserServiceClient) CreateUser(name string) (*User, error) {
	// Prepare the form data for application/x-www-form-urlencoded
	formData := url.Values{}
	formData.Set("name", name)

	req, err := http.NewRequest("POST", c.baseURL+"/users", bytes.NewBufferString(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request to User Service: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to User Service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("User Service returned non-OK status: %s", resp.Status)
	}

	var apiResp UserServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode User Service response: %w", err)
	}

	if !apiResp.Result {
		return nil, fmt.Errorf("User Service reported error: %s", apiResp.Error)
	}

	return apiResp.User, nil
}

// GetUserByID sends a GET request to the User Service to retrieve a user by ID.
func (c *UserServiceClient) GetUserByID(id int64) (*User, error) {
	url := fmt.Sprintf("%s/users/%d", c.baseURL, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to User Service: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request to User Service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil // User not found, return nil user and nil error
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("User Service returned non-OK status: %s", resp.Status)
	}

	var apiResp UserServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode User Service response: %w", err)
	}

	if !apiResp.Result {
		return nil, fmt.Errorf("User Service reported error: %s", apiResp.Error)
	}

	return apiResp.User, nil
}
