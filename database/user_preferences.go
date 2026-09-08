package database

import (
	"fmt"

	"jobFlow/models"
)

func CreateUserPreference(preference models.UserPreference) (int, error) {
	result, err := DB.Exec(`
		INSERT INTO user_preferences (
			user_id,
			salary_min,
			salary_max
		)
		VALUES (?, ?, ?)
	`,
		preference.UserID,
		preference.SalaryMin,
		preference.SalaryMax,
	)

	if err != nil {
		return 0, fmt.Errorf("creating user preference: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting user preference ID: %w", err)
	}

	return int(id), nil
}

func GetUserPreferenceByUserID(userID int) (models.UserPreference, error) {
	var preference models.UserPreference

	err := DB.QueryRow(`
		SELECT
			id,
			user_id,
			salary_min,
			salary_max,
			created_at,
			updated_at
		FROM user_preferences
		WHERE user_id = ?
	`, userID).Scan(
		&preference.ID,
		&preference.UserID,
		&preference.SalaryMin,
		&preference.SalaryMax,
		&preference.CreatedAt,
		&preference.UpdatedAt,
	)

	if err != nil {
		return models.UserPreference{}, fmt.Errorf(
			"getting user preference: %w",
			err,
		)
	}

	return preference, nil
}

func UpdateUserPreference(preference models.UserPreference) error {
	result, err := DB.Exec(`
		UPDATE user_preferences
		SET
			salary_min = ?,
			salary_max = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
	`,
		preference.SalaryMin,
		preference.SalaryMax,
		preference.UserID,
	)

	if err != nil {
		return fmt.Errorf("updating user preference: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking updated preference: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user preference not found")
	}

	return nil
}

func AddEmploymentTypeToUser(userID, employmentTypeID int) error {
	_, err := DB.Exec(`
		INSERT INTO user_employment_types (
			user_id,
			employment_type_id
		)
		VALUES (?, ?)
	`,
		userID,
		employmentTypeID,
	)

	if err != nil {
		return fmt.Errorf("adding employment type to user: %w", err)
	}

	return nil
}

func RemoveEmploymentTypeFromUser(userID, employmentTypeID int) error {
	result, err := DB.Exec(`
		DELETE FROM user_employment_types
		WHERE user_id = ? AND employment_type_id = ?
	`,
		userID,
		employmentTypeID,
	)

	if err != nil {
		return fmt.Errorf("removing employment type from user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking removed employment type: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("employment type not assigned to user")
	}

	return nil
}

func GetEmploymentTypesByUserID(userID int) ([]string, error) {
	rows, err := DB.Query(`
		SELECT
			et.name
		FROM employment_types et
		JOIN user_employment_types uet
			ON uet.employment_type_id = et.id
		WHERE uet.user_id = ?
		ORDER BY et.name ASC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("getting user employment types: %w", err)
	}
	defer rows.Close()

	var employmentTypes []string

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf(
				"scanning employment type: %w",
				err,
			)
		}

		employmentTypes = append(employmentTypes, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterating employment types: %w",
			err,
		)
	}

	return employmentTypes, nil
}

func AddWorkArrangementToUser(userID, workArrangementID int) error {
	_, err := DB.Exec(`
		INSERT INTO user_work_arrangements (
			user_id,
			work_arrangement_id
		)
		VALUES (?, ?)
	`,
		userID,
		workArrangementID,
	)

	if err != nil {
		return fmt.Errorf("adding work arrangement to user: %w", err)
	}

	return nil
}

func RemoveWorkArrangementFromUser(userID, workArrangementID int) error {
	result, err := DB.Exec(`
		DELETE FROM user_work_arrangements
		WHERE user_id = ? AND work_arrangement_id = ?
	`,
		userID,
		workArrangementID,
	)

	if err != nil {
		return fmt.Errorf("removing work arrangement from user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking removed work arrangement: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("work arrangement not assigned to user")
	}

	return nil
}

func GetWorkArrangementsByUserID(userID int) ([]string, error) {
	rows, err := DB.Query(`
		SELECT
			wa.name
		FROM work_arrangements wa
		JOIN user_work_arrangements uwa
			ON uwa.work_arrangement_id = wa.id
		WHERE uwa.user_id = ?
		ORDER BY wa.name ASC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("getting user work arrangements: %w", err)
	}
	defer rows.Close()

	var arrangements []string

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf(
				"scanning work arrangement: %w",
				err,
			)
		}

		arrangements = append(arrangements, name)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterating work arrangements: %w",
			err,
		)
	}

	return arrangements, nil
}
