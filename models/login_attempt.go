package models

import (
	"database/sql"
	"time"
)

type LoginAttempt struct {
	ID             int
	Username       string
	IPAddress      string
	FailedAttempts int
	LastAttemptAt  time.Time
	BlockedUntil   sql.NullTime
}
