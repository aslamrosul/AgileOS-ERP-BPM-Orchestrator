package handlers

import (
	"log"
	"net/http"

	"agileos-backend/auth"
	"agileos-backend/database"
	"agileos-backend/models"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	db *database.SurrealDB
}

func NewAuthHandler(db *database.SurrealDB) *AuthHandler {
	return &AuthHandler{db: db}
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with username and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse "Successfully authenticated"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 401 {object} map[string]string "Invalid username or password"
// @Failure 403 {object} map[string]string "Account deactivated"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Login attempt for username: %s", req.Username)

	// Get user from database
	user, err := h.db.GetUserByUsername(req.Username)
	if err != nil {
		log.Printf("Failed to get user '%s': %v", req.Username, err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	log.Printf("User found: %s (ID: %s, PasswordHash length: %d)", user.Username, user.ID, len(user.PasswordHash))

	log.Printf("User found: %s (ID: %s, PasswordHash length: %d)", user.Username, user.ID, len(user.PasswordHash))

	// Check if user is active
	if !user.IsActive {
		log.Printf("User %s is not active", user.Username)
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Account is deactivated. Please contact administrator.",
		})
		return
	}

	// Verify password
	log.Printf("Verifying password for user: %s", user.Username)
	if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
		log.Printf("Password verification failed for user %s: %v", user.Username, err)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid username or password",
		})
		return
	}

	log.Printf("Password verified successfully for user: %s", user.Username)

	// Generate tokens
	accessToken, err := auth.GenerateJWT(user.ID, user.Username, user.Role, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate access token",
		})
		return
	}

	refreshToken, err := auth.GenerateRefreshToken(user.ID, user.Username, user.Role, user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate refresh token",
		})
		return
	}

	// Update last login
	h.db.UpdateUserLastLogin(user.ID)

	// Return response
	c.JSON(http.StatusOK, models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: models.UserInfo{
			ID:         user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Role:       user.Role,
			FullName:   user.FullName,
			Department: user.Department,
		},
	})

	log.Printf("✓ User logged in: %s (Role: %s)", user.Username, user.Role)
}

// Register handles user registration
// @Summary Register new user
// @Description Create a new user account
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.RegisterRequest true "User registration data"
// @Success 201 {object} models.LoginResponse "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request body"
// @Failure 409 {object} map[string]string "Username or email already exists"
// @Failure 500 {object} map[string]string "Failed to create user"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if username already exists
	if _, err := h.db.GetUserByUsername(req.Username); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Username already exists",
		})
		return
	}

	// Check if email already exists
	if _, err := h.db.GetUserByEmail(req.Email); err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email already exists",
		})
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	// Create user
	user := &models.User{
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		FullName:     req.FullName,
		Department:   req.Department,
	}

	if err := h.db.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	// Generate tokens
	accessToken, _ := auth.GenerateJWT(user.ID, user.Username, user.Role, user.Email)
	refreshToken, _ := auth.GenerateRefreshToken(user.ID, user.Username, user.Role, user.Email)

	c.JSON(http.StatusCreated, models.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: models.UserInfo{
			ID:         user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Role:       user.Role,
			FullName:   user.FullName,
			Department: user.Department,
		},
	})

	log.Printf("✓ User registered: %s (Role: %s)", user.Username, user.Role)
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate new access token from refresh token
	accessToken, err := auth.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid or expired refresh token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": accessToken,
	})
}

// GetProfile returns current user profile
// @Summary Get user profile
// @Description Get current authenticated user's profile information
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.UserInfo "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Router /auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, models.UserInfo{
		ID:         user.ID,
		Username:   user.Username,
		Email:      user.Email,
		Role:       user.Role,
		FullName:   user.FullName,
		Department: user.Department,
	})
}

// ListUsers returns all users (admin only)
func (h *AuthHandler) ListUsers(c *gin.Context) {
	users, err := h.db.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve users",
		})
		return
	}

	// Remove password hashes from response
	userInfos := make([]models.UserInfo, len(users))
	for i, user := range users {
		userInfos[i] = models.UserInfo{
			ID:         user.ID,
			Username:   user.Username,
			Email:      user.Email,
			Role:       user.Role,
			FullName:   user.FullName,
			Department: user.Department,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"users": userInfos,
		"count": len(userInfos),
	})
}
