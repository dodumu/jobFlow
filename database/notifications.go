package database

import (
	"database/sql"
	"fmt"
	"jobFlow/models"
)

func CreateNotification(userID int64, actorID sql.NullInt64, notificationType, message, link string) (int64, error) {

	var linkValue any
	if link != "" {
		linkValue = link
	}
	result, err := DB.Exec(`
		INSERT INTO notifications (
			user_id,
			actor_id,
			notification_type,
			message,
			link
		)
		VALUES (?, ?, ?, ?, ?)
	`, userID, actorID, notificationType, message, linkValue)
	if err != nil {
		return 0, fmt.Errorf("creating notification: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting notification ID: %w", err)
	}

	return id, nil
}

func GetUserNotifications(userID int64) ([]models.Notification, error) {
	rows, err := DB.Query(`
		SELECT
			id,
			user_id,
			actor_id,
			notification_type,
			message,
			link,
			is_read,
			created_at
		FROM notifications
		WHERE user_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 20
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("getting user notifications: %w", err)
	}
	defer rows.Close()

	var notifications []models.Notification

	for rows.Next() {
		var notification models.Notification

		err := rows.Scan(
			&notification.ID,
			&notification.UserID,
			&notification.ActorID,
			&notification.Type,
			&notification.Message,
			&notification.Link,
			&notification.IsRead,
			&notification.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning notification: %w", err)
		}

		notifications = append(notifications, notification)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating notifications: %w", err)
	}

	return notifications, nil
}

func GetUnreadNotificationCount(userID int64) (int, error) {
	var count int

	err := DB.QueryRow(`
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = ?
		AND is_read = FALSE
	`, userID).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("getting unread notification count: %w", err)
	}
	return count, nil
}

func MarkNotificationAsRead(notificationID, userID int64) error {
	result, err := DB.Exec(`
		UPDATE notifications
		SET is_read = TRUE
		WHERE id = ?
		AND user_id = ?
	`, notificationID, userID)
	if err != nil {
		return fmt.Errorf("marking notification as read: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("checking updated notification: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("notification not found for this user")
	}

	return nil
}
