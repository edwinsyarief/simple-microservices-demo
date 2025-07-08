package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"user-service/internal/handler"
	"user-service/internal/repository"
	"user-service/internal/service"

	"github.com/gorilla/mux"
	_ "github.com/mattn/go-sqlite3" // Import for SQLite driver
)

func main() {
	// Define command-line flags for port and debug mode
	port := flag.Int("port", 7000, "The port number to run the User Service on")
	debug := flag.Bool("debug", true, "Runs the application in debug mode (currently no effect on auto-reload)")
	flag.Parse()

	// Initialize the SQLite database
	// This will create 'users.db' in the current directory if it doesn't exist.
	db, err := repository.NewSQLiteDB("users.db")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	// Initialize repository, service, and handler layers
	userRepo := repository.NewSQLiteUserRepository(db)
	userService := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userService)

	// Create a new Gorilla Mux router
	r := mux.NewRouter()

	// Define User Service API routes
	// GET /users: Get all users with pagination
	r.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	// GET /users/{id}: Get a specific user by ID
	r.HandleFunc("/users/{id}", userHandler.GetUserByID).Methods("GET")
	// POST /users: Create a new user
	r.HandleFunc("/users", userHandler.CreateUser).Methods("POST")

	// Configure HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", *port),
		Handler:      r,
		ReadTimeout:  15 * time.Second, // Max time to read request from client
		WriteTimeout: 15 * time.Second, // Max time to write response to client
		IdleTimeout:  60 * time.Second, // Max time for connections to remain idle
	}

	// Start the HTTP server
	log.Printf("User Service starting on port %d (Debug mode: %t)", *port, *debug)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Could not listen on port %d: %v", *port, err)
	}
}
