package models

import "database/sql"

type Notification struct {
	ID        int64
	UserID    int64
	ActorID   sql.NullInt64
	Type      string
	Message   string
	Link      sql.NullString
	IsRead    bool
	CreatedAt string
}
