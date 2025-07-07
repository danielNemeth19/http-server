package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
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
		return "", err
	}
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

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("No auth header found")
	}
	token, isBearer := strings.CutPrefix(authHeader, "Bearer ")
	if !isBearer {
		return "", fmt.Errorf("No auth header found")
	}
	return token, nil
}

func MakereRefreshToken() (string, error) {
	randString := make([]byte, 32)
	_, err := rand.Read(randString)
	if err != nil {
		return "", fmt.Errorf("Error generating random data: %s\n", err)
	}
	endodedStr := hex.EncodeToString(randString)
	return endodedStr, nil

}
