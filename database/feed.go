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
			p.company_id,
			p.content,
			p.type,
			p.shared_post_id,
			p.created_at,
			p.updated_at,

			CASE
				WHEN p.company_id IS NOT NULL THEN 'company'
				ELSE 'user'
			END AS author_type,

			CASE
				WHEN p.user_id = ? THEN 1
				WHEN p.company_id IS NOT NULL AND company.user_id = ? THEN 1
				ELSE 0
			END AS can_delete,

			-- Individual author
			COALESCE(u.first_name, '') AS author_first_name,
			COALESCE(u.last_name, '') AS author_last_name,
			COALESCE(u.username, '') AS author_username,

			-- Company author
			COALESCE(company.company_name, '') AS company_name,
			COALESCE(company.logo, '') AS company_logo,

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

			COALESCE(original_post.content, '') AS original_content,
			COALESCE(original_post.type, '') AS original_type

		FROM posts p

		LEFT JOIN users u
			ON u.id = p.user_id

		LEFT JOIN companies company
			ON company.id = p.company_id

		LEFT JOIN posts original_post
			ON original_post.id = p.shared_post_id

		LEFT JOIN users original_user
			ON original_user.id = original_post.user_id

		LEFT JOIN companies original_company
			ON original_company.id = original_post.company_id

		ORDER BY p.created_at DESC
	`, userID, userID, userID)

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
			&post.CompanyID,
			&post.Content,
			&post.Type,
			&sharedPostID,
			&post.CreatedAt,
			&post.UpdatedAt,

			&post.AuthorType,
			&post.CanDelete,

			&post.AuthorFirstName,
			&post.AuthorLastName,
			&post.AuthorUsername,

			&post.CompanyName,
			&post.CompanyLogo,

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
