package auth

import (
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func MakeJWT(userId uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userId.String(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		fmt.Printf("error  is: %s\n", err)
		return "", err
	}
	fmt.Printf("from jwt: %s\n", s)
	return s, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	emptyClaim := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, emptyClaim, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Printf("Problem: %s\n", err)
		return uuid.Nil, err
	} else if claims, ok := token.Claims.(*jwt.RegisteredClaims); ok {
		fmt.Println(claims.Issuer)
		id, err := uuid.Parse(claims.Subject)
		if err != nil {
			log.Printf("Problem: %s\n", err)
			return uuid.Nil, err
		}
		return id, nil
	} else {
		log.Println("problem with claim")
		return uuid.Nil, err
	}
}
