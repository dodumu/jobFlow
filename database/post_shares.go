package database

import (
	"fmt"

	"jobFlow/models"
)

func createPostShare(
	exec DBExecutor,
	share models.PostShare,
) (int, error) {

	result, err := exec.Exec(`
		INSERT INTO post_shares (
			post_id,
			user_id
		)
		VALUES (?, ?)
	`,
		share.PostID,
		share.UserID,
	)

	if err != nil {
		return 0, fmt.Errorf("creating post share: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting post share ID: %w", err)
	}

	return int(id), nil
}

func CreatePostShare(share models.PostShare) (int, error) {
	return createPostShare(DB, share)
}
func SharePost(
	share models.PostShare,
	sharedPost models.Post,
) error {

	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("beginning share transaction: %w", err)
	}

	defer tx.Rollback()

	_, err = createPostShare(tx, share)
	if err != nil {
		return fmt.Errorf("creating post share: %w", err)
	}

	_, err = createPost(tx, sharedPost)
	if err != nil {
		return fmt.Errorf("creating shared post: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing share transaction: %w", err)
	}

	return nil
}
func GetPostShareByID(id int) (models.PostShare, error) {
	var share models.PostShare

	err := DB.QueryRow(`
		SELECT
			id,
			post_id,
			user_id,
			created_at
		FROM post_shares
		WHERE id = ?
	`, id).Scan(
		&share.ID,
		&share.PostID,
		&share.UserID,
		&share.CreatedAt,
	)

	if err != nil {
		return models.PostShare{}, fmt.Errorf("getting post share: %w", err)
	}

	return share, nil
}

func GetPostSharesByPostID(postID int) ([]models.PostShare, error) {
	rows, err := DB.Query(`
		SELECT
			id,
			post_id,
			user_id,
			created_at
		FROM post_shares
		WHERE post_id = ?
		ORDER BY created_at DESC
	`, postID)

	if err != nil {
		return nil, fmt.Errorf("getting post shares: %w", err)
	}
	defer rows.Close()

	var shares []models.PostShare

	for rows.Next() {
		var share models.PostShare

		if err := rows.Scan(
			&share.ID,
			&share.PostID,
			&share.UserID,
			&share.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning post share: %w", err)
		}

		shares = append(shares, share)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating post shares: %w", err)
	}

	return shares, nil
}

func GetPostShareCount(postID int) (int, error) {
	var count int

	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM post_shares
		WHERE post_id = ?
	`, postID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("getting post share count: %w", err)
	}

	return count, nil
}
