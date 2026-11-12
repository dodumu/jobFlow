package database

import (
	"fmt"

	"jobFlow/models"
)

const aiJobColumns = `id, user_id, type, status, input, result, error,
	model, prompt_version, attempts, created_at, started_at, finished_at`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanAIJob(row rowScanner) (models.AIJob, error) {
	var job models.AIJob

	err := row.Scan(
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

func GetAIJobForUser(id, userID int) (models.AIJob, error) {
	query := `SELECT ` + aiJobColumns + ` FROM ai_jobs WHERE id  = ? AND user_id = ?`

	return scanAIJob(DB.QueryRow(query, id, userID))
}

func ClaimNextAIJob() (models.AIJob, error) {
	query := `
		UPDATE ai_jobs
		SET status = ?,
		    started_at = CURRENT_TIMESTAMP,
		    attempts = attempts + 1
		WHERE id = (
			SELECT id FROM ai_jobs
			WHERE status = ?
			ORDER BY created_at, id
			LIMIT 1
		)
		RETURNING ` + aiJobColumns

	return scanAIJob(DB.QueryRow(query, models.AIJobStatusRunning, models.AIJobStatusQueued))

}
func CompleteAIJob(id int, result, model, promptVersion string) error {
	query := `
		UPDATE ai_jobs
		SET status = ?,
		    result = ?,
		    model = NULLIF(?, ''),
		    prompt_version = NULLIF(?, ''),
		    error = NULL,
		    finished_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := DB.Exec(query, models.AIJobStatusDone, result, model, promptVersion, id)
	if err != nil {
		return fmt.Errorf("completing ai job %d: %w", id, err)
	}
	return nil
}

func FailAIJob(id int, message string) error {
	query := `
		UPDATE ai_jobs
		SET status = ?,
		    error = ?,
		    finished_at = CURRENT_TIMESTAMP
		WHERE id = ?`

	_, err := DB.Exec(query, models.AIJobStatusFailed, message, id)
	if err != nil {
		return fmt.Errorf("failing ai job %d: %w", id, err)
	}
	return nil
}

func RequeueRunningAIJobs() (int64, error) {
	query := `UPDATE ai_jobs SET status = ?, started_at = NULL WHERE status = ?`

	res, err := DB.Exec(query, models.AIJobStatusQueued, models.AIJobStatusRunning)
	if err != nil {
		return 0, fmt.Errorf("requeueing running ai jobs: %w", err)
	}
	return res.RowsAffected()
}
