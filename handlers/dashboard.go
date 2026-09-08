package handlers

import (
	"html/template"
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
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	posts, err := database.GetFeedPosts(userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	data := struct {
		User  any
		Posts any
	}{
		User:  user,
		Posts: posts,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/dashboard.html",
	)

	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
