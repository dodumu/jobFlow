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

var validPostTypes = map[string]bool{
	"normal":       true,
	"project":      true,
	"hiring":       true,
	"announcement": true,
	"job":          true,
}

func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
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
			"You must be logged in to create a post.",
		)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	postType := strings.TrimSpace(r.FormValue("type"))

	if content == "" {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Empty Post",
			"Post content cannot be empty.",
		)
		return
	}

	if !validPostTypes[postType] {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Invalid Post Type",
			"Please select a valid post type.",
		)
		return
	}

	post := models.Post{
		UserID:  &userID,
		Content: content,
		Type:    postType,
	}

	_, err := database.CreatePost(post)
	if err != nil {
		log.Printf("CreatePostHandler CreatePost error: %v", err)

		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't create your post right now.",
		)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func LikePostHandler(w http.ResponseWriter, r *http.Request) {
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
			"You must be logged in to like a post.",
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
			"A post ID is required.",
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

	like := models.PostLike{
		PostID: id,
		UserID: userID,
	}

	if err := database.LikePost(like); err != nil {
		log.Printf(
			"LikePostHandler LikePost error - postID=%d userID=%d: %v",
			id,
			r,
			userID,
			err,
		)

		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't like this post right now.",
		)
		return
	}

	utils.RedirectBack(w, r, "/home")
}

func UnlikePostHandler(w http.ResponseWriter, r *http.Request) {
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
			"You must be logged in to unlike a post.",
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
			"A post ID is required.",
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

	if err := database.UnlikePost(id, userID); err != nil {
		log.Printf(
			"UnlikePostHandler UnlikePost error - postID=%d userID=%d: %v",
			id,
			userID,
			err,
		)

		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't unlike this post right now.",
		)
		return
	}

	utils.RedirectBack(w, r, "/home")
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
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
			"You must be logged in to delete a post.",
		)
		return
	}

	postID, err := strconv.Atoi(
		strings.TrimSpace(r.FormValue("post_id")),
	)
	if err != nil || postID <= 0 {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Invalid Post",
			"The post ID provided is invalid.",
		)
		return
	}

	post, err := database.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RenderError(
				w,
				r,
				http.StatusNotFound,
				"Post Not Found",
				"The post you're trying to delete could not be found.",
			)
			return
		}

		log.Printf("DeletePostHandler GetPostByID error: %v", err)

		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load this post right now.",
		)
		return
	}

	if post.UserID == nil {
		RenderError(
			w,
			r,
			http.StatusForbidden,
			"Access Denied",
			"You don't have permission to delete this post.",
		)
		return
	}

	if *post.UserID != userID {
		RenderError(
			w,
			r,
			http.StatusForbidden,
			"Access Denied",
			"You don't have permission to delete this post.",
		)
		return
	}

	if err := database.DeletePost(postID); err != nil {
		log.Printf("DeletePostHandler DeletePost error: %v", err)

		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't delete this post right now.",
		)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
