package models

import "time"

type User struct {
	ID           int       `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	Email        string    `json:"email"`
	DOB          string    `json:"date_of_birth"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}
