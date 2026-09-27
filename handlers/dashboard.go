package handlers

import (
	"html/template"
	"log"
	"net/http"

	"jobFlow/database"
	"jobFlow/middleware"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
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
			"You must be logged in to view your dashboard.",
		)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("DashboardHandler GetUserByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your account information.",
		)
		return
	}

	posts, err := database.GetPostsByUserID(userID)
	if err != nil {
		log.Printf("DashboardHandler GetPostsByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your posts right now.",
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

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/dashboard.html",
	)
	if err != nil {
		log.Printf("DashboardHandler template parse error: %v", err)
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("DashboardHandler template execution error: %v", err)
		return
	}
}
