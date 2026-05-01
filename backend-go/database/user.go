package database

import (
	"fmt"
	"log"
	"time"

	"agileos-backend/models"
)

// CreateUser creates a new user in the database
func (s *SurrealDB) CreateUser(user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true

	query := `CREATE user CONTENT $user`

	// Use intermediate struct to handle RecordID
	type UserWithRecordID struct {
		ID             models.RecordID `json:"id"`
		Username       string          `json:"username"`
		Email          string          `json:"email"`
		PasswordHash   string          `json:"password_hash,omitempty"`
		Role           string          `json:"role"`
		FullName       string          `json:"full_name"`
		Department     string          `json:"department,omitempty"`
		IsActive       bool            `json:"is_active"`
		CreatedAt      time.Time       `json:"created_at"`
		UpdatedAt      time.Time       `json:"updated_at"`
		LastLoginAt    *time.Time      `json:"last_login_at,omitempty"`
	}

	var created []UserWithRecordID
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"user": user}, &created); err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	if len(created) > 0 {
		user.ID = created[0].ID.String()
		log.Printf("✓ User created: %s (ID: %s)", user.Username, user.ID)
		return nil
	}

	return fmt.Errorf("user created but ID extraction failed")
}

// GetUserByUsername retrieves a user by username
func (s *SurrealDB) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT * FROM user WHERE username = $username LIMIT 1`

	// Use intermediate struct to handle RecordID
	type UserWithRecordID struct {
		ID             models.RecordID `json:"id"`
		Username       string          `json:"username"`
		Email          string          `json:"email"`
		PasswordHash   string          `json:"password_hash,omitempty"`
		Role           string          `json:"role"`
		FullName       string          `json:"full_name"`
		Department     string          `json:"department,omitempty"`
		IsActive       bool            `json:"is_active"`
		CreatedAt      time.Time       `json:"created_at"`
		UpdatedAt      time.Time       `json:"updated_at"`
		LastLoginAt    *time.Time      `json:"last_login_at,omitempty"`
	}

	var users []UserWithRecordID
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"username": username}, &users); err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	// Convert to models.User
	user := &models.User{
		ID:           users[0].ID.String(),
		Username:     users[0].Username,
		Email:        users[0].Email,
		PasswordHash: users[0].PasswordHash,
		Role:         users[0].Role,
		FullName:     users[0].FullName,
		Department:   users[0].Department,
		IsActive:     users[0].IsActive,
		CreatedAt:    users[0].CreatedAt,
		UpdatedAt:    users[0].UpdatedAt,
		LastLoginAt:  users[0].LastLoginAt,
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (s *SurrealDB) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT * FROM user WHERE email = $email LIMIT 1`

	// Use intermediate struct to handle RecordID
	type UserWithRecordID struct {
		ID             models.RecordID `json:"id"`
		Username       string          `json:"username"`
		Email          string          `json:"email"`
		PasswordHash   string          `json:"password_hash,omitempty"`
		Role           string          `json:"role"`
		FullName       string          `json:"full_name"`
		Department     string          `json:"department,omitempty"`
		IsActive       bool            `json:"is_active"`
		CreatedAt      time.Time       `json:"created_at"`
		UpdatedAt      time.Time       `json:"updated_at"`
		LastLoginAt    *time.Time      `json:"last_login_at,omitempty"`
	}

	var users []UserWithRecordID
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"email": email}, &users); err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	// Convert to models.User
	user := &models.User{
		ID:           users[0].ID.String(),
		Username:     users[0].Username,
		Email:        users[0].Email,
		PasswordHash: users[0].PasswordHash,
		Role:         users[0].Role,
		FullName:     users[0].FullName,
		Department:   users[0].Department,
		IsActive:     users[0].IsActive,
		CreatedAt:    users[0].CreatedAt,
		UpdatedAt:    users[0].UpdatedAt,
		LastLoginAt:  users[0].LastLoginAt,
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *SurrealDB) GetUserByID(userID string) (*models.User, error) {
	query := `SELECT * FROM $user`

	// Use intermediate struct to handle RecordID
	type UserWithRecordID struct {
		ID             models.RecordID `json:"id"`
		Username       string          `json:"username"`
		Email          string          `json:"email"`
		PasswordHash   string          `json:"password_hash,omitempty"`
		Role           string          `json:"role"`
		FullName       string          `json:"full_name"`
		Department     string          `json:"department,omitempty"`
		IsActive       bool            `json:"is_active"`
		CreatedAt      time.Time       `json:"created_at"`
		UpdatedAt      time.Time       `json:"updated_at"`
		LastLoginAt    *time.Time      `json:"last_login_at,omitempty"`
	}

	var users []UserWithRecordID
	if err := s.queryAndUnmarshal(query, map[string]interface{}{"user": userID}, &users); err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	// Convert to models.User
	user := &models.User{
		ID:           users[0].ID.String(),
		Username:     users[0].Username,
		Email:        users[0].Email,
		PasswordHash: users[0].PasswordHash,
		Role:         users[0].Role,
		FullName:     users[0].FullName,
		Department:   users[0].Department,
		IsActive:     users[0].IsActive,
		CreatedAt:    users[0].CreatedAt,
		UpdatedAt:    users[0].UpdatedAt,
		LastLoginAt:  users[0].LastLoginAt,
	}

	return user, nil
}

// UpdateUserLastLogin updates the last login timestamp
func (s *SurrealDB) UpdateUserLastLogin(userID string) error {
	now := time.Now()
	query := fmt.Sprintf(`UPDATE %s SET last_login_at = $timestamp`, userID)

	_, err := s.query(query, map[string]interface{}{
		"timestamp": now,
	})
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

// UpdateUser updates user information
func (s *SurrealDB) UpdateUser(user *models.User) error {
	user.UpdatedAt = time.Now()
	query := fmt.Sprintf(`UPDATE %s MERGE $user`, user.ID)

	_, err := s.query(query, map[string]interface{}{
		"user": user,
	})
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	log.Printf("✓ User updated: %s", user.Username)
	return nil
}

// DeactivateUser deactivates a user account
func (s *SurrealDB) DeactivateUser(userID string) error {
	query := fmt.Sprintf(`UPDATE %s SET is_active = false, updated_at = $timestamp`, userID)

	_, err := s.query(query, map[string]interface{}{
		"timestamp": time.Now(),
	})
	if err != nil {
		return fmt.Errorf("failed to deactivate user: %w", err)
	}

	log.Printf("✓ User deactivated: %s", userID)
	return nil
}

// ListUsers retrieves all users (admin only)
func (s *SurrealDB) ListUsers() ([]models.User, error) {
	query := `SELECT * FROM user ORDER BY created_at DESC`

	var users []models.User
	if err := s.queryAndUnmarshal(query, nil, &users); err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}
