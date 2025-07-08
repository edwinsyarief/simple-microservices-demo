package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"user-service/internal/model"
	"user-service/internal/service"

	"github.com/gorilla/mux"
)

// UserHandler handles HTTP requests related to user operations.
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler creates a new instance of UserHandler.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Response structure for API responses.
type APIResponse struct {
	Result bool         `json:"result"`
	Users  []model.User `json:"users,omitempty"`
	User   *model.User  `json:"user,omitempty"`
	Error  string       `json:"error,omitempty"`
}

// GetAllUsers handles GET /users requests.
// It retrieves all users from the service, applying pagination if parameters are provided.
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse page_num and page_size from query parameters
	pageNumStr := r.URL.Query().Get("page_num")
	pageSizeStr := r.URL.Query().Get("page_size")

	pageNum, err := strconv.Atoi(pageNumStr)
	if err != nil || pageNum < 1 {
		pageNum = 1 // Default page number
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10 // Default page size
	}

	users, err := h.userService.GetAllUsers(pageNum, pageSize)
	if err != nil {
		log.Printf("Error getting all users: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Result: false, Error: "Internal server error"})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{Result: true, Users: users})
}

// GetUserByID handles GET /users/{id} requests.
// It retrieves a single user by their ID extracted from the URL path.
func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	idStr := vars["id"]

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Result: false, Error: "Invalid user ID format"})
		return
	}

	user, err := h.userService.GetUserByID(id)
	if err != nil {
		log.Printf("Error getting user by ID %d: %v", id, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Result: false, Error: "Internal server error"})
		return
	}

	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(APIResponse{Result: false, Error: "User not found"})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{Result: true, User: user})
}

// CreateUser handles POST /users requests.
// It parses form data to create a new user.
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Parse the form data for application/x-www-form-urlencoded
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Result: false, Error: "Failed to parse form data"})
		return
	}

	name := r.FormValue("name")
	if name == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(APIResponse{Result: false, Error: "User name is required"})
		return
	}

	user, err := h.userService.CreateUser(name)
	if err != nil {
		log.Printf("Error creating user with name '%s': %v", name, err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(APIResponse{Result: false, Error: "Internal server error"})
		return
	}

	json.NewEncoder(w).Encode(APIResponse{Result: true, User: user})
}
