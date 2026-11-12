package handlers

import (
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
	"log"
	"net/http"
)

func currentUser(w http.ResponseWriter, r *http.Request, loginMessage string) (models.User, bool) {
	userID, ok := utils.GetUserID(r)
	if !ok {
		unauthorized(w, r, loginMessage)
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("currentUser GetUserByID error: %v", err)
		serverError(w, r, "We couldn't load your account information")
		return models.User{}, false
	}
	return user, true
}
