package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
)

func CreateCommentHandler(w http.ResponseWriter, r *http.Request) {
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
			"You must be logged in to create a comment.",
		)
		return
	}

	postID := strings.TrimSpace(r.FormValue("post_id"))
	content := strings.TrimSpace(r.FormValue("content"))

	if postID == "" {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Missing Post",
			"A post ID is required to create a comment.",
		)
		return
	}

	if content == "" {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Empty Comment",
			"Your comment cannot be empty.",
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

	comment := models.Comment{
		PostID:  id,
		UserID:  userID,
		Content: content,
	}

	_, err = database.CreateComment(comment)
	if err != nil {
		log.Printf("CreateCommentHandler CreateComment error: %v", err)
		RenderError(
			w,
			r,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't create your comment right now.",
		)
		return
	}

	http.Redirect(w, r, "/home", http.StatusSeeOther)
}

func GetCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSONError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return
	}

	_, ok := utils.GetUserID(r)
	if !ok {
		writeJSONError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	postID := strings.TrimSpace(r.URL.Query().Get("post_id"))

	if postID == "" {
		writeJSONError(
			w,
			http.StatusBadRequest,
			"post ID is required",
		)
		return
	}

	id, err := strconv.Atoi(postID)
	if err != nil || id <= 0 {
		writeJSONError(
			w,
			http.StatusBadRequest,
			"invalid post ID",
		)
		return
	}

	comments, err := database.GetCommentsByPostID(id)
	if err != nil {
		log.Printf("GetCommentsHandler GetCommentsByPostID error: %v", err)
		writeJSONError(
			w,
			http.StatusInternalServerError,
			"failed to load comments",
		)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(comments); err != nil {
		log.Printf("GetCommentsHandler JSON encoding error: %v", err)
		return
	}
}

func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSONError(
			w,
			http.StatusMethodNotAllowed,
			"method not allowed",
		)
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		writeJSONError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	commentID, err := strconv.Atoi(
		strings.TrimSpace(r.FormValue("comment_id")),
	)
	if err != nil || commentID <= 0 {
		writeJSONError(
			w,
			http.StatusBadRequest,
			"invalid comment ID",
		)
		return
	}

	comment, err := database.GetCommentByID(commentID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(
				w,
				http.StatusNotFound,
				"comment not found",
			)
			return
		}

		log.Printf("DeleteCommentHandler GetCommentByID error: %v", err)
		writeJSONError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	post, err := database.GetPostByID(comment.PostID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSONError(
				w,
				http.StatusNotFound,
				"post not found",
			)
			return
		}

		log.Printf("DeleteCommentHandler GetPostByID error: %v", err)
		writeJSONError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	// The comment owner can delete their own comment.
	// The post owner can delete any comment on their post.
	if comment.UserID != userID {
		if post.UserID == nil || *post.UserID != userID {
			writeJSONError(
				w,
				http.StatusForbidden,
				"you don't have permission to delete this comment",
			)
			return
		}
	}

	if err := database.DeleteComment(commentID); err != nil {
		log.Printf("DeleteCommentHandler DeleteComment error: %v", err)
		writeJSONError(
			w,
			http.StatusInternalServerError,
			"internal server error",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
