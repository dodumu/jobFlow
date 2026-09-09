package handlers

import (
	"jobFlow/database"
	"jobFlow/middleware"
	"jobFlow/models"
	"net/http"
	"strconv"
	"strings"
)

func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
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
	content := strings.TrimSpace(r.FormValue("content"))

	if postID == "" {
		http.Error(w, "post ID is required", http.StatusBadRequest)
		return
	}

	if content == "" {
		http.Error(w, "comment cannot be empty", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	comment := models.Comment{
		PostID:  id,
		UserID:  userID,
		Content: content,
	}

	_, err = database.CreateComment(comment)
	if err != nil {
		http.Error(w, "failed to create comment", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}
