package models

import (
	"database/sql"
	"time"
)

const (
	AIJobTypeCVDraft  = "cv_draft"
	AIJobTypeJobMatch = "job_match"
)

const (
	AIJobStatusQueued  = "queued"
	AIJobStatusRunning = "running"
	AiJobStatusDone    = "done"
	AIJobStatusFailed  = "failed"
)

type AIJob struct {
	ID            int
	UserID        int
	Type          string
	Status        string
	Input         string // JSON text describing what to do
	Result        sql.NullString
	Error         sql.NullString
	Model         sql.NullString
	PromptVersion sql.NullString
	Attempts      int
	CreatedAt     time.Time
	StartedAt     sql.NullTime
	FinishedAt    sql.NullTime
}
