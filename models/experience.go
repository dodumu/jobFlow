package models

import "time"

type Experience struct {
	ID            int        `json:"id"`
	UserID        int        `json:"user_id"`
	Company       string     `json:"company"`
	Position      string     `json:"position"`
	StartDate     time.Time  `json:"start_date"`
	EndDate       *time.Time `json:"end_date"`
	Description   string     `json:"description"`
	CurrentlyWork bool       `json:"currently_working"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at"`
}
