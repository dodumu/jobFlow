package handlers

import (
	"log"
	"net/http"

	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}

	user, ok := currentUser(w, r, "You must be logged in to view your home feed.")
	if !ok {
		return
	}

	posts, err := database.GetFeedPosts(user.ID)
	if err != nil {
		log.Printf("HomeHandler GetFeedPosts error: %v", err)
		serverError(w, r, "We couldn't load your feed right now.")
		return
	}

	data := models.HomePageData{
		User:          user,
		CurrentUserID: user.ID,
		Posts:         posts,
	}

	if err := utils.RenderTemplate(w, r, "home.html", data); err != nil {
		log.Printf("HomeHandler RenderTemplate error: %v", err)
		return
	}
}
