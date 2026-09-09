package handlers

import (
	"encoding/json"
	"fmt"
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

	postID := strings.TrimSpace(r.FormValue("post_id"))
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

func GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	_, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	postID := r.URL.Query().Get("post_id")

	if postID == "" {
		http.Error(w, "post ID is required", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(postID)
	if err != nil {
		http.Error(w, "invalid post ID", http.StatusBadRequest)
		return
	}

	comments, err := database.GetCommentsByPostID(id)
	if err != nil {
		http.Error(w, "failed to get comments", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(comments); err != nil {
		http.Error(w, fmt.Sprintf("failed to encode comments: %v", err), http.StatusInternalServerError)
		return
	}
}
