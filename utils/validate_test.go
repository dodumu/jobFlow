package utils

import (
	"strings"
	"testing"
)

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		email   string
		wantErr bool
	}{
		{"david@example.com", false},
		{"first.last@mail.example.co.ng", false},
		{"", true},
		{"david", true},
		{"david@", true},
		{"@example.com", true},
		{"david@localhost", true},
		{"David <david@example.com>", true},
	}

	for _, tc := range tests {
		t.Run(tc.email, func(t *testing.T) {
			err := ValidateEmail(tc.email)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tc.email, err, tc.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"good password", "correct7horse", false},
		{"too short", "abc123", true},
		{"letters only", "abcdefghij", true},
		{"digits only", "1234567890", true},
		{"exactly 72 bytes", strings.Repeat("a", 71) + "1", false},
		{"73 bytes", strings.Repeat("a", 72) + "1", true},
		{"empty", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidatePassword(tc.password)
			if (err != nil) != tc.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tc.password, err, tc.wantErr)
			}
		})
	}
}
