package database_test

import (
	"crypto/rand"
	"encoding/hex"
	"path/filepath"
	"testing"
	"time"

	"jobFlow/database"
)

func setupTestDB(t *testing.T) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")

	if err := database.InitDB(dbPath); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	t.Cleanup(func() {
		if database.DB != nil {
			database.DB.Close()
			database.DB = nil
		}
	})
}

func setupTestUser(t *testing.T) int {
	t.Helper()

	result, err := database.DB.Exec(`
		INSERT INTO users (
			first_name,
			last_name,
			username,
			password_hash,
			email,
			date_of_birth,
			role
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		"Test",
		"User",
		"testuser",
		"fake-test-hash",
		"test@example.com",
		"2000-01-01",
		"individual",
	)

	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("failed to retrieve test user ID: %v", err)
	}

	return int(id)
}

func createTestAuthorization(t *testing.T, userID int) string {
	t.Helper()

	// Generate a random token locally without importing utils.
	tokenBytes := make([]byte, 32)

	if _, err := rand.Read(tokenBytes); err != nil {
		t.Fatalf("generating reset token: %v", err)
	}

	token := hex.EncodeToString(tokenBytes)
	tokenHash := database.HashResetAuthorizationToken(token)

	err := database.SavePasswordResetAuthorization(userID, tokenHash)
	if err != nil {
		t.Fatalf("saving reset authorization: %v", err)
	}

	return token
}

func TestSavePasswordResetAuthorization(t *testing.T) {
	setupTestDB(t)
	userID := setupTestUser(t)

	err := database.SavePasswordResetAuthorization(userID, "test-token-hash")
	if err != nil {
		t.Fatalf("failed to save authorization: %v", err)
	}

	var savedHash string
	var expiresAt time.Time

	err = database.DB.QueryRow(`
		SELECT token_hash, expires_at
		FROM password_reset_authorizations
		WHERE user_id = ?
	`, userID).Scan(&savedHash, &expiresAt)

	if err != nil {
		t.Fatalf("failed to retrieve saved authorization: %v", err)
	}

	if savedHash != "test-token-hash" {
		t.Errorf("expected hash %q, got %q", "test-token-hash", savedHash)
	}

	expectedExpiry := time.Now().UTC().Add(10 * time.Minute)

	if expiresAt.Before(expectedExpiry.Add(-time.Minute)) ||
		expiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Errorf("expiration time is not approximately 10 minutes from now: %v", expiresAt)
	}

	err = database.SavePasswordResetAuthorization(userID, "second-token-hash")
	if err != nil {
		t.Fatalf("failed to save second authorization: %v", err)
	}

	var count int

	err = database.DB.QueryRow(`
		SELECT COUNT(*)
		FROM password_reset_authorizations
		WHERE user_id = ?
	`, userID).Scan(&count)

	if err != nil {
		t.Fatalf("failed to count authorizations: %v", err)
	}

	if count != 1 {
		t.Fatalf("expected 1 authorization, got %d", count)
	}

	var currentHash string

	err = database.DB.QueryRow(`
		SELECT token_hash
		FROM password_reset_authorizations
		WHERE user_id = ?
	`, userID).Scan(&currentHash)

	if err != nil {
		t.Fatalf("failed to retrieve current authorization: %v", err)
	}

	if currentHash != "second-token-hash" {
		t.Errorf("expected new authorization hash, got %q", currentHash)
	}
}

func TestValidatePasswordResetAuthorization_Valid(t *testing.T) {
	setupTestDB(t)
	userID := setupTestUser(t)

	token := createTestAuthorization(t, userID)

	err := database.ValidatePasswordResetAuthorization(token)
	if err != nil {
		t.Fatalf("expected valid authorization, got error: %v", err)
	}
}

func TestValidatePasswordResetAuthorization_Expired(t *testing.T) {
	setupTestDB(t)
	userID := setupTestUser(t)

	token := createTestAuthorization(t, userID)
	tokenHash := database.HashResetAuthorizationToken(token)

	_, err := database.DB.Exec(`
		UPDATE password_reset_authorizations
		SET expires_at = ?
		WHERE token_hash = ?
	`, time.Now().UTC().Add(-time.Minute), tokenHash)

	if err != nil {
		t.Fatalf("making authorization expire: %v", err)
	}

	err = database.ValidatePasswordResetAuthorization(token)
	if err == nil {
		t.Fatal("expected expired authorization to fail")
	}
}

func TestValidatePasswordResetAuthorization_Used(t *testing.T) {
	setupTestDB(t)
	userID := setupTestUser(t)

	token := createTestAuthorization(t, userID)
	tokenHash := database.HashResetAuthorizationToken(token)

	_, err := database.DB.Exec(`
		UPDATE password_reset_authorizations
		SET used_at = ?
		WHERE token_hash = ?
	`, time.Now().UTC(), tokenHash)

	if err != nil {
		t.Fatalf("marking authorization as used: %v", err)
	}

	err = database.ValidatePasswordResetAuthorization(token)
	if err == nil {
		t.Fatal("expected used authorization to fail")
	}
}
