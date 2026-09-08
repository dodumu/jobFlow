package database

import (
	"fmt"

	"jobFlow/models"
)

func CreateFollow(follow models.Follow) error {
	_, err := DB.Exec(`
		INSERT INTO follows (
			follower_id,
			following_id
		)
		VALUES (?, ?)
	`,
		follow.FollowerID,
		follow.FollowingID,
	)

	if err != nil {
		return fmt.Errorf("creating follow: %w", err)
	}

	return nil
}

func DeleteFollow(followerID, followingID int) error {
	result, err := DB.Exec(`
		DELETE FROM follows
		WHERE follower_id = ? AND following_id = ?
	`, followerID, followingID)

	if err != nil {
		return fmt.Errorf("deleting follow: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking deleted follow: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("follow relationship not found")
	}

	return nil
}

func IsFollowing(followerID, followingID int) (bool, error) {
	var exists bool

	err := DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM follows
			WHERE follower_id = ? AND following_id = ?
		)
	`, followerID, followingID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("checking follow relationship: %w", err)
	}

	return exists, nil
}

func GetFollowers(userID int) ([]models.User, error) {
	rows, err := DB.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			u.email,
			u.date_of_birth,
			u.role,
			u.created_at
		FROM users u
		JOIN follows f
			ON f.follower_id = u.id
		WHERE f.following_id = ?
		ORDER BY f.created_at DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("getting followers: %w", err)
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User

		if err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Username,
			&user.Email,
			&user.DOB,
			&user.Role,
			&user.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning follower: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating followers: %w", err)
	}

	return users, nil
}

func GetFollowing(userID int) ([]models.User, error) {
	rows, err := DB.Query(`
		SELECT
			u.id,
			u.first_name,
			u.last_name,
			u.username,
			u.email,
			u.date_of_birth,
			u.role,
			u.created_at
		FROM users u
		JOIN follows f
			ON f.following_id = u.id
		WHERE f.follower_id = ?
		ORDER BY f.created_at DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("getting following users: %w", err)
	}
	defer rows.Close()

	var users []models.User

	for rows.Next() {
		var user models.User

		if err := rows.Scan(
			&user.ID,
			&user.FirstName,
			&user.LastName,
			&user.Username,
			&user.Email,
			&user.DOB,
			&user.Role,
			&user.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning following user: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating following users: %w", err)
	}

	return users, nil
}

func GetFollowerCount(userID int) (int, error) {
	var count int

	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM follows
		WHERE following_id = ?
	`, userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("getting follower count: %w", err)
	}

	return count, nil
}

func GetFollowingCount(userID int) (int, error) {
	var count int

	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM follows
		WHERE follower_id = ?
	`, userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("getting following count: %w", err)
	}

	return count, nil
}
