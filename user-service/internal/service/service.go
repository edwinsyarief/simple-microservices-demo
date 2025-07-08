package service

import (
	"fmt"

	"user-service/internal/model"
	"user-service/internal/repository"
)

// UserService defines the business logic for user management.
// It interacts with the UserRepository interface.
type UserService struct {
	repo repository.UserRepository
}

// NewUserService creates a new instance of UserService.
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// CreateUser handles the creation of a new user.
// It performs basic validation and calls the repository to persist the user.
func (s *UserService) CreateUser(name string) (*model.User, error) {
	if name == "" {
		return nil, fmt.Errorf("user name cannot be empty")
	}
	// Additional business logic/validation can be added here
	return s.repo.CreateUser(name)
}

// GetAllUsers retrieves all users with pagination.
func (s *UserService) GetAllUsers(page, pageSize int) ([]model.User, error) {
	// Business logic for pagination defaults or limits can be applied here
	return s.repo.GetAllUsers(page, pageSize)
}

// GetUserByID retrieves a user by their ID.
func (s *UserService) GetUserByID(id int64) (*model.User, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid user ID: %d", id)
	}
	return s.repo.GetUserByID(id)
}
