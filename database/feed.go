package database

import (
	"database/sql"
	"fmt"

	"jobFlow/models"
)

func GetFeedPosts(userID int) ([]models.FeedPost, error) {
	rows, err := DB.Query(`
		SELECT
			p.id,
			p.user_id,
			p.content,
			p.type,
			p.shared_post_id,
			p.created_at,
			p.updated_at,

			u.first_name,
			u.last_name,
			u.username,

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

			COALESCE(original_user.first_name, '') AS original_first_name,
			COALESCE(original_user.last_name, '') AS original_last_name,
			COALESCE(original_user.username, '') AS original_username,
			COALESCE(original_post.content, '') AS original_content,
			COALESCE(original_post.type, '') AS original_type

		FROM posts p

		JOIN users u
			ON u.id = p.user_id

		LEFT JOIN posts original_post
			ON original_post.id = p.shared_post_id

		LEFT JOIN users original_user
			ON original_user.id = original_post.user_id

		ORDER BY p.created_at DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("getting feed posts: %w", err)
	}
	defer rows.Close()

	var posts []models.FeedPost

	for rows.Next() {
		var post models.FeedPost
		var sharedPostID sql.NullInt64

		err := rows.Scan(
			&post.ID,
			&post.UserID,
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

			&post.OriginalAuthorFirstName,
			&post.OriginalAuthorLastName,
			&post.OriginalAuthorUsername,
			&post.OriginalContent,
			&post.OriginalType,
		)

		if err != nil {
			return nil, fmt.Errorf("scanning feed post: %w", err)
		}

		if sharedPostID.Valid {
			id := int(sharedPostID.Int64)
			post.SharedPostID = &id
		}

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating feed posts: %w", err)
	}

	return posts, nil
}
