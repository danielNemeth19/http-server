package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestMakeJWT(t *testing.T) {
	u := uuid.New()
	s, err := MakeJWT(u, "test", time.Hour)
	if len(s) == 0 {
		t.Errorf("Token should have been generated")
	}
	if err != nil {
		t.Errorf("Error shouldn't have been raised")
	}
}

func TestValidateJWT(t *testing.T) {
	u := uuid.New()
	s, _ := MakeJWT(u, "signKey", time.Hour)
	uFromToken, _ := ValidateJWT(s, "signKey")
	if u != uFromToken {
		t.Errorf("Validate failed")
	}
}

func TestValidateJWTFailsExpiredToken(t *testing.T) {
	u := uuid.New()
	s, _ := MakeJWT(u, "signKey", 1)
	_, err := ValidateJWT(s, "signKey")
	if err == nil {
		t.Errorf("Error should have been raised as token expired: %s\n", err)
	}
}

func TestValidateJWTFailsIncorrectSecret(t *testing.T) {
	u := uuid.New()
	s, _ := MakeJWT(u, "signKey", 1)
	_, err := ValidateJWT(s, "thisWasNotTheOne")
	if err == nil {
		t.Errorf("Error should have been raised as token secret incorrect: %s\n", err)
	}
}

func TestGetBearerTokenFailsIncorrectHeader(t *testing.T) {
	header := http.Header{
		"Irrelevant": []string{"barmi"},
	}
	_, err := GetBearerToken(header)
	if err == nil {
		t.Errorf("Error should have been raised as Authorization header missing: %s\n", err)
	}
}

func TestGetBearerTokenInvalidBearerToken(t *testing.T) {
	header := http.Header{
		"Authorization": []string{"invalid"},
	}
	_, err := GetBearerToken(header)
	if err == nil {
		t.Errorf("Error should have been raised as string not a bearer token: %s\n", err)
	}
}

func TestGetBearerToken(t *testing.T) {
	header := http.Header{
		"Authorization": []string{"Bearer testToken"},
	}
	token, _ := GetBearerToken(header)
	if token != "testToken" {
		t.Errorf("Token should have been parsed, got: %s\n", token)
	}
}

func TestMakeRefreshToken(t *testing.T) {
	s, _ := MakereRefreshToken()
	if len(s) != 64 {
		t.Errorf("Len of refresh token should have been 64, got %d\n", len(s))
	}
}
