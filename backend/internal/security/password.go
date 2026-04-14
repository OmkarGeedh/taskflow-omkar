package security

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const DefaultCost = 12

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), DefaultCost)
	if err != nil {
		return "", fmt.Errorf("Failed to hash password %w", err)
	}
	return string(hashedBytes), nil
}

func VerifyPassword(hashedPassword, plainPassword string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
}
