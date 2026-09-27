package handlers

import (
	"fmt"
	"html/template"
	"jobFlow/constants"
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
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed for this page.",
		)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to view applications.",
		)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("ApplicationHandler GetUserByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your account information.",
		)
		return
	}

	switch user.Role {

	case constants.RoleIndividual:
		applications, err := database.GetApplicationsByUserID(userID)
		if err != nil {
			log.Printf("ApplicationHandler GetApplicationsByUserID error: %v", err)
			RenderError(
				w,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't load your applications right now.",
			)
			return
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/applications.html",
		)
		if err != nil {
			log.Printf("ApplicationHandler template parse error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "base", applications)
		if err != nil {
			log.Printf("ApplicationHandler template execution error: %v", err)
			return
		}

	case constants.RoleCompany:
		company, err := database.GetCompanyByUserID(userID)
		if err != nil {
			log.Printf("ApplicationHandler GetCompanyByUserID error: %v", err)
			RenderError(
				w,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't load your company information.",
			)
			return
		}

		applications, err := database.GetApplicationsByCompanyID(company.ID)
		if err != nil {
			log.Printf("ApplicationHandler GetApplicationsByCompanyID error: %v", err)
			RenderError(
				w,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't load your company's applications right now.",
			)
			return
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/company_applications.html",
		)
		if err != nil {
			log.Printf("ApplicationHandler template parse error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "base", applications)
		if err != nil {
			log.Printf("ApplicationHandler template execution error: %v", err)
			return
		}

	default:
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"You don't have permission to view this page.",
		)
		return
	}
}

func ViewApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed for this page.",
		)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to view this application.",
		)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("ViewApplicationHandler GetUserByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your account information.",
		)
		return
	}

	if user.Role != "company" {
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"Only company accounts can review applications.",
		)
		return
	}

	company, err := database.GetCompanyByUserID(userID)
	if err != nil {
		log.Printf("ViewApplicationHandler GetCompanyByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your company information.",
		)
		return
	}

	applicationID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || applicationID <= 0 {
		RenderError(
			w,
			http.StatusBadRequest,
			"Invalid Application",
			"The application ID provided is invalid.",
		)
		return
	}

	application, err := database.GetApplicationByID(applicationID)
	if err != nil {
		RenderError(
			w,
			http.StatusNotFound,
			"Application Not Found",
			"The application you're looking for could not be found.",
		)
		return
	}

	job, err := database.GetJobByID(application.JobID)
	if err != nil {
		log.Printf("ViewApplicationHandler GetJobByID error: %v", err)
		RenderError(
			w,
			http.StatusNotFound,
			"Job Not Found",
			"The job associated with this application could not be found.",
		)
		return
	}

	if job.CompanyID != company.ID {
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"You don't have permission to view this application.",
		)
		return
	}

	applicant, err := database.GetUserByID(application.UserID)
	if err != nil {
		log.Printf("ViewApplicationHandler applicant GetUserByID error: %v", err)
		RenderError(
			w,
			http.StatusNotFound,
			"Applicant Not Found",
			"The applicant associated with this application could not be found.",
		)
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
		log.Printf("ViewApplicationHandler template parse error: %v", err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		log.Printf("ViewApplicationHandler template execution error: %v", err)
		return
	}
}

func updateApplicationStatusHandler(w http.ResponseWriter, r *http.Request, status string) {
	if r.Method != http.MethodPost {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed for this action.",
		)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to review applications.",
		)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("updateApplicationStatusHandler GetUserByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your account information.",
		)
		return
	}

	if user.Role != constants.RoleCompany {
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"Only company accounts can review applications.",
		)
		return
	}

	company, err := database.GetCompanyByUserID(userID)
	if err != nil {
		log.Printf("updateApplicationStatusHandler GetCompanyByUserID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your company information.",
		)
		return
	}

	applicationID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || applicationID <= 0 {
		RenderError(
			w,
			http.StatusBadRequest,
			"Invalid Application",
			"The application ID provided is invalid.",
		)
		return
	}

	application, err := database.GetApplicationByID(applicationID)
	if err != nil {
		RenderError(
			w,
			http.StatusNotFound,
			"Application Not Found",
			"The application you're looking for could not be found.",
		)
		return
	}

	job, err := database.GetJobByID(application.JobID)
	if err != nil {
		log.Printf("updateApplicationStatusHandler GetJobByID error: %v", err)
		RenderError(
			w,
			http.StatusNotFound,
			"Job Not Found",
			"The job associated with this application could not be found.",
		)
		return
	}

	if company.ID != job.CompanyID {
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"You don't have permission to review this application.",
		)
		return
	}

	if application.Status != constants.JobStatusPending {
		RenderError(
			w,
			http.StatusBadRequest,
			"Application Already Reviewed",
			"This application has already been accepted or rejected.",
		)
		return
	}

	err = database.UpdateApplicationStatus(application.ID, status)
	if err != nil {
		log.Printf("updateApplicationStatusHandler UpdateApplicationStatus error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't update the application status.",
		)
		return
	}

	http.Redirect(
		w,
		r,
		fmt.Sprintf("/applications/%d", application.ID),
		http.StatusSeeOther,
	)
}

func AcceptApplicationHandler(w http.ResponseWriter, r *http.Request) {
	updateApplicationStatusHandler(w, r, "accepted")
}

func RejectApplicationHandler(w http.ResponseWriter, r *http.Request) {
	updateApplicationStatusHandler(w, r, "rejected")
}
