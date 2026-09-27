package utils

import (
	"html/template"
	"net/http"
	"time"
)

var templateFuncs = template.FuncMap{
	"formatDate": func(t time.Time) string {
		return t.Format("2006-01-02")
	},
}

func RenderTemplate(w http.ResponseWriter, page string, data any) error {
	tmpl, err := template.New("base.html").
		Funcs(templateFuncs).
		ParseFiles(
			"templates/base.html",
			"templates/"+page,
		)
	if err != nil {
		return err
	}

	return tmpl.ExecuteTemplate(w, "base", data)
}
