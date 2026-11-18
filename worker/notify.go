package worker

import (
	"database/sql"
	"fmt"
	"jobFlow/database"
	"jobFlow/models"
	"log"
)

func notifyFinished(job models.AIJob, succeeded bool) {
	notificationType := "ai_job_done"
	message := "Your AI request is ready."

	if !succeeded {
		notificationType = "ai_job_failed"
		message = "Your AI request could not be completed. You can try again."
	}

	link := fmt.Sprintf("/ai/jobs/%d", job.ID)

	_, err := database.CreateNotification(int64(job.UserID), sql.NullInt64{}, notificationType, message, link)
	if err != nil {
		log.Printf("ai worker: could not notify user %d about job %d: %v", job.UserID, job.ID, err)
	}
}
