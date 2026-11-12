package handlers

import (
	"database/sql"
	"errors"
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
	"log"
	"net/http"
	"strconv"
)

func PostDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}

	user, ok := currentUser(w, r, "You must be logged in to view this post")
	if !ok {
		return
	}

	PostID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || PostID <= 0 {
		RenderError(w, r, http.StatusBadRequest, "Invalid Post", "THe requested post ID is inivalid.")
		return
	}

	post, err := database.GetFeedPostByID(PostID, user.ID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, r, "Post Not Found", "The post you're looking for doesn't exist.")
			return
		}

		log.Printf("PostDetailsHandler GetFeedPostByID error: %v", err)
		serverError(w, r, "We couldn't load this post right now.")
		return
	}

	comments, err := database.GetCommentsByPostID(PostID)
	if err != nil {
		log.Printf("PostDetailsHandler GetCOmmentsByPostID error: %v", err)
		serverError(w, r, "We couldn't load the comments.")
		return
	}

	likes, err := database.GetPostLikes(PostID)
	if err != nil {
		log.Printf("PostDetailsHandler GetPostLikes eerror: %v", err)
		serverError(w, r, "We couldn't load the likes")
		return
	}

	data := models.PostDetailsPageData{
		User:          user,
		CurrentUserID: user.ID,
		Post:          post,
		Comments:      comments,
		Likes:         likes,
	}

	if err := utils.RenderTemplate(w, r, "post_details.html", data); err != nil {
		log.Printf("PostDetailsHandler RenderTemplate error: %v", err)
		return
	}
}
