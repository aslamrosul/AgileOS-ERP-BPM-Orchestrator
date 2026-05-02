package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"agileos-backend/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
}

// TestAuthMiddleware_ValidToken tests authentication with valid JWT token
func TestAuthMiddleware_ValidToken(t *testing.T) {
	// Create a valid token
	claims := &auth.Claims{
		UserID:   "user123",
		Username: "testuser",
		Role:     "admin",
	}
	token, err := auth.GenerateJWT(claims)
	assert.NoError(t, err)

	// Setup test router
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "success",
			"user_id": c.GetString("user_id"),
		})
	})

	// Create request with valid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "user123")
}

// TestAuthMiddleware_MissingToken tests authentication without token
func TestAuthMiddleware_MissingToken(t *testing.T) {
	// Setup test router
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request without token
	req, _ := http.NewRequest("GET", "/protected", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authorization header required")
}

// TestAuthMiddleware_InvalidTokenFormat tests authentication with malformed token
func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	// Setup test router
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with invalid token format
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "InvalidFormat token123")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid authorization format")
}

// TestAuthMiddleware_ExpiredToken tests authentication with expired token
func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	// Create an expired token (set expiration to past)
	claims := &auth.Claims{
		UserID:   "user123",
		Username: "testuser",
		Role:     "admin",
	}
	
	// Generate token with custom expiration (expired)
	expiredToken, err := auth.GenerateJWTWithExpiration(claims, time.Now().Add(-1*time.Hour))
	assert.NoError(t, err)

	// Setup test router
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with expired token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+expiredToken)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// TestAuthMiddleware_InvalidSignature tests authentication with tampered token
func TestAuthMiddleware_InvalidSignature(t *testing.T) {
	// Create a token with invalid signature
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidXNlcjEyMyIsInVzZXJuYW1lIjoidGVzdHVzZXIiLCJyb2xlIjoiYWRtaW4ifQ.invalid_signature"

	// Setup test router
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with invalid token
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+invalidToken)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// TestAuthorizeRole_ValidRole tests role-based authorization with valid role
func TestAuthorizeRole_ValidRole(t *testing.T) {
	// Create a valid token with admin role
	claims := &auth.Claims{
		UserID:   "user123",
		Username: "testuser",
		Role:     "admin",
	}
	token, err := auth.GenerateJWT(claims)
	assert.NoError(t, err)

	// Setup test router
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/admin", AuthorizeRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Create request
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin access granted")
}

// TestAuthorizeRole_InvalidRole tests role-based authorization with insufficient role
func TestAuthorizeRole_InvalidRole(t *testing.T) {
	// Create a valid token with employee role
	claims := &auth.Claims{
		UserID:   "user123",
		Username: "testuser",
		Role:     "employee",
	}
	token, err := auth.GenerateJWT(claims)
	assert.NoError(t, err)

	// Setup test router
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/admin", AuthorizeRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "admin access granted"})
	})

	// Create request
	req, _ := http.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Insufficient permissions")
}

// TestAuthorizeRole_MultipleRoles tests role-based authorization with multiple allowed roles
func TestAuthorizeRole_MultipleRoles(t *testing.T) {
	tests := []struct {
		name           string
		userRole       string
		allowedRoles   []string
		expectedStatus int
	}{
		{
			name:           "Admin accessing manager endpoint",
			userRole:       "admin",
			allowedRoles:   []string{"admin", "manager"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Manager accessing manager endpoint",
			userRole:       "manager",
			allowedRoles:   []string{"admin", "manager"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Employee accessing manager endpoint",
			userRole:       "employee",
			allowedRoles:   []string{"admin", "manager"},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create token with specified role
			claims := &auth.Claims{
				UserID:   "user123",
				Username: "testuser",
				Role:     tt.userRole,
			}
			token, err := auth.GenerateJWT(claims)
			assert.NoError(t, err)

			// Setup test router
			router := gin.New()
			router.Use(AuthMiddleware())
			router.GET("/protected", AuthorizeRole(tt.allowedRoles...), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "access granted"})
			})

			// Create request
			req, _ := http.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", "Bearer "+token)

			// Execute request
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

// BenchmarkAuthMiddleware benchmarks the authentication middleware
func BenchmarkAuthMiddleware(b *testing.B) {
	// Create a valid token
	claims := &auth.Claims{
		UserID:   "user123",
		Username: "testuser",
		Role:     "admin",
	}
	token, _ := auth.GenerateJWT(claims)

	// Setup test router
	router := gin.New()
	router.Use(AuthMiddleware())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request
	req, _ := http.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}