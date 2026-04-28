package auth

import (
	"golang.org/x/crypto/bcrypt"
)

const (
	// Cost for bcrypt hashing (higher = more secure but slower)
	bcryptCost = 12
)

// HashPassword hashes a plain text password using bcrypt
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword compares a plain text password with a hashed password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}
