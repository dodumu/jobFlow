package models

import "time"

type Post struct {
	ID           int        `json:"id"`
	UserID       *int       `json:"user_id"`
	CompanyID    *int       `json:"company_id"`
	Content      string     `json:"content"`
	Type         string     `json:"type"`
	SharedPostID *int       `json:"shared_post_id"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at"`
}
