package database

import (
	"fmt"
	"jobFlow/models"
)

func CreatePost(post models.Post) (int, error) {
	result, err := DB.Exec(`
		INSERT INTO posts (
			user_id,
			content,
			type
		)
		VALUES (?, ?, ?)
	`,
		post.UserID,
		post.Content,
		post.Type,
	)

	if err != nil {
		return 0, fmt.Errorf("creating post: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting post ID: %w", err)
	}

	return int(id), nil
}

func GetPostByID(id int) (models.Post, error) {
	var post models.Post

	err := DB.QueryRow(`
		SELECT
			id,
			user_id,
			content,
			type,
			created_at,
			updated_at
		FROM posts
		WHERE id = ?
	`, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Content,
		&post.Type,
		&post.CreatedAt,
		&post.UpdatedAt,
	)

	if err != nil {
		return models.Post{}, fmt.Errorf("getting post: %w", err)
	}

	return post, nil
}

func GetPostsByUserID(userID int) ([]models.Post, error) {
	rows, err := DB.Query(`
		SELECT
			id,
			user_id,
			content,
			type,
			created_at,
			updated_at
		FROM posts
		WHERE user_id = ?
		ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("getting user posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post

	for rows.Next() {
		var post models.Post

		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&post.Type,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning post: %w", err)
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating posts: %w", err)
	}

	return posts, nil
}

func GetPosts() ([]models.Post, error) {
	rows, err := DB.Query(`
		SELECT
			id,
			user_id,
			content,
			type,
			created_at,
			updated_at
		FROM posts
		ORDER BY created_at DESC
	`)

	if err != nil {
		return nil, fmt.Errorf("getting posts: %w", err)
	}
	defer rows.Close()

	var posts []models.Post

	for rows.Next() {
		var post models.Post

		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Content,
			&post.Type,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning post: %w", err)
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating posts: %w", err)
	}

	return posts, nil
}

func UpdatePost(post models.Post) error {
	result, err := DB.Exec(`
		UPDATE posts
		SET
			content = ?,
			type = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		post.Content,
		post.Type,
		post.ID,
	)

	if err != nil {
		return fmt.Errorf("updating post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking updated post: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}

func DeletePost(id int) error {
	result, err := DB.Exec(`
		DELETE FROM posts
		WHERE id = ?
	`, id)

	if err != nil {
		return fmt.Errorf("deleting post: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking deleted post: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("post not found")
	}

	return nil
}
