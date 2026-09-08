package database

import (
	"fmt"
	"jobFlow/models"
)

func CreateExperience(experience models.Experience) (int, error) {
	result, err := DB.Exec(`
		INSERT INTO experiences (
			user_id,
			company,
			position,
			start_date,
			end_date,
			description,
			currently_working
		)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		experience.UserID,
		experience.Company,
		experience.Position,
		experience.StartDate,
		experience.EndDate,
		experience.Description,
		experience.CurrentlyWork,
	)

	if err != nil {
		return 0, fmt.Errorf("creating experience: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting experience ID: %w", err)
	}

	return int(id), nil
}

func GetExperienceByID(id int) (models.Experience, error) {
	var experience models.Experience

	err := DB.QueryRow(`
		SELECT
			id,
			user_id,
			company,
			position,
			start_date,
			end_date,
			description,
			currently_working,
			created_at,
			updated_at
		FROM experiences
		WHERE id = ?
	`, id).Scan(
		&experience.ID,
		&experience.UserID,
		&experience.Company,
		&experience.Position,
		&experience.StartDate,
		&experience.EndDate,
		&experience.Description,
		&experience.CurrentlyWork,
		&experience.CreatedAt,
		&experience.UpdatedAt,
	)

	if err != nil {
		return models.Experience{}, fmt.Errorf("getting experience: %w", err)
	}

	return experience, nil
}

func GetExperiencesByUserID(userID int) ([]models.Experience, error) {
	rows, err := DB.Query(`
		SELECT
			id,
			user_id,
			company,
			position,
			start_date,
			end_date,
			description,
			currently_working,
			created_at,
			updated_at
		FROM experiences
		WHERE user_id = ?
		ORDER BY start_date DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("getting user experiences: %w", err)
	}
	defer rows.Close()

	var experiences []models.Experience

	for rows.Next() {
		var experience models.Experience

		err := rows.Scan(
			&experience.ID,
			&experience.UserID,
			&experience.Company,
			&experience.Position,
			&experience.StartDate,
			&experience.EndDate,
			&experience.Description,
			&experience.CurrentlyWork,
			&experience.CreatedAt,
			&experience.UpdatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("scanning experience: %w", err)
		}

		experiences = append(experiences, experience)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating experiences: %w", err)
	}

	return experiences, nil
}
