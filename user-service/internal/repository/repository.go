package repository

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"user-service/internal/model"
)

// UserRepository defines the interface for user data operations.
// This abstraction allows for different database implementations (e.g., SQLite, PostgreSQL)
// without changing the service layer logic.
type UserRepository interface {
	CreateUser(name string) (*model.User, error)
	GetAllUsers(page, pageSize int) ([]model.User, error)
	GetUserByID(id int64) (*model.User, error)
}

// sqliteUserRepository implements UserRepository for SQLite database.
type sqliteUserRepository struct {
	db *sql.DB
}

// NewSQLiteDB initializes and returns a new SQLite database connection.
// It also ensures the 'users' table exists, creating it if necessary.
func NewSQLiteDB(dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings for better performance and resource management
	db.SetMaxOpenConns(10)                 // Max number of open connections
	db.SetMaxIdleConns(5)                  // Max number of idle connections
	db.SetConnMaxLifetime(5 * time.Minute) // Max time a connection can be reused

	// Ping the database to verify connection
	if err = db.Ping(); err != nil {
		db.Close() // Close the connection if ping fails
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Create the users table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		created_at INTEGER NOT NULL,
		updated_at INTEGER NOT NULL
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create users table: %w", err)
	}

	log.Printf("SQLite database '%s' initialized successfully.", dataSourceName)
	return db, nil
}

// NewSQLiteUserRepository creates a new instance of sqliteUserRepository.
func NewSQLiteUserRepository(db *sql.DB) UserRepository {
	return &sqliteUserRepository{db: db}
}

// CreateUser inserts a new user into the database.
// It generates current timestamps in microseconds for created_at and updated_at.
func (r *sqliteUserRepository) CreateUser(name string) (*model.User, error) {
	stmt, err := r.db.Prepare("INSERT INTO users(name, created_at, updated_at) VALUES(?, ?, ?)")
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement for creating user: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil {
			log.Printf("Error closing statement: %v", err)
		}
	}()

	now := time.Now().UnixMicro() // Get current time in microseconds
	result, err := stmt.Exec(name, now, now)
	if err != nil {
		return nil, fmt.Errorf("failed to execute statement for creating user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID after creating user: %w", err)
	}

	return &model.User{
		ID:        id,
		Name:      name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetAllUsers retrieves all users from the database with pagination.
// Results are sorted by 'created_at' in descending order.
func (r *sqliteUserRepository) GetAllUsers(page, pageSize int) ([]model.User, error) {
	// Ensure page and pageSize are positive
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10 // Default page size
	}

	offset := (page - 1) * pageSize
	query := `SELECT id, name, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT ? OFFSET ?`
	rows, err := r.db.Query(query, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query all users: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("Error closing rows: %v", err)
		}
	}()

	var users []model.User
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Name, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration for GetAllUsers: %w", err)
	}

	return users, nil
}

// GetUserByID retrieves a single user by their ID.
func (r *sqliteUserRepository) GetUserByID(id int64) (*model.User, error) {
	query := `SELECT id, name, created_at, updated_at FROM users WHERE id = ?`
	row := r.db.QueryRow(query, id)

	var user model.User
	err := row.Scan(&user.ID, &user.Name, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // User not found
		}
		return nil, fmt.Errorf("failed to scan user by ID: %w", err)
	}
	return &user, nil
}
