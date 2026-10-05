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
	"strings"
)

func SharePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(
			w,
			r,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed for this action.",
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
			"You must be logged in to share a post.",
		)
		return
	}

	postID := strings.TrimSpace(r.FormValue("post_id"))

	if postID == "" {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Missing Post",
			"A post ID is required to share a post.",
		)
		return
	}

	id, err := strconv.Atoi(postID)
	if err != nil || id <= 0 {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Invalid Post",
			"The post ID provided is invalid.",
		)
		return
	}

	// Make sure the original post exists.
	originalPost, err := database.GetPostByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RenderError(
				w,
				r,
				http.StatusNotFound,
				"Post Not Found",
				"The post you're trying to share could not be found.",
			)
			return
		}

		log.Printf(
			"SharePostHandler GetPostByID error: %v",
			err,
		)

		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load this post right now.",
		)
		return
	}

	// Record which user shared the original post.
	share := models.PostShare{
		PostID: id,
		UserID: userID,
	}

	// Create the feed post representing the share.
	sharedPost := models.Post{
		UserID:       &userID,
		Content:      "",
		Type:         "normal",
		SharedPostID: &originalPost.ID,
	}

	// Both writes happen inside one database transaction.
	err = database.SharePost(share, sharedPost)
	if err != nil {
		log.Printf(
			"SharePostHandler SharePost error: %v",
			err,
		)

		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't share this post right now.",
		)
		return
	}

	http.Redirect(
		w,
		r,
		"/dashboard",
		http.StatusSeeOther,
	)
}
