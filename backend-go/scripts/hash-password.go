package main

import (
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"
)

// Simple utility to hash passwords for manual user creation
// Usage: go run scripts/hash-password.go <password>

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/hash-password.go <password>")
		fmt.Println("Example: go run scripts/hash-password.go mySecurePassword123")
		os.Exit(1)
	}

	password := os.Args[1]

	// Hash password with bcrypt cost 12
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		fmt.Printf("Error hashing password: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Password hashed successfully!")
	fmt.Println("---")
	fmt.Printf("Plain text: %s\n", password)
	fmt.Printf("Hashed:     %s\n", string(hashedBytes))
	fmt.Println("---")
	fmt.Println("Use this hash in your database or seed script.")
}
