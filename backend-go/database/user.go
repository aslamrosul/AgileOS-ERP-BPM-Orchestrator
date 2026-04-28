package database

import (
	"fmt"
	"log"
	"time"

	"agileos-backend/models"

	"github.com/surrealdb/surrealdb.go"
)

// CreateUser creates a new user in the database
func (s *SurrealDB) CreateUser(user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.IsActive = true

	query := `CREATE user CONTENT $user`

	result, err := s.client.Query(query, map[string]interface{}{
		"user": user,
	})
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	// Extract ID from result
	if resultArray, ok := result.([]interface{}); ok && len(resultArray) > 0 {
		if outerMap, ok := resultArray[0].(map[string]interface{}); ok {
			if resultField, ok := outerMap["result"].([]interface{}); ok && len(resultField) > 0 {
				if innerMap, ok := resultField[0].(map[string]interface{}); ok {
					if id, ok := innerMap["id"].(string); ok {
						user.ID = id
						log.Printf("✓ User created: %s (ID: %s)", user.Username, user.ID)
						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("user created but ID extraction failed")
}

// GetUserByUsername retrieves a user by username
func (s *SurrealDB) GetUserByUsername(username string) (*models.User, error) {
	query := `SELECT * FROM user WHERE username = $username LIMIT 1`

	result, err := s.client.Query(query, map[string]interface{}{
		"username": username,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var users []models.User
	if err := surrealdb.Unmarshal(result, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &users[0], nil
}

// GetUserByEmail retrieves a user by email
func (s *SurrealDB) GetUserByEmail(email string) (*models.User, error) {
	query := `SELECT * FROM user WHERE email = $email LIMIT 1`

	result, err := s.client.Query(query, map[string]interface{}{
		"email": email,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var users []models.User
	if err := surrealdb.Unmarshal(result, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &users[0], nil
}

// GetUserByID retrieves a user by ID
func (s *SurrealDB) GetUserByID(userID string) (*models.User, error) {
	query := `SELECT * FROM $user`

	result, err := s.client.Query(query, map[string]interface{}{
		"user": userID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	var users []models.User
	if err := surrealdb.Unmarshal(result, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	return &users[0], nil
}

// UpdateUserLastLogin updates the last login timestamp
func (s *SurrealDB) UpdateUserLastLogin(userID string) error {
	now := time.Now()
	query := fmt.Sprintf(`UPDATE %s SET last_login_at = $timestamp`, userID)

	_, err := s.client.Query(query, map[string]interface{}{
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

	_, err := s.client.Query(query, map[string]interface{}{
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

	_, err := s.client.Query(query, map[string]interface{}{
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

	result, err := s.client.Query(query, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	var users []models.User
	if err := surrealdb.Unmarshal(result, &users); err != nil {
		return nil, fmt.Errorf("failed to unmarshal users: %w", err)
	}

	return users, nil
}
