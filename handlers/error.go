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
