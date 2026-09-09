package handlers

import (
	"log"
	"net/http"

	"jobFlow/database"
	"jobFlow/middleware"
	"jobFlow/utils"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("HOME ERROR - GetUserByID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	posts, err := database.GetFeedPosts(userID)
	if err != nil {
		log.Printf("HOME ERROR - GetFeedPosts: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
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
		log.Printf("HOME ERROR - RenderTemplate: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}
}
