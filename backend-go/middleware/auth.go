package middleware

import (
	"net/http"
	"strings"

	"agileos-backend/auth"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token from Authorization header
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header required",
			})
			c.Abort()
			return
		}

		// Check Bearer token format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Expected: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Invalid or expired token",
				"details": err.Error(),
			})
			c.Abort()
			return
		}

		// Store user info in context for use in handlers
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// AuthorizeRole checks if user has required role
func AuthorizeRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get role from context (set by AuthMiddleware)
		roleInterface, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "User role not found. Please authenticate first.",
			})
			c.Abort()
			return
		}

		userRole := roleInterface.(string)

		// Check if user role is in allowed roles
		allowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error":         "Access denied",
				"required_role": allowedRoles,
				"your_role":     userRole,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// OptionalAuth is like AuthMiddleware but doesn't abort if no token
// Useful for endpoints that work differently for authenticated users
func OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := auth.ValidateToken(tokenString)
		if err != nil {
			c.Next()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("email", claims.Email)
		c.Set("authenticated", true)

		c.Next()
	}
}
