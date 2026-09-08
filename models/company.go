package models

import "time"

type Company struct {
	ID                 int       `json:"id"`
	UserID             int       `json:"user_id"`
	CompanyName        string    `json:"company_name"`
	Description        string    `json:"description"`
	Website            string    `json:"website"`
	Location           string    `json:"location"`
	Logo               string    `json:"logo"`
	Industry           string    `json:"industry"`
	CompanySize        string    `json:"company_size"`
	FoundedYear        int       `json:"founded_year"`
	VerificationStatus string    `json:"verification_status"`
	CreatedAt          time.Time `json:"created_at"`
}
