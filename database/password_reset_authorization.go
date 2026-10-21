package database

import (
	"fmt"
	"time"
)

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
