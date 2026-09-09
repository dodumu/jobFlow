package handlers

import (
	"database/sql"
	"errors"
	"jobFlow/database"
	"jobFlow/middleware"
	"jobFlow/models"
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
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	content := strings.TrimSpace(r.FormValue("content"))
	postType := strings.TrimSpace(r.FormValue("type"))

	if content == "" {
		http.Error(w, "post content cannot be empty", http.StatusBadRequest)
		return
	}

	if !validPostTypes[postType] {
		http.Error(w, "invalid post type", http.StatusBadRequest)
		return
	}

	post := models.Post{
		UserID:  &userID,
		Content: content,
		Type:    postType,
	}

	_, err := database.CreatePost(post)
	if err != nil {
		http.Error(w, "failed to create post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func LikePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	postID := r.FormValue("post_id")

	if postID == "" {
		http.Error(w, "post ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	like := models.PostLike{
		PostID: id,
		UserID: userID,
	}

	err = database.LikePost(like)
	if err != nil {
		http.Error(w, "failed to like post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
func UnlikePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	postID := r.FormValue("post_id")

	if postID == "" {
		http.Error(w, "post ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	err = database.UnlikePost(id, userID)
	if err != nil {
		http.Error(w, "failed to unlike post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	postID, err := strconv.Atoi(r.FormValue("post_id"))
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	post, err := database.GetPostByID(postID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}

		log.Printf("DELETE POST ERROR - GetPostByID: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	// A post must have a user owner for this version.
	if post.UserID == nil {
		log.Printf("DELETE POST ERROR: post %d has no user owner", postID)
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	// Compare the VALUES, not the pointer addresses.
	log.Printf("Post owner: %d, Current user: %d\n", *post.UserID, userID)

	if *post.UserID != userID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	if err := database.DeletePost(postID); err != nil {
		log.Printf("DELETE POST ERROR - DeletePost: %v\n", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
