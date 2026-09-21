package handlers

import (
	"database/sql"
	"errors"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"jobFlow/database"
	"jobFlow/models"
)

func JobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	jobs, err := database.GetJobs()
	if err != nil {
		http.Error(w, "failed to load jobs", http.StatusInternalServerError)
		return
	}

	data := struct {
		Jobs []models.Job
	}{
		Jobs: jobs,
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
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	companyID := r.FormValue("company_id")
	title := r.FormValue("title")
	description := r.FormValue("description")
	location := r.FormValue("location")
	employmentType := r.FormValue("employment_type")
	salaryMin := r.FormValue("salary_min")
	salaryMax := r.FormValue("salary_max")
	deadline := r.FormValue("deadline")
	status := r.FormValue("status")

	compID, err := strconv.Atoi(companyID)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	minSalary, err := strconv.Atoi(companyID)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	maxSalary, err := strconv.Atoi(companyID)
	if err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
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
		Status:         status,
	}
}
