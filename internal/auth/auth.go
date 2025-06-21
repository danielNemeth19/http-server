package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)


func HashPassword(password string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	if err != nil {
		return "", fmt.Errorf("Hashing password failed: %v\n", err)
	}
	return string(b), nil
}

func CheckPasswordHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}
