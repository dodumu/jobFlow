package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"jobFlow/constants"
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
	"log"
	"net/http"
)

type ApplicationDetailsData struct {
	Application models.Application
	Job         models.Job
	Applicant   models.User
}

func ApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {

		methodNotAllowed(w, r)
		return
	}

	user, ok := currentUser(w, r, "You must be logged in to view applications.")
	if !ok {
		return
	}

	switch user.Role {

	case constants.RoleIndividual:
		applications, err := database.GetApplicationsByUserID(user.ID)
		if err != nil {
			log.Printf("ApplicationHandler GetApplicationsByUserID error: %v", err)
			serverError(w, r, "We couldn't load your applications right now.")
			return
		}
		pageData := models.UserApplicationsPageData{
			Applications: applications,
			HideNavbar:   false,
		}
		err = utils.RenderTemplate(w, r, "applications.html", pageData)
		if err != nil {
			log.Printf(
				"ApplicationHandler RenderTemplate applications.html error: %v",
				err,
			)
			return
		}

	case constants.RoleCompany:
		company, err := database.GetCompanyByUserID(user.ID)
		if err != nil {
			log.Printf("ApplicationHandler GetCompanyByUserID error: %v", err)
			serverError(w, r, "We couldn't load your company information.")
			return
		}

		applications, err := database.GetApplicationsByCompanyID(company.ID)
		if err != nil {
			log.Printf("ApplicationHandler GetApplicationsByCompanyID error: %v", err)
			serverError(w, r, "We couldn't load your company's applications right now.")
			return
		}
		pageData := models.CompanyApplicationsPageData{
			Applications: applications,
			HideNavbar:   false,
		}
		err = utils.RenderTemplate(w, r, "company_applications.html", pageData)
		if err != nil {
			log.Printf(
				"ApplicationHandler RenderTemplate company_applications.html error: %v",
				err,
			)
			return
		}

	default:
		forbidden(w, r, "You don't have permission to view this page.")
		return
	}
}

func ViewApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}

	user, ok := currentUser(w, r, "You must be logged in to view this application.")
	if !ok {
		return
	}

	if user.Role != constants.RoleCompany {
		forbidden(w, r, "Only company accounts can review applications.")
		return
	}

	company, err := database.GetCompanyByUserID(user.ID)
	if err != nil {
		log.Printf("ViewApplicationHandler GetCompanyByUserID error: %v", err)
		serverError(w, r, "We couldn't load your company information.")
		return
	}

	applicationID, err := utils.GetPathID(r, "id")
	if err != nil || applicationID <= 0 {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Invalid Application",
			"The application ID provided is invalid.",
		)
		return
	}

	application, err := database.GetApplicationByID(applicationID)
	if err != nil {
		notFound(w, r, "Application Not Found", "The application you're looking for could not be found.")
		return
	}

	job, err := database.GetJobByID(application.JobID)
	if err != nil {
		log.Printf("ViewApplicationHandler GetJobByID error: %v", err)
		notFound(w, r, "Job Not Found", "The job associated with this application could not be found.")
		return
	}

	if job.CompanyID != company.ID {
		forbidden(w, r, "You don't have permission to view this application.")
		return
	}

	applicant, err := database.GetUserByID(application.UserID)
	if err != nil {
		log.Printf("ViewApplicationHandler applicant GetUserByID error: %v", err)
		notFound(w, r, "Applicant Not Found", "The applicant associated with this application could not be found.")
		return
	}

	data := ApplicationDetailsData{
		Application: application,
		Job:         job,
		Applicant:   applicant,
	}

	err = utils.RenderTemplate(w, r, "application.html", data)
	if err != nil {
		log.Printf(
			"ApplicationHandler RenderTemplate application.html error: %v",
			err,
		)
		return
	}
}

func updateApplicationStatusHandler(w http.ResponseWriter, r *http.Request, status string) {
	if r.Method != http.MethodPost {
		methodNotAllowed(w, r)
		return
	}

	user, ok := currentUser(w, r, "You must be logged in to review applications.")
	if !ok {
		return
	}

	if user.Role != constants.RoleCompany {
		forbidden(w, r, "Only company accounts can review applications.")
		return
	}

	company, err := database.GetCompanyByUserID(user.ID)
	if err != nil {
		log.Printf("updateApplicationStatusHandler GetCompanyByUserID error: %v", err)
		serverError(w, r, "We couldn't load your company information.")
		return
	}

	applicationID, err := utils.GetPathID(r, "id")
	if err != nil || applicationID <= 0 {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Invalid Application",
			"The application ID provided is invalid.",
		)
		return
	}

	application, err := database.GetApplicationByID(applicationID)
	if err != nil {
		notFound(w, r, "Application Not Found", "The application you're looking for could not be found.")
		return
	}

	job, err := database.GetJobByID(application.JobID)
	if err != nil {
		log.Printf("updateApplicationStatusHandler GetJobByID error: %v", err)
		notFound(w, r, "Job Not Found", "The job associated with this application could not be found.")
		return
	}

	if company.ID != job.CompanyID {
		forbidden(w, r, "You don't have permission to review this application.")
		return
	}

	if application.Status != constants.ApplicationStatusPending {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Application Already Reviewed",
			"This application has already been accepted or rejected.",
		)
		return
	}

	err = database.UpdateApplicationStatus(application.ID, status)
	if err != nil {
		log.Printf("updateApplicationStatusHandler UpdateApplicationStatus error: %v", err)
		serverError(w, r, "We couldn't update the application status.")
		return
	}

	_, err = database.CreateNotification(
		int64(application.UserID),
		sql.NullInt64{
			Int64: int64(user.ID),
			Valid: true,
		},
		"application_status",
		fmt.Sprintf(
			"Your application for %s has been %s",
			job.Title,
			status,
		),
		fmt.Sprintf("/my-applications/%d", application.ID),
	)
	if err != nil {
		log.Printf(
			"updateApplicationStatusHandler CreateNotification error: %v",
			err,
		)
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

func ViewMyApplicationHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		methodNotAllowed(w, r)
		return
	}

	user, ok := currentUser(w, r, "You must be logged in to view you application.")
	if !ok {
		return
	}

	if user.Role != constants.RoleIndividual {
		forbidden(w, r, "Only individual account can view their own applications here.")
		return
	}
	applicationID, err := utils.GetPathID(r, "id")
	if err != nil || applicationID <= 0 {
		RenderError(
			w,
			r,
			http.StatusBadRequest,
			"Invalid Application",
			"The application ID provided is invalid.",
		)
		return
	}
	application, err := database.GetApplicationByID(applicationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			notFound(w, r, "Application Not Found", "The application you're looking for could not be found")
			return
		}
		log.Printf("ViewMyApplicationHandler GetApplicationsByID error: %v", err)
		serverError(w, r, "We couldn't load your application")
		return
	}

	if application.UserID != user.ID {
		forbidden(w, r, "You don't have permission to view the application.")
		return
	}

	job, err := database.GetJobByID(application.JobID)
	if err != nil {
		log.Printf("ViewMyApplicationHandler GetJobByID error : %v", err)
		serverError(w, r, "We couldn't load the job associated with your application.")
		return
	}
	data := ApplicationDetailsData{
		Application: application,
		Job:         job,
		Applicant:   user,
	}
	if err := utils.RenderTemplate(w, r, "application.html", data); err != nil {
		log.Printf("ViewMyApplicationHandler RenderTemplate error: %v ", err)
	}
}
