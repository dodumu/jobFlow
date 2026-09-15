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

func GetPostsByUserID(userID int) ([]models.FeedPost, error) {
	rows, err := DB.Query(`
		SELECT
			p.id,
			p.user_id,
			p.company_id,
			p.content,
			p.type,
			p.created_at,
			p.updated_at,

			u.first_name,
			u.last_name,
			u.username,

			COUNT(DISTINCT pl.user_id) AS like_count,
			COUNT(DISTINCT c.id) AS comment_count,
			COUNT(DISTINCT ps.id) AS share_count

		FROM posts p

		LEFT JOIN users u
			ON u.id = p.user_id

		LEFT JOIN post_likes pl
			ON pl.post_id = p.id

		LEFT JOIN comments c
			ON c.post_id = p.id

		LEFT JOIN post_shares ps
			ON ps.post_id = p.id

		WHERE p.user_id = ?

		GROUP BY p.id

		ORDER BY p.created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []models.FeedPost

	for rows.Next() {
		var post models.FeedPost

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.CompanyID,
			&post.Content,
			&post.Type,
			&post.CreatedAt,
			&post.UpdatedAt,

			&post.AuthorFirstName,
			&post.AuthorLastName,
			&post.AuthorUsername,

			&post.LikeCount,
			&post.CommentCount,
			&post.ShareCount,
		)

		if err != nil {
			return nil, err
		}

		post.AuthorType = "user"
		post.CanDelete = true

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
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
