package handlers

import (
	"jobFlow/models"
	"jobFlow/utils"
	"log"
	"net/http"
)

func RenderError(w http.ResponseWriter, r *http.Request, statusCode int, title string, message string) {
	pageError := models.ErrorPageData{
		StatusCode: statusCode,
		Title:      title,
		Message:    message,
	}

	w.WriteHeader(statusCode)
	err := utils.RenderTemplate(w, r, "error.html", pageError)
	if err != nil {
		log.Printf(
			"ApplicationHandler RenderTemplate error.html error: %v",
			err,
		)
	}
}

func serverError(w http.ResponseWriter, r *http.Request, message string) {
	RenderError(w, r, http.StatusInternalServerError, "Something Went Wrong", message)
}

func unauthorized(w http.ResponseWriter, r *http.Request, message string) {
	RenderError(w, r, http.StatusUnauthorized, "Unauthorized", message)
}

func forbidden(w http.ResponseWriter, r *http.Request, message string) {
	RenderError(w, r, http.StatusForbidden, "Access Denied", message)
}

func notFound(w http.ResponseWriter, r *http.Request, title, message string) {
	RenderError(w, r, http.StatusNotFound, title, message)
}

func methodNotAllowed(w http.ResponseWriter, r *http.Request) {
	RenderError(w, r, http.StatusMethodNotAllowed, "Method Not Allowed", "The requested method is not allowed on the page.")
}
