package utils

import (
	"jobFlow/middleware"
	"net/http"
)

func GetUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok || userID <= 0 {
		return 0, false
	}

	return userID, true
}
