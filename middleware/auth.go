package middleware

import (
	"context"
	"jobFlow/constants"
	"jobFlow/database"
	"log"
	"net/http"
	"time"
)

type contextKey string

const UserIDKey contextKey = "userID"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(constants.SessionCookieName)
		if err != nil {
			log.Printf("AuthMiddleware cookie error: %v", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		log.Printf("AuthMiddleware session cookie received")

		session, err := database.GetSessionByToken(cookie.Value)
		if err != nil {
			log.Printf("AuthMiddleware session lookup error: %v", err)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		if time.Now().After(session.ExpiresAt) {
			_ = database.DeleteSession(session.Token)
			http.Error(w, "session expired", http.StatusUnauthorized)
			return
		}

		newContext := context.WithValue(
			r.Context(),
			UserIDKey,
			session.UserID,
		)

		next.ServeHTTP(w, r.WithContext(newContext))
	})
}
