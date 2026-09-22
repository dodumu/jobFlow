package handlers

import (
	"fmt"
	"html/template"
	"jobFlow/database"
	"jobFlow/middleware"
	"jobFlow/models"
	"log"
	"net/http"
	"strconv"
)

type ApplicationDetailsData struct {
	Application models.Application
	Job         models.Job
	Applicant   models.User
}

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

func ViewApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	user, err := database.GetUserByID(userID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	if user.Role != "company" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	company, err := database.GetCompanyByUserID(userID)
	if err != nil {
		http.Error(w, "company not found", http.StatusNotFound)
		return
	}
	applicationID := r.PathValue("id")
	applicationIDInt, err := strconv.Atoi(applicationID)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	application, err := database.GetApplicationByID(applicationIDInt)
	if err != nil {
		http.Error(w, "application not found", http.StatusNotFound)
		return
	}
	job, err := database.GetJobByID(application.JobID)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	if job.CompanyID != company.ID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	applicant, err := database.GetUserByID(application.UserID)
	if err != nil {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	data := ApplicationDetailsData{
		Application: application,
		Job:         job,
		Applicant:   applicant,
	}
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/application.html",
	)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("ViewApplicationHandler template error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func updateApplicationStatusHandler(w http.ResponseWriter, r *http.Request, status string) {
	if r.Method != http.MethodPost {
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
	if user.Role != "company" {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	company, err := database.GetCompanyByUserID(userID)
	if err != nil {
		http.Error(w, "company not found", http.StatusNotFound)
		return
	}
	applicationID := r.PathValue("id")
	applicationIDInt, err := strconv.Atoi(applicationID)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	application, err := database.GetApplicationByID(applicationIDInt)
	if err != nil {
		http.Error(w, "application not found", http.StatusNotFound)
		return
	}
	job, err := database.GetJobByID(application.JobID)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	if company.ID != job.CompanyID {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if application.Status != "pending" {
		http.Error(w, "application not pending", http.StatusBadRequest)
		return
	}
	err = database.UpdateApplicationStatus(application.ID, status)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/applications/%d", application.ID), http.StatusSeeOther)
}

func AcceptApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	updateApplicationStatusHandler(w, r, "accepted")
}

func RejectApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	updateApplicationStatusHandler(w, r, "rejected")
}
