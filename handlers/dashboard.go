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
		methodNotAllowed(w, r)
		return
	}

	user, ok := currentUser(w, r, "You must be logged in to view your dashboard.")
	if !ok {
		return
	}

	posts, err := database.GetPostsByUserID(user.ID)
	if err != nil {
		log.Printf("DashboardHandler GetPostsByUserID error: %v", err)
		serverError(w, r, "We couldn't load your posts right now.")
		return
	}

	data := models.DashboardPageData{
		User:          user,
		CurrentUserID: user.ID,
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
