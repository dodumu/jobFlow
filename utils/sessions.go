package utils

import (
	"errors"
	"net/http"
	"time"

	"jobFlow/database"
)

func GetUserIDFromSession(r *http.Request) (int, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0, err
	}

	session, err := database.GetSessionByToken(cookie.Value)
	if err != nil {
		return 0, err
	}

	if session.ExpiresAt.Before(time.Now()) {
		return 0, errors.New("session expired")
	}

	return session.UserID, nil
}
