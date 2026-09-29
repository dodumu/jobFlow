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
	"strconv"
	"strings"
	"time"
)

func JobsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to view jobs.",
		)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("JobsHandler GetUserByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your account information.",
		)
		return
	}

	jobs, err := database.GetJobs()
	if err != nil {
		log.Printf("JobsHandler GetJobs error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load jobs right now.",
		)
		return
	}

	data := models.JobsPageData{
		Jobs:      jobs,
		IsCompany: user.Role == constants.RoleCompany,
	}

	if err := utils.RenderTemplate(w, "jobs.html", data); err != nil {
		log.Printf("JobsHandler RenderTemplate jobs.html error: %v", err)
		return
	}
}

func JobDetailsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	idStr := strings.TrimPrefix(r.URL.Path, "/jobs/")

	if idStr == "" {
		RenderError(
			w,
			http.StatusBadRequest,
			"Invalid Job",
			"A valid job ID is required.",
		)
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		RenderError(
			w,
			http.StatusBadRequest,
			"Invalid Job",
			"The job ID provided is invalid.",
		)
		return
	}

	job, err := database.GetJobByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RenderError(
				w,
				http.StatusNotFound,
				"Job Not Found",
				"The job you're looking for could not be found.",
			)
			return
		}

		log.Printf("JobDetailsHandler GetJobByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load this job right now.",
		)
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to view this job.",
		)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("JobDetailsHandler GetUserByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your account information.",
		)
		return
	}

	canEdit := false

	if user.Role == constants.RoleCompany {
		company, err := database.GetCompanyByUserID(userID)
		if err != nil {
			log.Printf("JobDetailsHandler GetCompanyByUserID error: %v", err)
		} else if company.ID == job.CompanyID {
			canEdit = true
		}
	}

	data := models.JobDetailsData{
		Job:      job,
		CanEdit:  canEdit,
		UserRole: user.Role,
	}

	if err := utils.RenderTemplate(w, "job.html", data); err != nil {
		log.Printf(
			"JobDetailsHandler RenderTemplate job.html error: %v",
			err,
		)
		return
	}
}
func CreateJobHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to create a job.",
		)
		return
	}

	company, err := database.GetCompanyByUserID(userID)
	if err != nil {
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"Only company accounts can create jobs.",
		)
		return
	}

	switch r.Method {
	case http.MethodGet:
		data := struct {
			CompanyID int
		}{
			CompanyID: company.ID,
		}

		if err := utils.RenderTemplate(
			w,
			"create-job.html",
			data,
		); err != nil {
			log.Printf(
				"CreateJobHandler RenderTemplate create-job.html error: %v",
				err,
			)
			return
		}

	case http.MethodPost:
		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))
		requirements := strings.TrimSpace(r.FormValue("requirements"))
		location := strings.TrimSpace(r.FormValue("location"))
		employmentType := strings.TrimSpace(r.FormValue("employment_type"))
		salaryMin := strings.TrimSpace(r.FormValue("salary_min"))
		salaryMax := strings.TrimSpace(r.FormValue("salary_max"))
		deadlineStr := strings.TrimSpace(r.FormValue("deadline"))

		if title == "" ||
			description == "" ||
			requirements == "" ||
			location == "" ||
			employmentType == "" ||
			salaryMin == "" ||
			salaryMax == "" ||
			deadlineStr == "" {

			RenderError(
				w,
				http.StatusBadRequest,
				"Missing Information",
				"Please complete all required job fields.",
			)
			return
		}

		minSalary, err := strconv.Atoi(salaryMin)
		if err != nil || minSalary < 0 {
			RenderError(
				w,
				http.StatusBadRequest,
				"Invalid Salary",
				"Minimum salary must be a valid non-negative number.",
			)
			return
		}

		maxSalary, err := strconv.Atoi(salaryMax)
		if err != nil || maxSalary < 0 {
			RenderError(
				w,
				http.StatusBadRequest,
				"Invalid Salary",
				"Maximum salary must be a valid non-negative number.",
			)
			return
		}

		if minSalary > maxSalary {
			RenderError(
				w,
				http.StatusBadRequest,
				"Invalid Salary Range",
				"Minimum salary cannot be greater than maximum salary.",
			)
			return
		}

		deadline, err := time.Parse("2006-01-02", deadlineStr)
		if err != nil {
			RenderError(
				w,
				http.StatusBadRequest,
				"Invalid Deadline",
				"Please provide a valid application deadline.",
			)
			return
		}

		job := models.Job{
			CompanyID:      company.ID,
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
			log.Printf("CreateJobHandler CreateJob error: %v", err)
			RenderError(
				w,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't create the job right now.",
			)
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/jobs/%d", newJob),
			http.StatusSeeOther,
		)

	default:
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
	}
}
func EditJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		log.Printf("Request Method: %v", r.Method)
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to edit a job.",
		)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/jobs/")
	jobID = strings.TrimSuffix(jobID, "/edit")

	jobIDInt, err := strconv.Atoi(jobID)
	if err != nil || jobIDInt <= 0 {
		RenderError(
			w,
			http.StatusBadRequest,
			"Invalid Job",
			"The job ID provided is invalid.",
		)
		return
	}

	job, err := database.GetJobByID(jobIDInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RenderError(
				w,
				http.StatusNotFound,
				"Job Not Found",
				"The job you're looking for could not be found.",
			)
			return
		}

		log.Printf("EditJobHandler GetJobByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load this job right now.",
		)
		return
	}

	company, err := database.GetCompanyByUserID(userID)
	if err != nil {
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"Only company accounts can edit jobs.",
		)
		return
	}

	if job.CompanyID != company.ID {
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"You don't have permission to edit this job.",
		)
		return
	}

	if r.Method == http.MethodGet {
		// funcMap := template.FuncMap{
		// 	"formatDate": func(t time.Time) string {
		// 		return t.Format("2006-01-02")
		// 	},
		// }

		err = utils.RenderTemplate(w, "edit-job.html", job)
		if err != nil {
			log.Printf(
				"ApplicationHandler RenderTemplate edit-job.html error: %v",
				err)
		}
		return
	}

	if r.Method == http.MethodPost {
		title := strings.TrimSpace(r.FormValue("title"))
		description := strings.TrimSpace(r.FormValue("description"))
		location := strings.TrimSpace(r.FormValue("location"))
		employmentType := strings.TrimSpace(r.FormValue("employment_type"))
		salaryMin := strings.TrimSpace(r.FormValue("salary_min"))
		salaryMax := strings.TrimSpace(r.FormValue("salary_max"))
		deadlineStr := strings.TrimSpace(r.FormValue("deadline"))

		if title == "" ||
			description == "" ||
			location == "" ||
			employmentType == "" ||
			salaryMin == "" ||
			salaryMax == "" ||
			deadlineStr == "" {

			RenderError(
				w,
				http.StatusBadRequest,
				"Missing Information",
				"Please complete all required job fields.",
			)
			return
		}

		minSalary, err := strconv.Atoi(salaryMin)
		if err != nil || minSalary < 0 {
			RenderError(
				w,
				http.StatusBadRequest,
				"Invalid Salary",
				"Minimum salary must be a valid non-negative number.",
			)
			return
		}

		maxSalary, err := strconv.Atoi(salaryMax)
		if err != nil || maxSalary < 0 {
			RenderError(
				w,
				http.StatusBadRequest,
				"Invalid Salary",
				"Maximum salary must be a valid non-negative number.",
			)
			return
		}

		if minSalary > maxSalary {
			RenderError(
				w,
				http.StatusBadRequest,
				"Invalid Salary Range",
				"Minimum salary cannot be greater than maximum salary.",
			)
			return
		}

		deadline, err := time.Parse("2006-01-02", deadlineStr)
		if err != nil {
			RenderError(
				w,
				http.StatusBadRequest,
				"Invalid Deadline",
				"Please provide a valid application deadline.",
			)
			return
		}

		job.Title = title
		job.Description = description
		job.Location = location
		job.EmploymentType = employmentType
		job.SalaryMin = minSalary
		job.SalaryMax = maxSalary
		job.Deadline = deadline

		if err := database.UpdateJob(job.ID, job); err != nil {
			log.Printf("EditJobHandler UpdateJob error: %v", err)
			RenderError(
				w,
				http.StatusInternalServerError,
				"Something Went Wrong",
				"We couldn't update this job right now.",
			)
			return
		}

		http.Redirect(
			w,
			r,
			fmt.Sprintf("/jobs/%d", job.ID),
			http.StatusSeeOther,
		)
	}
}
func CloseJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed for this action.",
		)
		return
	}
	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to close a job.",
		)
		return
	}

	jobID, err := utils.GetPathID(r, "id")
	if err != nil || jobID <= 0 {
		RenderError(
			w,
			http.StatusBadRequest,
			"Invalid Job",
			"The job ID provided is invalid.",
		)
		return
	}

	job, err := database.GetJobByID(jobID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RenderError(
				w,
				http.StatusNotFound,
				"Job Not Found",
				"The job you're looking for could not be found.",
			)
			return
		}

		log.Printf("CloseJobHandler GetJobByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load this job right now.",
		)
		return
	}

	company, err := database.GetCompanyByUserID(userID)
	if err != nil {
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"Only company accounts can close jobs.",
		)
		return
	}

	if job.CompanyID != company.ID {
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"You don't have permission to close this job.",
		)
		return
	}

	if job.Status == constants.JobStatusClosed {
		http.Redirect(
			w,
			r,
			fmt.Sprintf("/jobs/%d", job.ID),
			http.StatusSeeOther,
		)
		return
	}

	if err := database.CloseJob(job.ID); err != nil {
		log.Printf("CloseJobHandler CloseJob error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't close this job right now.",
		)
		return
	}

	http.Redirect(
		w,
		r,
		fmt.Sprintf("/jobs/%d", job.ID),
		http.StatusSeeOther,
	)
}
func ApplyJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		RenderError(
			w,
			http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"The requested method is not allowed on this page.",
		)
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(
			w,
			http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to apply for a job.",
		)
		return
	}

	jobID := strings.TrimPrefix(r.URL.Path, "/jobs/")
	jobID = strings.TrimSuffix(jobID, "/apply")

	jobIDInt, err := strconv.Atoi(jobID)
	if err != nil || jobIDInt <= 0 {
		RenderError(
			w,
			http.StatusBadRequest,
			"Invalid Job",
			"The job ID provided is invalid.",
		)
		return
	}

	job, err := database.GetJobByID(jobIDInt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			RenderError(
				w,
				http.StatusNotFound,
				"Job Not Found",
				"The job you're looking for could not be found.",
			)
			return
		}

		log.Printf("ApplyJobHandler GetJobByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load this job right now.",
		)
		return
	}

	if job.Status != constants.JobStatusOpen {
		RenderError(
			w,
			http.StatusBadRequest,
			"Applications Closed",
			"This job is no longer accepting applications.",
		)
		return
	}

	user, err := database.GetUserByID(userID)
	if err != nil {
		log.Printf("ApplyJobHandler GetUserByID error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your account information.",
		)
		return
	}

	if user.Role != constants.RoleIndividual {
		RenderError(
			w,
			http.StatusForbidden,
			"Access Denied",
			"Only individual accounts can apply for jobs.",
		)
		return
	}

	hasApplied, err := database.HasUserApplied(jobIDInt, userID)
	if err != nil {
		log.Printf("ApplyJobHandler HasUserApplied error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't check your application status.",
		)
		return
	}

	if hasApplied {
		RenderError(
			w,
			http.StatusBadRequest,
			"Already Applied",
			"You have already applied for this job.",
		)
		return
	}

	if r.Method == http.MethodGet {
		err = utils.RenderTemplate(w, "apply-job.html", job)
		if err != nil {
			log.Printf(
				"ApplicationHandler RenderTemplate applications.html error: %v",
				err)
		}

		return
	}

	coverLetter := strings.TrimSpace(r.FormValue("cover_letter"))

	if coverLetter == "" {
		RenderError(
			w,
			http.StatusBadRequest,
			"Cover Letter Required",
			"Please provide a cover letter before submitting your application.",
		)
		return
	}

	application := models.Application{
		JobID:       jobIDInt,
		UserID:      userID,
		CoverLetter: coverLetter,
		Status:      "pending",
	}

	_, err = database.CreateApplication(application)
	if err != nil {
		log.Printf("ApplyJobHandler CreateApplication error: %v", err)
		RenderError(
			w,
			http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't submit your application right now.",
		)
		return
	}

	http.Redirect(w, r, "/applications", http.StatusSeeOther)
}
