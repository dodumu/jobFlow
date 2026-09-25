package handlers

import (
	"html/template"
	"jobFlow/models"
	"log"
	"net/http"
)

func RenderError(w http.ResponseWriter, statusCode int, title string, message string) {
	pageError := models.ErrorPageData{
		StatusCode: statusCode,
		Title:      title,
		Message:    message,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/error.html",
	)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	err = tmpl.ExecuteTemplate(w, "base", pageError)
	if err != nil {
		log.Print("rendering error template: %v", err)
		return
	}
}
