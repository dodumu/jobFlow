package database

import (
	"fmt"

	"jobFlow/models"
)

func CreateComment(comment models.Comment) (int, error) {
	result, err := DB.Exec(`
		INSERT INTO comments (
			post_id,
			user_id,
			content
		)
		VALUES (?, ?, ?)
	`,
		comment.PostID,
		comment.UserID,
		comment.Content,
	)

	if err != nil {
		return 0, fmt.Errorf("creating comment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting comment ID: %w", err)
	}

	return int(id), nil
}

func GetCommentByID(id int) (models.Comment, error) {
	var comment models.Comment

	err := DB.QueryRow(`
		SELECT
			id,
			post_id,
			user_id,
			content,
			created_at
		FROM comments
		WHERE id = ?
	`, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.UserID,
		&comment.Content,
		&comment.CreatedAt,
	)

	if err != nil {
		return models.Comment{}, fmt.Errorf("getting comment: %w", err)
	}

	return comment, nil
}

func GetCommentsByPostID(postID int) ([]models.Comment, error) {
	rows, err := DB.Query(`
		SELECT
			c.id,
			c.post_id,
			c.user_id,
			u.username,
			c.content,
			c.created_at
		FROM comments c
		JOIN users u
			ON u.id = c.user_id
		WHERE c.post_id = ?
		ORDER BY c.created_at ASC
	`, postID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var comments []models.Comment

	for rows.Next() {
		var comment models.Comment

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Username,
			&comment.Content,
			&comment.CreatedAt,
		)

		if err != nil {
			return nil, err
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func DeleteComment(id int) error {
	result, err := DB.Exec(`
		DELETE FROM comments
		WHERE id = ?
	`, id)

	if err != nil {
		return fmt.Errorf("deleting comment: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking deleted comment: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("comment not found")
	}

	return nil
}
