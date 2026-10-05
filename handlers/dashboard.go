package handlers

import (
	"log"
	"net/http"

	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
)

func DashboardHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(
			w,
			r,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(
			w,
			r,
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
			r,
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
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your posts right now.",
		)
		return
	}

	data := models.DashboardPageData{
		User:          user,
		CurrentUserID: userID,
		Posts:         posts,
	}

	err = utils.RenderTemplate(w, r, "dashboard.html", data)
	if err != nil {
		log.Printf(
			"DashboardHandler RenderTemplate dashboard.html error: %v",
			err,
		)
	}
}
