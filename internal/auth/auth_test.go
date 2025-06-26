package auth

import (
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
