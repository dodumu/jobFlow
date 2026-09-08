package database

import (
	"fmt"
	"jobFlow/models"
)

func LikePost(like models.PostLike) error {
	_, err := DB.Exec(`
		INSERT INTO post_likes (
			post_id,
			user_id
		)
		VALUES (?, ?)
	`,
		like.PostID,
		like.UserID,
	)

	if err != nil {
		return fmt.Errorf("liking post: %w", err)
	}

	return nil
}

func UnlikePost(postID, userID int) error {
	result, err := DB.Exec(`
		DELETE FROM post_likes
		WHERE post_id = ? AND user_id = ?
	`, postID, userID)

	if err != nil {
		return fmt.Errorf("unliking post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking removed like: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("post was not liked by user")
	}

	return nil
}

func HasUserLikedPost(postID, userID int) (bool, error) {
	var exists bool

	err := DB.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM post_likes
			WHERE post_id = ? AND user_id = ?
		)
	`, postID, userID).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("checking post like: %w", err)
	}

	return exists, nil
}

func GetPostLikeCount(postID int) (int, error) {
	var count int

	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM post_likes
		WHERE post_id = ?
	`, postID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("getting post like count: %w", err)
	}

	return count, nil
}

func GetPostLikes(postID int) ([]models.User, error) {
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
		JOIN post_likes pl
			ON pl.user_id = u.id
		WHERE pl.post_id = ?
		ORDER BY pl.created_at DESC
	`, postID)

	if err != nil {
		return nil, fmt.Errorf("getting post likes: %w", err)
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
			return nil, fmt.Errorf("scanning post liker: %w", err)
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating post likes: %w", err)
	}

	return users, nil
}
