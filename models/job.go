package models

import (
	"database/sql"
	"time"
)

type Job struct {
	ID             int          `json:"id"`
	CompanyID      int          `json:"company_id"`
	Title          string       `json:"title"`
	Description    string       `json:"description"`
	Requirements   string       `json:"requirements"`
	Location       string       `json:"location"`
	EmploymentType string       `json:"employment_type"`
	SalaryMin      int          `json:"salary_min"`
	SalaryMax      int          `json:"salary_max"`
	Deadline       time.Time    `json:"deadline"`
	Status         string       `json:"status"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      sql.NullTime `json:"updated_at"`
}
