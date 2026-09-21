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

	// Get job ID from URL: /jobs/{id}
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

	// Get the job.
	job, err := database.GetJobByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.NotFound(w, r)
			return
		}

		http.Error(w, "failed to load job", http.StatusInternalServerError)
		return
	}

	// By default, the current user cannot edit this job.
	canEdit := false

	// Get authenticated user's ID.
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)

	if ok {
		// Check whether this user owns a company.
		company, err := database.GetCompanyByUserID(userID)

		// If they own a company, check whether that company owns this job.
		if err == nil && company.ID == job.CompanyID {
			canEdit = true
		}
	}

	// Combine the Job model with page-specific information.
	data := models.JobDetailsData{
		Job:     job,
		CanEdit: canEdit,
	}

	// Parse template.
	tmpl, err := template.ParseFiles(
		"templates/base.html",
		"templates/job.html",
	)
	if err != nil {
		http.Error(w, "failed to load template", http.StatusInternalServerError)
		return
	}

	// Render template using JobDetailsData instead of just Job.
	err = tmpl.ExecuteTemplate(w, "base", data)
	if err != nil {
		http.Error(w, "failed to render job", http.StatusInternalServerError)
		return
	}
}

func CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {

	case http.MethodGet:
		// Show the create-job form.
		userID, ok := r.Context().Value(middleware.UserIDKey).(int)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		company, err := database.GetCompanyByUserID(userID)
		if err != nil {
			http.Error(w, "company not found", http.StatusForbidden)
			return
		}

		tmpl, err := template.ParseFiles(
			"templates/base.html",
			"templates/create-job.html",
		)
		if err != nil {
			http.Error(w, "failed to load template", http.StatusInternalServerError)
			return
		}

		data := struct {
			CompanyID int
		}{
			CompanyID: company.ID,
		}

		err = tmpl.ExecuteTemplate(w, "base", data)
		if err != nil {
			http.Error(w, "failed to render template", http.StatusInternalServerError)
			return
		}
	case http.MethodPost:
		// Read form values.
		companyID := strings.TrimSpace(r.FormValue("company_id"))
		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))
		requirements := strings.TrimSpace(r.FormValue("requirements"))
		location := strings.TrimSpace(r.FormValue("location"))
		employmentType := strings.TrimSpace(r.FormValue("employment_type"))
		salaryMin := strings.TrimSpace(r.FormValue("salary_min"))
		salaryMax := strings.TrimSpace(r.FormValue("salary_max"))
		deadlineStr := strings.TrimSpace(r.FormValue("deadline"))

		// Validate required fields.
		if companyID == "" ||
			title == "" ||
			description == "" ||
			requirements == "" ||
			location == "" ||
			employmentType == "" ||
			salaryMin == "" ||
			salaryMax == "" ||
			deadlineStr == "" {

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

		// The HTML form uses <input type="date">,
		// so the value comes as YYYY-MM-DD.
		deadline, err := time.Parse("2006-01-02", deadlineStr)
		if err != nil {
			http.Error(w, "invalid deadline", http.StatusBadRequest)
			return
		}

		job := models.Job{
			CompanyID:      compID,
			Title:          title,
			Description:    description,
			Requirements:   requirements,
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

func EditJobHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	jobID := strings.TrimPrefix(r.URL.Path, "/jobs/")
	jobID = strings.TrimSuffix(jobID, "/edit")
	jobIDInt, err := strconv.Atoi(jobID)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	job, err := database.GetJobByID(jobIDInt)
	if err != nil {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}
	company, err := database.GetCompanyByUserID(userID)
	if err != nil {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	if job.CompanyID != company.ID {
		http.Error(w, "forbiddent", http.StatusForbidden)
		return
	}
	if r.Method == http.MethodGet {
		funcMap := template.FuncMap{
			"formatDate": func(t time.Time) string {
				return t.Format("2006-01-02")
			},
		}

		tmpl, err := template.New("base.html").
			Funcs(funcMap).
			ParseFiles(
				"templates/base.html",
				"templates/edit-job.html",
			)
		if err != nil {
			http.Error(w, "failed to load template", http.StatusInternalServerError)
			return
		}
		err = tmpl.ExecuteTemplate(w, "base", job)
		if err != nil {
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	title := r.FormValue("title")
	description := r.FormValue("description")
	location := r.FormValue("location")
	employmentType := r.FormValue("employment_type")
	salaryMin := r.FormValue("salary_min")
	salaryMax := r.FormValue("salary_max")
	deadlineStr := r.FormValue("deadline")

	if title == "" || description == "" || location == "" || employmentType == "" || salaryMin == "" || salaryMax == "" || deadlineStr == "" {
		http.Error(w, "required feilds can not be empty", http.StatusBadRequest)
		return
	}
	minSalary, err := strconv.Atoi(salaryMin)
	if err != nil {
		http.Error(w, "invalid minimum salary", http.StatusBadRequest)
		return
	}
	maxSalary, err := strconv.Atoi(salaryMax)
	if err != nil {
		http.Error(w, "invalid maximum salary", http.StatusBadRequest)
		return
	}
	deadline, err := time.Parse("2006-01-02", deadlineStr)
	if err != nil {
		http.Error(w, "invalid deadline", http.StatusBadRequest)
		return
	}
	job.Title = title
	job.Description = description
	job.Location = location
	job.EmploymentType = employmentType
	job.SalaryMin = minSalary
	job.SalaryMax = maxSalary
	job.Deadline = deadline

	err = database.UpdateJob(job.ID, job)
	if err != nil {
		http.Error(w, "unable to update job", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/jobs/%d", job.ID), http.StatusSeeOther)
}
