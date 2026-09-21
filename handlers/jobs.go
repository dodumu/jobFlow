package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"jobFlow/database"
	"jobFlow/middleware"
	"jobFlow/models"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func JobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	jobs, err := database.GetJobs()
	if err != nil {
		http.Error(w, "failed to load jobs", http.StatusInternalServerError)
		return
	}

	_, err = database.GetCompanyByUserID(userID)

	isCompany := err == nil

	data := struct {
		Jobs      []models.Job
		IsCompany bool
	}{
		Jobs:      jobs,
		IsCompany: isCompany,
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/jobs.html",
	)
	if err != nil {
		http.Error(w, "failed to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "failed to render jobs", http.StatusInternalServerError)
		return
	}
}

func JobDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/jobs/")

	if idStr == "" {
		http.NotFound(w, r)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	job, err := database.GetJobByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "failed to load job", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/job.html",
	)
	if err != nil {
		http.Error(w, "failed to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", job)
	if err != nil {
		http.Error(w, "failed to render job", http.StatusInternalServerError)
		return
	}
}

func CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		// Show the create-job form.
		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/create-job.html",
		)
		if err != nil {
			http.Error(w, "failed to load template", http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "base", nil)
		if err != nil {
			http.Error(w, "failed to render template", http.StatusInternalServerError)
			return
		}

	case http.MethodPost:
		// Read form values.
		companyID := strings.TrimSpace(r.FormValue("company_id"))
		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))
		location := strings.TrimSpace(r.FormValue("location"))
		employmentType := strings.TrimSpace(r.FormValue("employment_type"))
		salaryMin := strings.TrimSpace(r.FormValue("salary_min"))
		salaryMax := strings.TrimSpace(r.FormValue("salary_max"))
		deadlineStr := strings.TrimSpace(r.FormValue("deadline"))

		deadline, err := time.Parse("2006-01-02T15:04", deadlineStr)
		if err != nil {
			http.Error(w, "invalid deadline", http.StatusBadRequest)
			return
		}
		// Validate required fields.
		if companyID == "" ||
			title == "" ||
			description == "" ||
			location == "" ||
			employmentType == "" ||
			salaryMin == "" ||
			salaryMax == "" {

			http.Error(
				w,
				"all required fields must be filled",
				http.StatusBadRequest,
			)
			return
		}

		// Convert company ID.
		compID, err := strconv.Atoi(companyID)
		if err != nil {
			http.Error(w, "invalid company ID", http.StatusBadRequest)
			return
		}

		// Convert minimum salary.
		minSalary, err := strconv.Atoi(salaryMin)
		if err != nil {
			http.Error(w, "invalid minimum salary", http.StatusBadRequest)
			return
		}

		// Convert maximum salary.
		maxSalary, err := strconv.Atoi(salaryMax)
		if err != nil {
			http.Error(w, "invalid maximum salary", http.StatusBadRequest)
			return
		}

		// Make sure minimum salary isn't greater than maximum salary.
		if minSalary > maxSalary {
			http.Error(
				w,
				"minimum salary cannot be greater than maximum salary",
				http.StatusBadRequest,
			)
			return
		}

		job := models.Job{
			CompanyID:      compID,
			Title:          title,
			Description:    description,
			Location:       location,
			EmploymentType: employmentType,
			SalaryMin:      minSalary,
			SalaryMax:      maxSalary,
			Deadline:       deadline,
			Status:         "open",
		}

		newJob, err := database.CreateJob(job)
		if err != nil {
			http.Error(
				w,
				"failed to create job",
				http.StatusInternalServerError,
			)
			return
		}

		// Redirect to the newly created job.
		http.Redirect(
			w,
			r,
			fmt.Sprintf("/jobs/%d", newJob),
			http.StatusSeeOther,
		)

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
