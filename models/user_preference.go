package models

import "time"

type UserPreference struct {
	ID        int        `json:"id"`
	UserID    int        `json:"user_id"`
	SalaryMin *int       `json:"salary_min"`
	SalaryMax *int       `json:"salary_max"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at"`
}
