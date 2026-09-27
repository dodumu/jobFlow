package handlers

import (
	"log"
	"net/http"

	"jobFlow/database"
	"jobFlow/middleware"
	"jobFlow/utils"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to view your home feed.",
		)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("HomeHandler GetUserByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your account information.",
		)
		return
	}

	posts, err := database.GetFeedPosts(userID)
	if err != nil {
		log.Printf("HomeHandler GetFeedPosts error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your feed right now.",
		)
		return
	}

	data := struct {
		User          any
		CurrentUserID int
		Posts         any
	}{
		User:          user,
		CurrentUserID: userID,
		Posts:         posts,
	}

	if err := utils.RenderTemplate(w, "home.html", data); err != nil {
		log.Printf("HomeHandler RenderTemplate error: %v", err)
		return
	}
}
