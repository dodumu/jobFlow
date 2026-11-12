package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"jobFlow/constants"
	"jobFlow/database"
	"jobFlow/middleware"
	"jobFlow/models"
)

// setupTestDB gives the test its own empty database in a temporary folder.
func setupTestDB(t *testing.T) {
	t.Helper()

	path := filepath.Join(t.TempDir(), "test.db")
	if err := database.InitDB(path); err != nil {
		t.Fatalf("initialising test database: %v", err)
	}

	t.Cleanup(func() {
		database.DB.Close()
	})
}

// postForm sends a fake form submission straight to a handler.
func postForm(handler http.HandlerFunc, target string, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	rec := httptest.NewRecorder()
	handler(rec, req)

	return rec
}

// validRegistration returns a registration form that should be accepted.
func validRegistration() url.Values {
	return url.Values{
		"username":         {"david"},
		"password":         {"correct7horse"},
		"confirm_password": {"correct7horse"},
		"first_name":       {"David"},
		"last_name":        {"Tester"},
		"email":            {"david@example.com"},
		"date_of_birth":    {"2000-01-01"},
		"role":             {"individual"},
	}
}

// registerTestUser creates the "david" account through the real handler.
func registerTestUser(t *testing.T) {
	t.Helper()

	rec := postForm(RegisterHandler, "/register", validRegistration())
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("registering test user: status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
}

func registerUser(t *testing.T, username, role string) int {
	t.Helper()

	form := validRegistration()
	form.Set("username", username)
	form.Set("email", username+"@example.com")
	form.Set("role", role)

	if role == constants.RoleCompany {
		form.Set("company_name", username+" Ltd")
		form.Set("description", "A test company")
		form.Set("website", "https://example.com")
		form.Set("location", "Lagos")
		form.Set("industry", "Technology")
		form.Set("company_size", "11-50")
	}

	rec := postForm(RegisterHandler, "/register", form)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("registering %s: status = %d, want %d", username, rec.Code, http.StatusSeeOther)
	}

	user, err := database.GetUserByUsername(username)
	if err != nil {
		t.Fatalf("loading %s: %v", username, err)
	}

	return user.ID
}

// createJob inserts a job owned by the company account companyUserID.
func createJob(t *testing.T, companyUserID int, status string, deadline time.Time) int {
	t.Helper()

	company, err := database.GetCompanyByUserID(companyUserID)
	if err != nil {
		t.Fatalf("loading company: %v", err)
	}

	jobID, err := database.CreateJob(models.Job{
		CompanyID:      company.ID,
		Title:          "Backend Developer",
		Description:    "Build APIs in Go.",
		Requirements:   "Go, SQL",
		Location:       "Remote",
		EmploymentType: "full-time",
		SalaryMin:      1000,
		SalaryMax:      2000,
		Deadline:       deadline,
		Status:         status,
	})
	if err != nil {
		t.Fatalf("creating job: %v", err)
	}

	return jobID
}

// requestAs calls a handler as a logged-in user. A userID of 0 means "not logged in".
// Pass a nil form for requests without a body.
func requestAs(handler http.HandlerFunc, method, target string, userID int, form url.Values) *httptest.ResponseRecorder {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	req := httptest.NewRequest(method, target, body)
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if userID > 0 {
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
	}

	rec := httptest.NewRecorder()
	handler(rec, req)

	return rec
}

func createApplication(t *testing.T, jobID, applicantID int) int {
	t.Helper()

	applicationID, err := database.CreateApplication(models.Application{
		JobID:       jobID,
		UserID:      applicantID,
		CoverLetter: "I would like this job.",
		Status:      constants.ApplicationStatusPending,
	})
	if err != nil {
		t.Fatalf("createing application: %v", err)
	}

	return applicationID
}

func requestWithID(handler http.HandlerFunc, method, target string, userID, id int) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, target, nil)
	req.SetPathValue("id", strconv.Itoa(id))

	if userID > 0 {
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, userID)
		req = req.WithContext(ctx)
	}

	rec := httptest.NewRecorder()
	handler(rec, req)
	return rec
}

func applicationStatus(t *testing.T, applicationID int) string {
	t.Helper()

	application, err := database.GetApplicationByID(applicationID)
	if err != nil {
		t.Fatalf("loading application: %v", err)
	}

	return application.Status
}
