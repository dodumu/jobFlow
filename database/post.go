package database

import (
	"database/sql"
	"fmt"
	"jobFlow/models"
)

func createPost(
	exec DBExecutor,
	post models.Post,
) (int, error) {

	result, err := exec.Exec(`
		INSERT INTO posts (
			user_id,
			content,
			type,
			shared_post_id
		)
		VALUES (?, ?, ?, ?)
	`,
		post.UserID,
		post.Content,
		post.Type,
		post.SharedPostID,
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

func CreatePost(post models.Post) (int, error) {
	return createPost(DB, post)
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
			p.shared_post_id,
			p.created_at,
			p.updated_at,

			-- Current post author
			COALESCE(u.first_name, '') AS author_first_name,
			COALESCE(u.last_name, '') AS author_last_name,
			COALESCE(u.username, '') AS author_username,

			-- Engagement
			(
				SELECT COUNT(*)
				FROM post_likes pl
				WHERE pl.post_id = p.id
			) AS like_count,

			(
				SELECT COUNT(*)
				FROM comments c
				WHERE c.post_id = p.id
			) AS comment_count,

			(
				SELECT COUNT(*)
				FROM post_shares ps
				WHERE ps.post_id = p.id
			) AS share_count,

			EXISTS (
				SELECT 1
				FROM post_likes ul
				WHERE ul.post_id = p.id
				AND ul.user_id = ?
			) AS has_liked,

			-- Original post author type
			CASE
				WHEN original_post.company_id IS NOT NULL THEN 'company'
				WHEN original_post.user_id IS NOT NULL THEN 'user'
				ELSE ''
			END AS original_author_type,

			-- Original individual author
			COALESCE(original_user.first_name, '') AS original_first_name,
			COALESCE(original_user.last_name, '') AS original_last_name,
			COALESCE(original_user.username, '') AS original_username,

			-- Original company author
			COALESCE(original_company.company_name, '') AS original_company_name,
			COALESCE(original_company.logo, '') AS original_company_logo,

			-- Original post
			COALESCE(original_post.content, '') AS original_content,
			COALESCE(original_post.type, '') AS original_type

		FROM posts p

		LEFT JOIN users u
			ON u.id = p.user_id

		LEFT JOIN posts original_post
			ON original_post.id = p.shared_post_id

		LEFT JOIN users original_user
			ON original_user.id = original_post.user_id

		LEFT JOIN companies original_company
			ON original_company.id = original_post.company_id

		WHERE p.user_id = ?

		ORDER BY p.created_at DESC
	`, userID, userID)

	if err != nil {
		return nil, fmt.Errorf(
			"getting posts by user ID: %w",
			err,
		)
	}
	defer rows.Close()

	var posts []models.FeedPost

	for rows.Next() {
		var post models.FeedPost
		var sharedPostID sql.NullInt64

		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.CompanyID,
			&post.Content,
			&post.Type,
			&sharedPostID,
			&post.CreatedAt,
			&post.UpdatedAt,

			&post.AuthorFirstName,
			&post.AuthorLastName,
			&post.AuthorUsername,

			&post.LikeCount,
			&post.CommentCount,
			&post.ShareCount,
			&post.HasLiked,

			&post.OriginalAuthorType,

			&post.OriginalAuthorFirstName,
			&post.OriginalAuthorLastName,
			&post.OriginalAuthorUsername,

			&post.OriginalCompanyName,
			&post.OriginalCompanyLogo,

			&post.OriginalContent,
			&post.OriginalType,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"scanning user post: %w",
				err,
			)
		}

		if sharedPostID.Valid {
			id := int(sharedPostID.Int64)
			post.SharedPostID = &id
		}

		post.AuthorType = "user"
		post.CanDelete = true

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterating user posts: %w",
			err,
		)
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
