package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func ResetPassword(token, newPasswordHash string) error {
	if token == "" {
		return errors.New("invalid reset authorization")
	}

	if newPasswordHash == "" {
		return errors.New("password hash cannot be empty")
	}

	tokenHash := hashResetAuthorizationToken(token)

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("starting password reset transaction: %w", err)
	}
	defer tx.Rollback()

	var authorizationID int
	var userID int
	var expiresAt time.Time
	var usedAt sql.NullTime

	err = tx.QueryRow(`
		SELECT id, user_id, expires_at, used_at
		FROM password_reset_authorizations
		WHERE token_hash = ?
		LIMIT 1
	`, tokenHash).Scan(
		&authorizationID,
		&userID,
		&expiresAt,
		&usedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errors.New("invalid reset authorization")
		}
		return fmt.Errorf("looking up reset authorization: %w", err)
	}

	if usedAt.Valid {
		return errors.New("reset authorization has already been used")
	}

	if !time.Now().UTC().Before(expiresAt) {
		return errors.New("reset authorization has expired")
	}

	// We'll add the password update and token consumption here next.

	now := time.Now().UTC()

	// 1. Consume the authorization, but only if it is still valid.
	result, err := tx.Exec(`
		UPDATE password_reset_authorizations
		SET used_at = ?
		WHERE id = ?
		  AND used_at IS NULL
		  AND expires_at > ?
	`, now, authorizationID, now)
	if err != nil {
		return fmt.Errorf("consuming reset authorization: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking reset authorization: %w", err)
	}
	if rowsAffected != 1 {
		return errors.New("invalid, expired, or already-used reset authorization")
	}

	// 2. Update the user's password hash.
	result, err = tx.Exec(`
		UPDATE users
		SET password_hash = ?
		WHERE id = ?
	`, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("updating password hash: %w", err)
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking password update: %w", err)
	}
	if rowsAffected != 1 {
		return errors.New("user account not found")
	}

	// 3. Consume the verified password-reset code.
	result, err = tx.Exec(`
		UPDATE password_reset_codes
		SET used_at = ?
		WHERE user_id = ?
		  AND verified_at IS NOT NULL
		  AND used_at IS NULL
	`, now, userID)
	if err != nil {
		return fmt.Errorf("consuming reset code: %w", err)
	}

	rowsAffected, err = result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking reset code consumption: %w", err)
	}
	if rowsAffected < 1 {
		return errors.New("verified reset code not found")
	}

	// 4. Revoke all existing sessions for this user.
	_, err = tx.Exec(`
		DELETE FROM sessions
		WHERE user_id = ?
	`, userID)
	if err != nil {
		return fmt.Errorf("revoking user sessions: %w", err)
	}

	// 5. Commit every change together.
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing password reset: %w", err)
	}
	return nil
}

type PasswordResetCode struct {
	ID         int
	UserID     int
	CodeHash   string
	ExpiresAt  time.Time
	Attempts   int
	VerifiedAt sql.NullTime
	UsedAt     sql.NullTime
}

func SavePasswordResetCode(userID int, codeHash string) error {

	expiresAt := time.Now().UTC().Add(15 * time.Minute)

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.Exec(
		`DELETE FROM password_reset_codes WHERE user_id = ?`,
		userID,
	)
	if err != nil {
		return fmt.Errorf("deleting previous reset codes: %w", err)
	}

	_, err = tx.Exec(
		`INSERT INTO password_reset_codes
		(user_id, code_hash, expires_at)
		VALUES (?, ?, ?)`,
		userID,
		codeHash,
		expiresAt,
	)
	if err != nil {
		return fmt.Errorf("saving reset code: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing transaction: %w", err)
	}

	return nil
}

func GetPasswordResetCode(userID int) (PasswordResetCode, error) {
	var reset PasswordResetCode

	query := `
		SELECT
			id,
			user_id,
			code_hash,
			expires_at,
			attempts,
			verified_at,
			used_at
		FROM password_reset_codes
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`

	err := DB.QueryRow(query, userID).Scan(
		&reset.ID,
		&reset.UserID,
		&reset.CodeHash,
		&reset.ExpiresAt,
		&reset.Attempts,
		&reset.VerifiedAt,
		&reset.UsedAt,
	)

	if err != nil {
		return PasswordResetCode{}, err
	}

	return reset, nil
}
