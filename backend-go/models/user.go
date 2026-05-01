package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// User represents a user in the system
// RecordID represents a SurrealDB record ID that can be either a string or an object
type RecordID struct {
	TB string `json:"tb,omitempty"`
	ID string `json:"id,omitempty"`
}

// String returns the full record ID as a string
func (r RecordID) String() string {
	if r.TB != "" && r.ID != "" {
		return r.TB + ":" + r.ID
	}
	return r.ID
}

// UnmarshalJSON handles both string and object formats for record IDs
func (r *RecordID) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		r.ID = str
		return nil
	}
	
	// Try to unmarshal as object
	var obj struct {
		TB string `json:"tb"`
		ID string `json:"id"`
	}
	if err := json.Unmarshal(data, &obj); err == nil {
		r.TB = obj.TB
		r.ID = obj.ID
		return nil
	}
	
	return fmt.Errorf("invalid record ID format")
}

// MarshalJSON converts RecordID to JSON string
func (r RecordID) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.String())
}

type User struct {
	ID             string    `json:"id,omitempty"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"password_hash,omitempty"` // Never send to client
	Role           string    `json:"role"`                     // admin, manager, employee
	FullName       string    `json:"full_name"`
	Department     string    `json:"department,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	LastLoginAt    *time.Time `json:"last_login_at,omitempty"`
}

// UserRole defines available user roles
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleManager  UserRole = "manager"
	RoleEmployee UserRole = "employee"
	RoleFinance  UserRole = "finance"
	RoleProcurement UserRole = "procurement"
)

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents login response with tokens
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         UserInfo `json:"user"`
}

// UserInfo represents safe user information (without password)
type UserInfo struct {
	ID         string `json:"id"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Role       string `json:"role"`
	FullName   string `json:"full_name"`
	Department string `json:"department,omitempty"`
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Username   string `json:"username" binding:"required,min=3,max=50"`
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required,min=8"`
	FullName   string `json:"full_name" binding:"required"`
	Department string `json:"department"`
	Role       string `json:"role" binding:"required,oneof=admin manager employee finance procurement"`
}

// RefreshTokenRequest represents refresh token request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
