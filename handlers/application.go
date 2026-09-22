package handlers

import (
	"html/template"
	"jobFlow/database"
	"jobFlow/middleware"
	"log"
	"net/http"
)

func ApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}

	switch user.Role {

	case "individual":
		applications, err := database.GetApplicationsByUserID(userID)
		if err != nil {
			http.Error(w, "failed to load applications", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/applications.html",
		)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "base", applications)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

	case "company":
		company, err := database.GetCompanyByUserID(userID)
		if err != nil {
			http.Error(w, "company not found", http.StatusNotFound)
			return
		}

		applications, err := database.GetApplicationsByCompanyID(company.ID)
		if err != nil {
			log.Printf("GetApplicationsByCompanyID error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/company_applications.html",
		)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "base", applications)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
}
