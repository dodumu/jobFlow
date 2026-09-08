package models

import (
	"database/sql"
	"time"
)

type UserProfile struct {
	ID             int          `json:"id"`
	UserID         int          `json:"user_id"`
	ProfilePicture string       `json:"profile_picture"`
	Headline       string       `json:"headline"`
	Bio            string       `json:"bio"`
	Location       string       `json:"location"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      sql.NullTime `json:"updated_at"`
}
