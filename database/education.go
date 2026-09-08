package database

import (
	"fmt"
	"jobFlow/models"
)

func CreateEducation(education models.Education) (int, error) {
	result, err := DB.Exec(`
		INSERT INTO education (
			user_id,
			institution,
			degree,
			field_of_study,
			start_date,
			end_date,
			description
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		education.UserID,
		education.Institution,
		education.Degree,
		education.FieldOfStudy,
		education.StartDate,
		education.EndDate,
		education.Description,
	)

	if err != nil {
		return 0, fmt.Errorf("creating education: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting education ID: %w", err)
	}

	return int(id), nil
}

func GetEducationByID(id int) (models.Education, error) {
	var education models.Education

	err := DB.QueryRow(`
		SELECT
			id,
			user_id,
			institution,
			degree,
			field_of_study,
			start_date,
			end_date,
			description,
			created_at,
			updated_at
		FROM education
		WHERE id = ?
	`, id).Scan(
		&education.ID,
		&education.UserID,
		&education.Institution,
		&education.Degree,
		&education.FieldOfStudy,
		&education.StartDate,
		&education.EndDate,
		&education.Description,
		&education.CreatedAt,
		&education.UpdatedAt,
	)

	if err != nil {
		return models.Education{}, fmt.Errorf("getting education: %w", err)
	}

	return education, nil
}

func GetEducationByUserID(userID int) ([]models.Education, error) {
	rows, err := DB.Query(`
		SELECT
			id,
			user_id,
			institution,
			degree,
			field_of_study,
			start_date,
			end_date,
			description,
			created_at,
			updated_at
		FROM education
		WHERE user_id = ?
		ORDER BY start_date DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("getting user education: %w", err)
	}
	defer rows.Close()

	var educationList []models.Education

	for rows.Next() {
		var education models.Education

		err := rows.Scan(
			&education.ID,
			&education.UserID,
			&education.Institution,
			&education.Degree,
			&education.FieldOfStudy,
			&education.StartDate,
			&education.EndDate,
			&education.Description,
			&education.CreatedAt,
			&education.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scanning education: %w", err)
		}

		educationList = append(educationList, education)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating education: %w", err)
	}

	return educationList, nil
}
