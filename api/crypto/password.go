package crypto

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// HashPassword utilizes the bcrypt algorithm to transform plaintext into hashes
// which are safe to save in the database.
func HashPassword(pass string) (string, error) {
	passBytes, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	hashedPass := string(passBytes)
	return hashedPass, nil
}

// ComparePassword takes a hash and a plaintext password and determines whether they match.
// Returns nil if they do, error if they don't.
func ComparePassword(hash, pass string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(pass))
	if err != nil {
		return fmt.Errorf("compare password: %w", err)
	}
	return nil
}
