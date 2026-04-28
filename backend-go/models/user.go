package models

import "time"

// User represents a user in the system
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
