package database

import (
	"path/filepath"
	"testing"
	"time"

	"jobFlow/utils"
)

func TestSavePasswordResetAuthorization(t *testing.T) {
	// Create a temporary database file for this test.
	dbPath := filepath.Join(t.TempDir(), "test.db")

	// Initialize the database.
	if err := InitDB(dbPath); err != nil {
		t.Fatalf("failed to initialize test database: %v", err)
	}

	// Close the database connection when the test finishes.
	t.Cleanup(func() {
		if DB != nil {
			DB.Close()
			DB = nil
		}
	})
	_, err := DB.Exec(`
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
	var userID int

	err = DB.QueryRow(`
    SELECT id FROM users WHERE username = ?
`, "testuser").Scan(&userID)

	if err != nil {
		t.Fatalf("failed to get test user ID: %v", err)
	}
	// Try saving an authorization.
	err = SavePasswordResetAuthorization(1, "test-token-hash")
	if err != nil {
		t.Fatalf("failed to save authorization: %v", err)
	}

	var savedHash string
	var expiresAt time.Time

	err = DB.QueryRow(`
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

	if expiresAt.Before(expectedExpiry.Add(-1*time.Minute)) ||
		expiresAt.After(expectedExpiry.Add(1*time.Minute)) {
		t.Errorf("expiration time is not approximately 10 minutes from now: %v", expiresAt)
	}

	// Save a second authorization for the same user.
	err = SavePasswordResetAuthorization(userID, "second-token-hash")
	if err != nil {
		t.Fatalf("failed to save second authorization: %v", err)
	}

	// Confirm only one authorization remains for this user.
	var count int
	err = DB.QueryRow(`
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

	// Confirm the remaining authorization is the new one.
	var currentHash string
	err = DB.QueryRow(`
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

func createTestAuthorization(t *testing.T, userID int) string {
	t.Helper()

	token, err := utils.GenerateResetAuthorizationToken()
	if err != nil {
		t.Fatalf("generating reset token: %v", err)
	}

	tokenHash := hashResetAuthorizationToken(token)

	err = SavePasswordResetAuthorization(userID, tokenHash)
	if err != nil {
		t.Fatalf("saving reset authorization: %v", err)
	}

	return token
}

func TestValidatePasswordResetAuthorization_Valid(t *testing.T) {
	userID := setupTestUser(t) // Reuse your existing test-user setup.

	token := createTestAuthorization(t, userID)

	err := ValidatePasswordResetAuthorization(token)
	if err != nil {
		t.Fatalf("expected valid authorization, got error: %v", err)
	}
}

func TestValidatePasswordResetAuthorization_Expired(t *testing.T) {
	userID := setupTestUser(t)

	token := createTestAuthorization(t, userID)
	tokenHash := hashResetAuthorizationToken(token)

	_, err := DB.Exec(`
		UPDATE password_reset_authorizations
		SET expires_at = ?
		WHERE token_hash = ?
	`, time.Now().UTC().Add(-1*time.Minute), tokenHash)
	if err != nil {
		t.Fatalf("making authorization expire: %v", err)
	}

	err = ValidatePasswordResetAuthorization(token)
	if err == nil {
		t.Fatal("expected expired authorization to fail")
	}
}

func TestValidatePasswordResetAuthorization_Used(t *testing.T) {
	userID := setupTestUser(t)

	token := createTestAuthorization(t, userID)
	tokenHash := hashResetAuthorizationToken(token)

	_, err := DB.Exec(`
		UPDATE password_reset_authorizations
		SET used_at = ?
		WHERE token_hash = ?
	`, time.Now().UTC(), tokenHash)
	if err != nil {
		t.Fatalf("marking authorization as used: %v", err)
	}

	err = ValidatePasswordResetAuthorization(token)
	if err == nil {
		t.Fatal("expected used authorization to fail")
	}
}
