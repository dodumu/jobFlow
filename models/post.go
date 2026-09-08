package models

import "time"

type Post struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	Content   string     `json:"content"`
	Type      string     `json:"type"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
