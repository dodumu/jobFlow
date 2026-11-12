package database_test

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	"jobFlow/database"
	"jobFlow/models"
)

func setupAIJobsTestDB(t *testing.T) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(path); err != nil {
		t.Fatalf("initialising test database: %v", err)
	}
	t.Cleanup(func() { database.DB.Close() })
}

func createAIJobsTestUser(t *testing.T, username string) int {
	t.Helper()

	res, err := database.DB.Exec(`
		INSERT INTO users (first_name, last_name, username, password_hash, email, date_of_birth, role)
		VALUES ('Test', 'User', ?, 'fake-hash', ?, '2000-01-01', 'individual')`,
		username, username+"@example.com")
	if err != nil {
		t.Fatalf("creating user: %v", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("reading user id: %v", err)
	}
	return int(id)
}

func TestCreateAIJobStartsQueued(t *testing.T) {
	setupAIJobsTestDB(t)
	userID := createAIJobsTestUser(t, "ada")

	id, err := database.CreateAIJob(userID, models.AIJobTypeCVDraft, `{"job_id": 7}`)
	if err != nil {
		t.Fatalf("CreateAIJob: %v", err)
	}

	job, err := database.GetAIJobForUser(id, userID)
	if err != nil {
		t.Fatalf("GetAIJobForUser: %v", err)
	}

	if job.Status != models.AIJobStatusQueued {
		t.Errorf("status = %q, want %q", job.Status, models.AIJobStatusQueued)
	}
	if job.Attempts != 0 {
		t.Errorf("attempts = %d, want 0", job.Attempts)
	}
	if job.Result.Valid {
		t.Error("a new job must not have a result yet")
	}
	if job.Input != `{"job_id": 7}` {
		t.Errorf("input = %q, want it stored unchanged", job.Input)
	}
}

func TestAIJobIsPrivateToItsOwner(t *testing.T) {
	setupAIJobsTestDB(t)
	owner := createAIJobsTestUser(t, "ada")
	stranger := createAIJobsTestUser(t, "grace")

	id, err := database.CreateAIJob(owner, models.AIJobTypeCVDraft, `{}`)
	if err != nil {
		t.Fatalf("CreateAIJob: %v", err)
	}

	_, err = database.GetAIJobForUser(id, stranger)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("another user's lookup returned %v, want sql.ErrNoRows", err)
	}
}