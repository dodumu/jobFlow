package database

import (
	"fmt"
	"jobFlow/models"
)

func CreateApplication(application models.Application) (int, error) {
	query := `
		INSERT INTO applications (
			job_id,
			user_id,
			cover_letter,
			status
		)
		VALUES (?, ?, ?, ?)
	`

	res, err := DB.Exec(
		query,
		application.JobID,
		application.UserID,
		application.CoverLetter,
		application.Status,
	)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(id), nil
}

func GetApplicationByID(id int) (models.Application, error) {
	query := `SELECT id, job_id, user_id, cover_letter, status, applied_at, updated_at FROM applications WHERE id = ?`
	var application models.Application
	err := DB.QueryRow(query, id).Scan(&application.ID,
		&application.JobID,
		&application.UserID,
		&application.CoverLetter,
		&application.Status,
		&application.AppliedAt,
		&application.UpdatedAt)
	if err != nil {
		return models.Application{}, err
	}
	return application, nil
}

func GetApplicationByUserID(UserID int) ([]models.Application, error) {
	query := `SELECT id, job_id, user_id, cover_letter, status, applied_at, updated_at FROM applications WHERE user_id = ?`

	res, err := DB.Query(query, UserID)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var applications []models.Application
	for res.Next() {
		var application models.Application
		err := res.Scan(&application.ID,
			&application.JobID,
			&application.UserID,
			&application.CoverLetter,
			&application.Status,
			&application.AppliedAt,
			&application.UpdatedAt)
		if err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}
	err = res.Err()
	if err != nil {
		return nil, err
	}
	return applications, nil
}

func GetApplicationsByJobID(JobID int) ([]models.Application, error) {
	query := `SELECT id, job_id, user_id, cover_letter, status, applied_at, updated_at FROM applications WHERE job_id = ?`

	res, err := DB.Query(query, JobID)
	if err != nil {
		return nil, err
	}
	defer res.Close()
	var applications []models.Application
	for res.Next() {
		var application models.Application
		err := res.Scan(&application.ID,
			&application.JobID,
			&application.UserID,
			&application.CoverLetter,
			&application.Status,
			&application.AppliedAt,
			&application.UpdatedAt)
		if err != nil {
			return nil, err
		}
		applications = append(applications, application)
	}
	err = res.Err()
	if err != nil {
		return nil, err
	}
	return applications, nil
}

func UpdateApplicationStatus(id int, status string) error {
	query := `UPDATE applications
	SET status = ?, updated_at = CURRENT_TIMESTAMP
	WHERE id = ?`

	res, err := DB.Exec(query, status, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("application with id %d not found", id)
	}
	return nil
}

func WithdrawApplication(id int) error {
	query := `UPDATE  applications 
	SET status = 'withdrawn', updated_at = CURRENT_TIMESTAMP
	WHERE id = ?`
	res, err := DB.Exec(query, id)
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return fmt.Errorf("application with id %d not found", id)
	}
	return nil
}

func HasUserApplied(jobID int, userID int) (bool, error) {
	var exists bool

	query := `
		SELECT EXISTS(
			SELECT 1
			FROM applications
			WHERE job_id = ?
			AND user_id = ?
		)
	`

	err := DB.QueryRow(query, jobID, userID).Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func GetApplicationsByUserID(userID int) ([]models.UserApplication, error) {
	query := `
		SELECT
			a.id,
			a.job_id,
			j.title,
			j.location,
			j.employment_type,
			a.status,
			a.applied_at
		FROM applications a
		INNER JOIN jobs j ON a.job_id = j.id
		WHERE a.user_id = ?
		ORDER BY a.applied_at DESC
	`

	rows, err := DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []models.UserApplication

	for rows.Next() {
		var application models.UserApplication

		err := rows.Scan(
			&application.ApplicationID,
			&application.JobID,
			&application.JobTitle,
			&application.Location,
			&application.EmploymentType,
			&application.Status,
			&application.AppliedAt,
		)
		if err != nil {
			return nil, err
		}

		applications = append(applications, application)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applications, nil
}

func GetApplicationsByCompanyID(companyID int) ([]models.CompanyApplication, error) {
	query := `
		SELECT
			a.id,
			j.id,
			j.title,
			u.id,
			u.username,
			a.status,
			a.applied_at
		FROM applications a
		INNER JOIN jobs j ON a.job_id = j.id
		INNER JOIN users u ON a.user_id = u.id
		WHERE j.company_id = ?
		ORDER BY a.applied_at DESC
	`

	rows, err := DB.Query(query, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []models.CompanyApplication

	for rows.Next() {
		var application models.CompanyApplication

		err := rows.Scan(
			&application.ApplicationID,
			&application.JobID,
			&application.JobTitle,
			&application.UserID,
			&application.ApplicantName,
			&application.Status,
			&application.AppliedAt,
		)
		if err != nil {
			return nil, err
		}

		applications = append(applications, application)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return applications, nil
}
