package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"jobFlow/models"
)

func GetLoginAttempt(username, ipAddress string) (*models.LoginAttempt, error) {
	query := `
		SELECT
			id,
			username,
			ip_address,
			failed_attempts,
			last_attempt_at,
			blocked_until
		FROM login_attempts
		WHERE username = ? AND ip_address = ?
	`

	attempt := &models.LoginAttempt{}

	err := DB.QueryRow(query, username, ipAddress).Scan(
		&attempt.ID,
		&attempt.Username,
		&attempt.IPAddress,
		&attempt.FailedAttempts,
		&attempt.LastAttemptAt,
		&attempt.BlockedUntil,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("getting login attempt: %w", err)
	}

	return attempt, nil
}

func RecordFailedLoginAttempt(username, ipAddress string) error {
	query := `
		INSERT INTO login_attempts (
			username,
			ip_address,
			failed_attempts,
			last_attempt_at
		)
		VALUES (?, ?, 1, CURRENT_TIMESTAMP)

		ON CONFLICT(username, ip_address)
		DO UPDATE SET
			failed_attempts = failed_attempts + 1,
			last_attempt_at = CURRENT_TIMESTAMP
	`

	_, err := DB.Exec(query, username, ipAddress)
	if err != nil {
		return fmt.Errorf("recording failed login attempt: %w", err)
	}

	return nil
}

func BlockLoginAttempts(username, ipAddress string, blockedUntil time.Time) error {
	query := `
		UPDATE login_attempts
		SET blocked_until = ?
		WHERE username = ? AND ip_address = ?
	`

	_, err := DB.Exec(
		query,
		blockedUntil,
		username,
		ipAddress,
	)
	if err != nil {
		return fmt.Errorf("blocking login attempts: %w", err)
	}

	return nil
}

func ResetLoginAttempts(username, ipAddress string) error {
	query := `
		DELETE FROM login_attempts
		WHERE username = ? AND ip_address = ?
	`

	_, err := DB.Exec(query, username, ipAddress)
	if err != nil {
		return fmt.Errorf("resetting login attempts: %w", err)
	}

	return nil
}
