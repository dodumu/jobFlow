package database

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type PasswordResetAuthorization struct {
	ID        int
	UserID    int
	TokenHash string
	ExpiresAt time.Time
	UsedAt    sql.NullTime
	CreatedAt time.Time
}

func hashResetAuthorizationToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
func GetPasswordResetAuthorization(
	userID int,
) (PasswordResetAuthorization, error) {
	var authorization PasswordResetAuthorization

	err := DB.QueryRow(`
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_authorizations
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, userID).Scan(
		&authorization.ID,
		&authorization.UserID,
		&authorization.TokenHash,
		&authorization.ExpiresAt,
		&authorization.UsedAt,
		&authorization.CreatedAt,
	)

	if err != nil {
		return PasswordResetAuthorization{},
			fmt.Errorf("getting reset authorization: %w", err)
	}

	return authorization, nil
}

func SavePasswordResetAuthorization(
	userID int,
	tokenHash string,
) error {
	expiresAt := time.Now().UTC().Add(10 * time.Minute)

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		DELETE FROM password_reset_authorizations
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return fmt.Errorf("deleting previous authorizations: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO password_reset_authorizations
			(user_id, token_hash, expires_at)
		VALUES (?, ?, ?)
	`, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("saving reset authorization: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func ValidatePasswordResetAuthorization(token string) error {
	if token == "" {
		return errors.New("invalid reset authorization")
	}

	tokenHash := hashResetAuthorizationToken(token)

	var authorization PasswordResetAuthorization

	err := DB.QueryRow(`
		SELECT id, user_id, token_hash, expires_at, used_at, created_at
		FROM password_reset_authorizations
		WHERE token_hash = ?
		LIMIT 1
	`, tokenHash).Scan(
		&authorization.ID,
		&authorization.UserID,
		&authorization.TokenHash,
		&authorization.ExpiresAt,
		&authorization.UsedAt,
		&authorization.CreatedAt,
	)

	if err != nil {
		return errors.New("invalid reset authorization")
	}

	if authorization.UsedAt.Valid {
		return errors.New("reset authorization has already been used")
	}

	if !time.Now().UTC().Before(authorization.ExpiresAt) {
		return errors.New("reset authorization has expired")
	}

	return nil
}

func ConsumePasswordResetAuthorization(token string) error {
	if token == "" {
		return errors.New("invalid reset authorization")
	}

	tokenHash := hashResetAuthorizationToken(token)

	result, err := DB.Exec(`
		UPDATE password_reset_authorizations
		SET used_at = ?
		WHERE token_hash = ?
		  AND used_at IS NULL
		  AND expires_at > ?
	`,
		time.Now().UTC(),
		tokenHash,
		time.Now().UTC(),
	)
	if err != nil {
		return fmt.Errorf("consuming reset authorization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking consumed authorization: %w", err)
	}

	if rowsAffected != 1 {
		return errors.New("invalid, expired, or already-used reset authorization")
	}

	return nil
}

func VerifyPasswordResetCode(userID int, submittedCode string) error {
	if userID <= 0 {
		return errors.New("invalid user")
	}

	if len(submittedCode) != 8 {
		return errors.New("invalid verification code")
	}

	for _, char := range submittedCode {
		if char < '0' || char > '9' {
			return errors.New("invalid verification code")
		}
	}

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("starting verification transaction: %w", err)
	}
	defer tx.Rollback()

	var reset PasswordResetCode

	err = tx.QueryRow(`
		SELECT id, user_id, code_hash, expires_at,
		       attempts, verified_at, used_at
		FROM password_reset_codes
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, userID).Scan(
		&reset.ID,
		&reset.UserID,
		&reset.CodeHash,
		&reset.ExpiresAt,
		&reset.Attempts,
		&reset.VerifiedAt,
		&reset.UsedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("invalid or expired verification code")
		}
		return fmt.Errorf("retrieving verification code: %w", err)
	}

	now := time.Now().UTC()

	if reset.UsedAt.Valid {
		return errors.New("verification code has already been used")
	}

	if reset.VerifiedAt.Valid {
		return errors.New("verification code has already been verified")
	}

	if !now.Before(reset.ExpiresAt) {
		return errors.New("verification code has expired")
	}

	if reset.Attempts >= 5 {
		return errors.New("too many verification attempts")
	}

	secret := os.Getenv("RESET_CODE_SECRET")
	if secret == "" {
		return errors.New("RESET_CODE_SECRET is not configured")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	message := fmt.Sprintf("%d:%s", userID, submittedCode)
	_, _ = mac.Write([]byte(message))
	submittedHash := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal(
		[]byte(reset.CodeHash),
		[]byte(submittedHash),
	) {
		_, err = tx.Exec(`
			UPDATE password_reset_codes
			SET attempts = attempts + 1
			WHERE id = ?
		`, reset.ID)
		if err != nil {
			return fmt.Errorf("recording failed verification attempt: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("saving verification attempt: %w", err)
		}

		return errors.New("incorrect verification code")
	}

	result, err := tx.Exec(`
		UPDATE password_reset_codes
		SET verified_at = ?
		WHERE id = ?
		  AND verified_at IS NULL
		  AND used_at IS NULL
		  AND expires_at > ?
		  AND attempts < 5
	`, now, reset.ID, now)
	if err != nil {
		return fmt.Errorf("marking verification code as verified: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking verification result: %w", err)
	}
	if rowsAffected != 1 {
		return errors.New("verification code is no longer valid")
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing code verification: %w", err)
	}

	return nil
}
