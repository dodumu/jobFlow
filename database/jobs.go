package database

import "jobFlow/models"

func CreateJob(job models.Job) (int, error) {
	query := `INSERT INTO jobs (company_id, title, description, requirements, location, employment_type, salary_min, salary_max, deadline, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	res, err := DB.Exec(query, job.CompanyID, job.Title, job.Description, job.Requirements, job.Location, job.EmploymentType, job.SalaryMin, job.SalaryMax, job.Deadline, job.Status)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func GetJobByID(id int) (models.Job, error) {
	query := `SELECT id, company_id, title, description, requirements, location, employment_type, salary_min, salary_max, deadline, status, created_at, updated_at
	FROM jobs
	WHERE id = ?`
	var job models.Job
	err := DB.QueryRow(query, id).Scan(&job.ID,
		&job.CompanyID,
		&job.Title,
		&job.Description,
		&job.Requirements,
		&job.Location,
		&job.EmploymentType,
		&job.SalaryMin,
		&job.SalaryMax,
		&job.Deadline,
		&job.Status,
		&job.CreatedAt,
		&job.UpdatedAt)
	if err != nil {
		return models.Job{}, err
	}
	return job, nil
}
