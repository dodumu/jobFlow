package utils

import (
	"testing"
)

func TestGenerateResetAuthorizationToken(t *testing.T) {
	token, err := GenerateResetAuthorizationToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	if len(token) != 64 {
		t.Fatalf("expected 64 characters, got %d", len(token))
	}

	anotherToken, err := GenerateResetAuthorizationToken()
	if err != nil {
		t.Fatalf("failed to generate second token: %v", err)
	}

	if token == anotherToken {
		t.Fatal("expected different tokens, but got identical values")
	}
}
