package worker

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"jobFlow/database"
	"jobFlow/models"
)

func setupWorkerTestDB(t *testing.T) int {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(path); err != nil {
		t.Fatalf("initialising test database: %v", err)
	}
	t.Cleanup(func() { database.DB.Close() })

	res, err := database.DB.Exec(`
		INSERT INTO users (first_name, last_name, username, password_hash, email, date_of_birth, role)
		VALUES ('Test', 'User', 'ada', 'fake-hash', 'ada@example.com', '2000-01-01', 'individual')`)
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("reading user id: %v", err)
	}
	return int(id)
}

func queueJob(t *testing.T, userID int, jobType string) int {
	t.Helper()

	id, err := database.CreateAIJob(userID, jobType, `{}`)
	if err != nil {
		t.Fatalf("queueing job: %v", err)
	}
	return id
}

func loadJob(t *testing.T, id, userID int) models.AIJob {
	t.Helper()

	job, err := database.GetAIJobForUser(id, userID)
	if err != nil {
		t.Fatalf("loading job: %v", err)
	}
	return job
}

func TestWorkerCompletesAJob(t *testing.T) {
	userID := setupWorkerTestDB(t)
	id := queueJob(t, userID, "echo")

	w := New()
	w.Register("echo", func(ctx context.Context, job models.AIJob) (Output, error) {
		return Output{Result: `{"ok": true}`, Model: "test-model", PromptVersion: "v1"}, nil
	})

	if !w.runOne(context.Background()) {
		t.Fatal("runOne found no job, want one")
	}

	job := loadJob(t, id, userID)
	if job.Status != models.AIJobStatusDone {
		t.Errorf("status = %q, want %q", job.Status, models.AIJobStatusDone)
	}
	if job.Result.String != `{"ok": true}` {
		t.Errorf("result = %q", job.Result.String)
	}
	if job.Model.String != "test-model" || job.PromptVersion.String != "v1" {
		t.Errorf("model/prompt = %q/%q", job.Model.String, job.PromptVersion.String)
	}
	if job.Attempts != 1 {
		t.Errorf("attempts = %d, want 1", job.Attempts)
	}
	if !job.FinishedAt.Valid {
		t.Error("finished_at should be set")
	}
}

func TestWorkerRecordsHandlerErrors(t *testing.T) {
	userID := setupWorkerTestDB(t)
	id := queueJob(t, userID, "broken")

	w := New()
	w.Register("broken", func(ctx context.Context, job models.AIJob) (Output, error) {
		return Output{}, errors.New("provider unavailable")
	})

	w.runOne(context.Background())

	job := loadJob(t, id, userID)
	if job.Status != models.AIJobStatusFailed {
		t.Errorf("status = %q, want %q", job.Status, models.AIJobStatusFailed)
	}
	if job.Error.String != "provider unavailable" {
		t.Errorf("error = %q", job.Error.String)
	}
}

func TestWorkerSurvivesAPanic(t *testing.T) {
	userID := setupWorkerTestDB(t)
	panicID := queueJob(t, userID, "explodes")
	goodID := queueJob(t, userID, "fine")

	w := New()
	w.Register("explodes", func(ctx context.Context, job models.AIJob) (Output, error) {
		panic("something went badly wrong")
	})
	w.Register("fine", func(ctx context.Context, job models.AIJob) (Output, error) {
		return Output{Result: `{}`}, nil
	})

	w.runOne(context.Background())
	w.runOne(context.Background())

	if got := loadJob(t, panicID, userID).Status; got != models.AIJobStatusFailed {
		t.Errorf("panicking job status = %q, want %q", got, models.AIJobStatusFailed)
	}
	if got := loadJob(t, goodID, userID).Status; got != models.AIJobStatusDone {
		t.Errorf("next job status = %q, want %q", got, models.AIJobStatusDone)
	}
}

func TestWorkerFailsUnknownJobTypes(t *testing.T) {
	userID := setupWorkerTestDB(t)
	id := queueJob(t, userID, "mystery")

	New().runOne(context.Background())

	if got := loadJob(t, id, userID).Status; got != models.AIJobStatusFailed {
		t.Errorf("status = %q, want %q", got, models.AIJobStatusFailed)
	}
}

func TestJobsAreClaimedOnceAndOldestFirst(t *testing.T) {
	userID := setupWorkerTestDB(t)
	first := queueJob(t, userID, "echo")
	second := queueJob(t, userID, "echo")

	a, err := database.ClaimNextAIJob()
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}
	b, err := database.ClaimNextAIJob()
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}

	if a.ID != first || b.ID != second {
		t.Errorf("claimed %d then %d, want %d then %d", a.ID, b.ID, first, second)
	}

	if _, err := database.ClaimNextAIJob(); err == nil {
		t.Error("third claim succeeded, want sql.ErrNoRows because the queue is empty")
	}
}

func TestRunningJobsAreRequeuedAtStartup(t *testing.T) {
	userID := setupWorkerTestDB(t)
	id := queueJob(t, userID, "echo")

	if _, err := database.ClaimNextAIJob(); err != nil { // now "running"
		t.Fatalf("claim: %v", err)
	}

	count, err := database.RequeueRunningAIJobs()
	if err != nil {
		t.Fatalf("requeue: %v", err)
	}
	if count != 1 {
		t.Errorf("requeued %d jobs, want 1", count)
	}
	if got := loadJob(t, id, userID).Status; got != models.AIJobStatusQueued {
		t.Errorf("status = %q, want %q", got, models.AIJobStatusQueued)
	}
}
