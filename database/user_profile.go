package database

import (
	"fmt"
	"jobFlow/models"
)

func CreateUserProfile(profile models.UserProfile) (int, error) {
	result, err := DB.Exec(`
		INSERT INTO user_profiles (
			user_id,
			profile_picture,
			headline,
			bio,
			location
		)
		VALUES (?, ?, ?, ?, ?)
	`,
		profile.UserID,
		profile.ProfilePicture,
		profile.Headline,
		profile.Bio,
		profile.Location,
	)

	if err != nil {
		return 0, fmt.Errorf("creating user profile: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting user profile ID: %w", err)
	}

	return int(id), nil
}

func GetUserProfileByUserID(userID int) (models.UserProfile, error) {
	var profile models.UserProfile

	err := DB.QueryRow(`
		SELECT
			id,
			user_id,
			profile_picture,
			headline,
			bio,
			location,
			created_at,
			updated_at
		FROM user_profiles
		WHERE user_id = ?
	`, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.ProfilePicture,
		&profile.Headline,
		&profile.Bio,
		&profile.Location,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		return models.UserProfile{}, fmt.Errorf("getting user profile: %w", err)
	}

	return profile, nil
}

func UpdateUserProfile(profile models.UserProfile) error {
	_, err := DB.Exec(`
		UPDATE user_profiles
		SET
			profile_picture = ?,
			headline = ?,
			bio = ?,
			location = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE user_id = ?
	`,
		profile.ProfilePicture,
		profile.Headline,
		profile.Bio,
		profile.Location,
		profile.UserID,
	)

	if err != nil {
		return fmt.Errorf("updating user profile: %w", err)
	}

	return nil
}
