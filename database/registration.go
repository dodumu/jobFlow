package database

import (
	"fmt"
	"jobFlow/models"
)

func CreateAccount(
	user models.User,
	profile models.UserProfile,
	company *models.Company,
) (int, error) {

	tx, err := DB.Begin()
	if err != nil {
		return 0, fmt.Errorf(
			"beginning registration transaction: %w",
			err,
		)
	}

	userID, err := createUser(tx, user)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("creating user: %w", err)
	}

	profile.UserID = userID

	_, err = createUserProfile(tx, profile)
	if err != nil {
		tx.Rollback()
		return 0, fmt.Errorf("creating user profile: %w", err)
	}

	if company != nil {
		company.UserID = userID

		_, err = createCompany(tx, *company)
		if err != nil {
			tx.Rollback()
			return 0, fmt.Errorf("creating company: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf(
			"committing registration transaction: %w",
			err,
		)
	}

	return userID, nil
}
