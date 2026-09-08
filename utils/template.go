package utils

import (
	"html/template"
	"net/http"
)

func RenderTemplate(w http.ResponseWriter, page string, data any) error {
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/"+page,
	)
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, "base", data)
}
