package database

import (
	"fmt"

	"jobFlow/models"
)

// CreateAIJob queues a new AI job for a user and returns its ID.
// The worker will pick it up later; this function never calls the AI itself.
func CreateAIJob(userID int, jobType string, input string) (int, error) {
	query := `INSERT INTO ai_jobs (user_id, type, input) VALUES (?, ?, ?)`

	res, err := DB.Exec(query, userID, jobType, input)
	if err != nil {
		return 0, fmt.Errorf("creating ai job: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("reading new ai job id: %w", err)
	}

	return int(id), nil
}

// GetAIJobForUser loads a job, but only if it belongs to that user.
// For someone else's job it returns sql.ErrNoRows, exactly as if it did not exist.
func GetAIJobForUser(id, userID int) (models.AIJob, error) {
	query := `
		SELECT id, user_id, type, status, input, result, error,
		       model, prompt_version, attempts,
		       created_at, started_at, finished_at
		FROM ai_jobs
		WHERE id = ? AND user_id = ?`

	var job models.AIJob

	err := DB.QueryRow(query, id, userID).Scan(
		&job.ID,
		&job.UserID,
		&job.Type,
		&job.Status,
		&job.Input,
		&job.Result,
		&job.Error,
		&job.Model,
		&job.PromptVersion,
		&job.Attempts,
		&job.CreatedAt,
		&job.StartedAt,
		&job.FinishedAt,
	)
	if err != nil {
		return models.AIJob{}, err
	}

	return job, nil
}