package database

import (
	"database/sql"
	"fmt"
	"time"
)

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
