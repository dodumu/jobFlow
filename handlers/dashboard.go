package handlers

import (
	"html/template"
	"log"
	"net/http"

	"jobFlow/database"
	"jobFlow/middleware"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		http.Error(w, "INTERNAL server error", http.StatusInternalServerError)
		return
	}

	posts, err := database.GetFeedPosts(userID)
	if err != nil {
		log.Printf("DASHBOARD ERROR - GetFeedPosts: %v\n", err)
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

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/dashboard.html",
	)
	if err != nil {
		http.Error(w, " server error", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base", data)

	if err != nil {
		log.Printf("TEMPLATE ERROR: %v\n", err)
		http.Error(w, "KINI server error", http.StatusInternalServerError)
		return
	}
}
