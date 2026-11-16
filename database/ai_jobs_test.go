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


func TestCreateAIJobWithLimitRefusesAfterLimit(t *testing.T) {
	setupAIJobsTestDB(t)
	userID := createAIJobsTestUser(t, "limit_user")

	// Limit of 2 per day: the first two are fine.
	for i := 1; i <= 2; i++ {
		if _, err := database.CreateAIJobWithLimit(userID, models.AIJobTypeCVDraft, "{}", 2); err != nil {
			t.Fatalf("job %d should be allowed, got error: %v", i, err)
		}
	}

	// The third must be refused with ErrDailyAILimit.
	_, err := database.CreateAIJobWithLimit(userID, models.AIJobTypeCVDraft, "{}", 2)
	if !errors.Is(err, database.ErrDailyAILimit) {
		t.Fatalf("expected ErrDailyAILimit, got %v", err)
	}
}

func TestAIJobLimitIsPerUser(t *testing.T) {
	setupAIJobsTestDB(t)
	first := createAIJobsTestUser(t, "limit_first")
	second := createAIJobsTestUser(t, "limit_second")

	// The first user uses up their single allowed job.
	if _, err := database.CreateAIJobWithLimit(first, models.AIJobTypeCVDraft, "{}", 1); err != nil {
		t.Fatalf("first user's job should be allowed: %v", err)
	}

	// The second user must not be affected.
	if _, err := database.CreateAIJobWithLimit(second, models.AIJobTypeCVDraft, "{}", 1); err != nil {
		t.Fatalf("second user should still be allowed, got: %v", err)
	}
}

func TestAIJobLimitIgnoresYesterday(t *testing.T) {
	setupAIJobsTestDB(t)
	userID := createAIJobsTestUser(t, "limit_old")

	// Create a job, then pretend it was made two days ago.
	id, err := database.CreateAIJob(userID, models.AIJobTypeCVDraft, "{}")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.DB.Exec(
		`UPDATE ai_jobs SET created_at = datetime('now', '-2 days') WHERE id = ?`, id,
	); err != nil {
		t.Fatal(err)
	}

	// With a limit of 1, an old job must not count against today.
	if _, err := database.CreateAIJobWithLimit(userID, models.AIJobTypeCVDraft, "{}", 1); err != nil {
		t.Fatalf("yesterday's job should not count, got: %v", err)
	}
}

// finishJobWithTokens marks a job done and stores a result with token counts,
// the way a real handler's result would look.
func finishJobWithTokens(t *testing.T, id, in, out int) {
	t.Helper()
	result := fmt.Sprintf(`{"input_tokens":%d,"output_tokens":%d}`, in, out)
	_, err := database.DB.Exec(
		`UPDATE ai_jobs SET status = 'done', result = ? WHERE id = ?`, result, id,
	)
	if err != nil {
		t.Fatal(err)
	}
}

func TestTokensUsedTodayAddsUpFinishedJobs(t *testing.T) {
	setupAIJobsTestDB(t)
	userID := createAIJobsTestUser(t, "token_user")

	// No jobs yet: zero, not an error.
	used, err := database.TokensUsedToday(userID)
	if err != nil || used != 0 {
		t.Fatalf("expected 0 tokens and no error, got %d, %v", used, err)
	}

	id1, _ := database.CreateAIJob(userID, models.AIJobTypeCVDraft, "{}")
	id2, _ := database.CreateAIJob(userID, models.AIJobTypeCVDraft, "{}")
	finishJobWithTokens(t, id1, 100, 50)
	finishJobWithTokens(t, id2, 10, 5)

	used, err = database.TokensUsedToday(userID)
	if err != nil {
		t.Fatal(err)
	}
	if used != 165 {
		t.Fatalf("expected 165 tokens, got %d", used)
	}
}

func TestCreateAIJobRefusedWhenTokenBudgetSpent(t *testing.T) {
	setupAIJobsTestDB(t)
	userID := createAIJobsTestUser(t, "budget_user")

	id, _ := database.CreateAIJob(userID, models.AIJobTypeCVDraft, "{}")
	finishJobWithTokens(t, id, 600, 500) // 1100 used

	limits := database.AILimits{MaxJobsPerDay: 10, MaxTokensPerDay: 1000}
	_, err := database.CreateAIJobWithLimits(userID, models.AIJobTypeCVDraft, "{}", limits)
	if !errors.Is(err, database.ErrDailyTokenBudget) {
		t.Fatalf("expected ErrDailyTokenBudget, got %v", err)
	}
}

func TestZeroTokenBudgetMeansNoTokenLimit(t *testing.T) {
	setupAIJobsTestDB(t)
	userID := createAIJobsTestUser(t, "nobudget_user")

	id, _ := database.CreateAIJob(userID, models.AIJobTypeCVDraft, "{}")
	finishJobWithTokens(t, id, 999999, 999999)

	limits := database.AILimits{MaxJobsPerDay: 10} // MaxTokensPerDay left at 0
	if _, err := database.CreateAIJobWithLimits(userID, models.AIJobTypeCVDraft, "{}", limits); err != nil {
		t.Fatalf("0 should mean unlimited tokens, got: %v", err)
	}
}