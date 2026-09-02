package database

import (
	"fmt"
	"jobFlow/models"
)

func CreateApplication(application models.Application) (int, error) {
	query := `INSERT INTO applications (job_id, user_id, cover_letter, resume, status) VALUES (?, ?, ?, ?, ?)`
	res, err := DB.Exec(query, application.JobID, application.UserID, application.CoverLetter, application.Resume, application.Status)
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
	query := `SELECT id, job_id, user_id, cover_letter, resume, status, applied_at, updated_at FROM applications WHERE id = ?`
	var application models.Application
	err := DB.QueryRow(query, id).Scan(&application.ID,
		&application.JobID,
		&application.UserID,
		&application.CoverLetter,
		&application.Resume,
		&application.Status,
		&application.AppliedAt,
		&application.UpdatedAt)
	if err != nil {
		return models.Application{}, err
	}
	return application, nil
}

func GetApplicationByUserID(UserID int) ([]models.Application, error) {
	query := `SELECT id, job_id, user_id, cover_letter, resume, status, applied_at, updated_at FROM applications WHERE user_id = ?`

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
			&application.Resume,
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
	query := `SELECT id, job_id, user_id, cover_letter, resume, status, applied_at, updated_at FROM applications WHERE job_id = ?`

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
			&application.Resume,
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
