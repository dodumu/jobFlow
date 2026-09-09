package handlers

import (
	"jobFlow/database"
	"jobFlow/middleware"
	"jobFlow/models"
	"net/http"
	"strconv"
	"strings"
)

func SharePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	postID := strings.TrimSpace(r.FormValue("post_id"))

	if postID == "" {
		http.Error(w, "post ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	// Make sure the original post exists.
	originalPost, err := database.GetPostByID(id)
	if err != nil {
		http.Error(w, "post not found", http.StatusNotFound)
		return
	}

	// Record the share.
	share := models.PostShare{
		PostID: id,
		UserID: userID,
	}

	_, err = database.CreatePostShare(share)
	if err != nil {
		http.Error(w, "failed to record share", http.StatusInternalServerError)
		return
	}

	// Create a new post representing the share.
	sharedPost := models.Post{
		UserID:       &userID,
		Content:      originalPost.Content,
		Type:         originalPost.Type,
		SharedPostID: &originalPost.ID,
	}

	_, err = database.CreatePost(sharedPost)
	if err != nil {
		http.Error(w, "failed to create shared post", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
