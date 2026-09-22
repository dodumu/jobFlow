package models

import (
	"database/sql"
	"time"
)

type Application struct {
	ID          int          `json:"id"`
	JobID       int          `json:"job_id"`
	UserID      int          `json:"user_id"`
	CoverLetter string       `json:"cover_letter"`
	Resume      string       `json:"resume"`
	Status      string       `json:"status"`
	AppliedAt   time.Time    `json:"applied_at"`
	UpdatedAt   sql.NullTime `json:"updated_at"`
}

type UserApplication struct {
	ApplicationID  int
	JobID          int
	JobTitle       string
	Location       string
	EmploymentType string
	Status         string
	AppliedAt      time.Time
}
type CompanyApplication struct {
	ApplicationID int
	JobID         int
	JobTitle      string
	UserID        int
	ApplicantName string
	Status        string
	AppliedAt     time.Time
}
