package worker

import (
	"jobFlow/database"
	"jobFlow/models"
	"testing"
)

func TestNotifyFinishedCreatesNotification(t *testing.T) {
	setupWorkerTestDB(t)           // use your existing helper's name
	userID := setupWorkerTestDB(t) // and your user helper

	notifyFinished(models.AIJob{ID: 7, UserID: userID}, true)
	notifyFinished(models.AIJob{ID: 8, UserID: userID}, false)

	notes, err := database.GetUserNotifications(int64(userID))
	if err != nil {
		t.Fatal(err)
	}
	if len(notes) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(notes))
	}

	// Newest first, so the failed one (job 8) comes first.
	if notes[0].Type != "ai_job_failed" || notes[1].Type != "ai_job_done" {
		t.Fatalf("unexpected types: %q, %q", notes[0].Type, notes[1].Type)
	}
	if notes[1].Link.String != "/ai/jobs/7" {
		t.Fatalf("unexpected link: %q", notes[1].Link.String)
	}
}
