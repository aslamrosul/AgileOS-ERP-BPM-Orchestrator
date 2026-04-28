package auth

import (
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents JWT claims structure
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

var (
	// Get JWT secret from environment or use default (CHANGE IN PRODUCTION!)
	jwtSecret = []byte(getEnv("JWT_SECRET", "agileos-super-secret-key-change-in-production"))
	
	// Token expiration times
	accessTokenExpiry  = 24 * time.Hour      // 24 hours
	refreshTokenExpiry = 7 * 24 * time.Hour  // 7 days
)

// GenerateJWT generates a new JWT token for a user
func GenerateJWT(userID, username, role, email string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "agileos-bpm",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GenerateRefreshToken generates a refresh token with longer expiry
func GenerateRefreshToken(userID, username, role, email string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExpiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "agileos-bpm",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshAccessToken generates a new access token from a valid refresh token
func RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := ValidateToken(refreshToken)
	if err != nil {
		return "", err
	}

	// Generate new access token
	return GenerateJWT(claims.UserID, claims.Username, claims.Role, claims.Email)
}

// Helper function to get environment variable with fallback
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
