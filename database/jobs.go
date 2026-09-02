package database

import (
	"fmt"
	"jobFlow/models"
)

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

func GetJobsByCompanyID(companyID int) ([]models.Job, error) {
	query := `SELECT id, company_id, title, description, requirements, location, employment_type, salary_min, salary_max, deadline, status, created_at, updated_at
	FROM jobs
	WHERE company_id = ?`

	var jobs []models.Job

	resp, err := DB.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	for resp.Next() {
		var job models.Job
		err := resp.Scan(&job.ID,
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
			return nil, err
		}
		jobs = append(jobs, job)
	}
	err = resp.Err()
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func GetJobs() ([]models.Job, error) {
	query := `SELECT id, company_id, title, description, requirements, location, employment_type, salary_min, salary_max, deadline, status, created_at, updated_at
	FROM jobs`
	var jobs []models.Job

	resp, err := DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer resp.Close()
	for resp.Next() {
		var job models.Job
		err := resp.Scan(&job.ID,
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
			return nil, err
		}
		jobs = append(jobs, job)
	}
	err = resp.Err()
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func UpdateJob(id int, job models.Job) error {
	query := `UPDATE jobs
	SET title = ?, description = ?, requirements = ?, location = ?, employment_type = ?, salary_min = ?, salary_max = ?, deadline = ?, status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	res, err := DB.Exec(query, job.Title, job.Description, job.Requirements, job.Location, job.EmploymentType, job.SalaryMin, job.SalaryMax, job.Deadline, job.Status, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("job with id %d not found", id)
	}
	return nil
}

func CloseJob(id int) error {
	query := `UPDATE jobs SET status = 'closed', updated_at = CURRENT_TIMESTAMP WHERE id = ? AND status = 'open'`
	res, err := DB.Exec(query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("job with id %d not found", id)
	}
	return nil
}
