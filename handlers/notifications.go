package handlers

import (
	"encoding/json"
	"fmt"
	"jobFlow/database"
	"jobFlow/models"
	"jobFlow/utils"
	"log"
	"net/http"
	"strconv"
	"time"
)

func NotificationsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, r, http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"This page only accept GET requests.")
		return
	}
	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(w, r, http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to view notifications")
		return
	}
	notifications, err := database.GetUserNotifications(int64(userID))
	if err != nil {
		log.Printf("NotificationsHandler: %v", err)
		RenderError(w, r, http.StatusInternalServerError,
			"Something Went Wrong",
			"We couldn't load your notifications.")
		return
	}
	data := struct {
		Notification []models.Notification
	}{
		Notification: notifications,
	}

	if err := utils.RenderTemplate(w, r, "notifications.html", data); err != nil {
		log.Printf("NotificationsHandler template error: %v", err)
	}
}

func UnreadNotificationCountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		RenderError(w, r, http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"This page only accept GET requests.")
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		RenderError(w, r, http.StatusUnauthorized,
			"Unauthorized",
			"You must be logged in to view notifications")
		return
	}

	count, err := database.GetUnreadNotificationCount(int64(userID))
	if err != nil {
		log.Printf("UnreadNotificationCountHanlder: %v", err)
		http.Error(w, "Unable to get unread count", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(struct {
		Count int `json:"count"`
	}{
		Count: count,
	}); err != nil {
		log.Printf("UnreadNotificationCountHandler response error %v", err)
	}
}

func MarkNotificationAsReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		RenderError(w, r, http.StatusMethodNotAllowed,
			"Method Not Allowed",
			"This page only accept POST requests.")
		return
	}

	UserID, ok := utils.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form submission", http.StatusBadRequest)
		return
	}

	notificationID, err := strconv.ParseInt(
		r.FormValue("notification_id"), 10, 64,
	)
	if err != nil || notificationID <= 0 {
		http.Error(w, "Invalid notification ID", http.StatusBadRequest)
		return
	}

	err = database.MarkNotificationAsRead(notificationID, int64(UserID))
	if err != nil {
		log.Printf("MarkNotificationAsReadHandler: %v", err)
		http.Error(w, "Unable to mark notification as read", http.StatusNotFound)
		return
	}
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

func NotificationStreamHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := utils.GetUserID(r)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Straming unsupported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accell-Buffering", "no")

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	var lastCount = -1

	for {
		count, err := database.GetUnreadNotificationCount(int64(userID))
		if err != nil {
			log.Printf("NotificationStreamerHandler: %v", err)
			return
		}

		if count != lastCount {
			data, err := json.Marshal(struct {
				Count int `json:"count"`
			}{
				Count: count,
			})

			if err != nil {
				log.Printf("Encoding notification count: %v", err)
				return
			}

			if _, err := fmt.Fprintf(
				w, "event: unread-count\ndata:%s\n\n",
				data,
			); err != nil {
				return
			}
			flusher.Flush()
			lastCount = count
		}
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
		}
	}
}
