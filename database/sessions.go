package database

import (
	"fmt"
	"jobFlow/models"
)

func CreateSession(session models.Session) error {
	query := `INSERT INTO sessions (user_id, token, expires_at) VALUES (?, ?, ?)`

	_, err := DB.Exec(query, session.UserID, session.Token, session.ExpiresAt)
	if err != nil {
		return err
	}
	return nil
}

func GetSessionByToken(token string) (models.Session, error) {
	query := `SELECT id, user_id, token, expires_at, created_at FROM sessions WHERE token = ?`
	var session models.Session
	err := DB.QueryRow(query, token).Scan(&session.ID,
		&session.UserID,
		&session.Token,
		&session.ExpiresAt,
		&session.CreatedAt)
	if err != nil {
		return models.Session{}, err
	}
	return session, nil
}

func DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = ?`
	res, err := DB.Exec(query, token)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("session with token %v not found", token)
	}
	return nil
}
