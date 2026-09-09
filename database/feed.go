package database

import (
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
			) AS has_liked

		FROM posts p
		JOIN users u
			ON u.id = p.user_id

		ORDER BY p.created_at DESC
	`, userID)

	if err != nil {
		return nil, fmt.Errorf("getting feed posts: %w", err)
	}
	defer rows.Close()

	var posts []models.FeedPost

	for rows.Next() {
		var post models.FeedPost

		err := rows.Scan(
			&post.ID,
			&post.UserID,
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
			&post.HasLiked,
		)

		if err != nil {
			return nil, fmt.Errorf("scanning feed post: %w", err)
		}
		comments, err := getFeedComments(post.ID)
		if err != nil {
			return nil, err
		}

		post.Comments = comments

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating feed posts: %w", err)
	}

	return posts, nil
}

func getFeedComments(postID int) ([]models.FeedComment, error) {
	rows, err := DB.Query(`
		SELECT
			c.id,
			c.post_id,
			c.user_id,
			c.content,
			c.created_at,

			u.first_name,
			u.last_name,
			u.username

		FROM comments c
		JOIN users u
			ON u.id = c.user_id

		WHERE c.post_id = ?

		ORDER BY c.created_at ASC
	`, postID)

	if err != nil {
		return nil, fmt.Errorf("getting feed comments: %w", err)
	}
	defer rows.Close()

	var comments []models.FeedComment

	for rows.Next() {
		var comment models.FeedComment

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.UserID,
			&comment.Content,
			&comment.CreatedAt,

			&comment.AuthorFirstName,
			&comment.AuthorLastName,
			&comment.AuthorUsername,
		)

		if err != nil {
			return nil, fmt.Errorf("scanning feed comment: %w", err)
		}

		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating feed comments: %w", err)
	}

	return comments, nil
}
