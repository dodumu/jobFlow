package database

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
	
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

	tokenHash := HashResetAuthorizationToken(token)

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

	tokenHash := HashResetAuthorizationToken(token)

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

func hashResetAuthorizationToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}