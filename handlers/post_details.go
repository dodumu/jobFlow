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
		RenderError(w, r, http.StatusMethodNotAllowed, "Method Not Allowed", "The requested method is not allowed on this page.")
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(w, r, http.StatusUnauthorized, "Unauthorized", "You must be logged in to view this post")
		return
	}

	PostID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || PostID <= 0 {
		RenderError(w, r, http.StatusBadRequest, "Invalid Post", "THe requested post ID is inivalid.")
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("PostDetailsHandler GetUserByID error: %v", err)
		RenderError(w, r, http.StatusInternalServerError, "Something Went Wrong", "We couldn't load your account information.")
		return
	}

	post, err := database.GetFeedPostByID(PostID, userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RenderError(w, r, http.StatusNotFound, "Post Not Found", "The post you're looking for doesn't exist.")
			return
		}

		log.Printf("PostDetailsHandler GetFeedPostByID error: %v", err)
		RenderError(w, r, http.StatusInternalServerError, "Something Went Wrong", "We couldn't load this post right now.")
		return
	}

	comments, err := database.GetCommentsByPostID(PostID)
	if err != nil {
		log.Printf("PostDetailsHandler GetCOmmentsByPostID error: %v", err)
		RenderError(w, r, http.StatusInternalServerError, "Something Went Wrong", "We couldn't load the comments.")
		return
	}

	likes, err := database.GetPostLikes(PostID)
	if err != nil {
		log.Printf("PostDetailsHandler GetPostLikes eerror: %v", err)
		RenderError(w, r, http.StatusInternalServerError, "Something Went Wrong", "We couldn't load the likes")
		return
	}

	data := models.PostDetailsPageData{
		User:          user,
		CurrentUserID: userID,
		Post:          post,
		Comments:      comments,
		Likes:         likes,
	}

	if err := utils.RenderTemplate(w, r, "post_details.html", data); err != nil {
		log.Printf("PostDetailsHandler RenderTemplate error: %v", err)
		return
	}
}
