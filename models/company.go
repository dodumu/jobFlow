package models

import "time"

type Company struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	CompanyName string    `json:"company_name"`
	Description string    `json:"description"`
	Website     string    `json:"website"`
	Location    string    `json:"location"`
	Logo        string    `json:"logo"`
	CreatedAt   time.Time `json:"created_at"`
}
