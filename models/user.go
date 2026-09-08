package models

import "time"

type User struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Email        string    `json:"email"`
	DOB          string    `json:"date_of_birth"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
